package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/maximilianfeix/shortfree/internal/names"
	"github.com/maximilianfeix/shortfree/internal/store"
)

// defaultDataURL serves the files written by "shortfree export" (GitHub Pages).
const defaultDataURL = "https://maximilianfeix.github.io/shortfree/data/"

// Status of a name in the published snapshot.
const (
	stVerified   = "rdap_verified" // 1-3 chars, RDAP said not registered
	stUnverified = "unverified"    // 1-3 chars, not in DNS, RDAP check pending
	stDropping   = "dropping"      // registered, in redemption / pending delete
	stDNSOnly    = "dns_only"      // 4 chars, not in DNS (never RDAP-checked)
)

func statusRank(s string) int {
	switch s {
	case stVerified:
		return 0
	case stUnverified:
		return 1
	case stDNSOnly:
		return 2
	}
	return 3
}

// snapshot is available.json plus a lookup index built once per download.
type snapshot struct {
	*Available
	status  map[string]map[string]string // tld -> label -> status
	expires map[string]map[string]time.Time
}

func newSnapshot(a *Available) *snapshot {
	s := &snapshot{Available: a, status: map[string]map[string]string{}, expires: map[string]map[string]time.Time{}}
	for tld, l := range a.TLDs {
		m := map[string]string{}
		ex := map[string]time.Time{}
		for _, n := range l.Unverified {
			m[n] = stUnverified
		}
		for _, d := range l.Dropping {
			m[d.Name] = stDropping
			ex[d.Name] = d.Expires
		}
		for _, n := range l.Available {
			m[n] = stVerified
		}
		s.status[tld], s.expires[tld] = m, ex
	}
	return s
}

// tldOrder returns the TLDs in the snapshot, DefaultTLDs first.
func (s *snapshot) tldOrder() []string {
	var out []string
	for _, t := range DefaultTLDs {
		if s.TLDs[t] != nil {
			out = append(out, t)
		}
	}
	var extra []string
	for t := range s.TLDs {
		if !contains(DefaultTLDs, t) {
			extra = append(extra, t)
		}
	}
	sort.Strings(extra)
	return append(out, extra...)
}

// pickTLDs validates a user-supplied TLD list against the snapshot.
func (s *snapshot) pickTLDs(req []string) ([]string, error) {
	all := s.tldOrder()
	if len(req) == 0 {
		return all, nil
	}
	var out []string
	for _, t := range req {
		t = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(t)), ".")
		if t == "" || contains(out, t) {
			continue
		}
		if s.TLDs[t] == nil {
			return nil, fmt.Errorf("unsupported TLD %q: the published data covers %s (use check_domains for any other TLD)", t, "."+strings.Join(all, " ."))
		}
		out = append(out, t)
	}
	if len(out) == 0 {
		return all, nil
	}
	return out, nil
}

// cached is one downloaded file. ok is false until the first success.
type cached[T any] struct {
	v  T
	at time.Time
	ok bool
}

// dataSource downloads the published data files and keeps them in memory
// for ttl. If a refresh fails, the stale copy is served.
type dataSource struct {
	base   string
	client *http.Client
	ttl    time.Duration
	now    func() time.Time

	mu      sync.Mutex
	avail   cached[*snapshot]
	stats   cached[*Stats]
	bitmaps map[string]*cached[store.Bitmap]
}

func newDataSource() *dataSource {
	base := os.Getenv("SHORTFREE_DATA_URL")
	if base == "" {
		base = defaultDataURL
	}
	if !strings.HasSuffix(base, "/") {
		base += "/"
	}
	return &dataSource{
		base:    base,
		client:  &http.Client{Timeout: 30 * time.Second},
		ttl:     time.Hour,
		now:     time.Now,
		bitmaps: map[string]*cached[store.Bitmap]{},
	}
}

var errNotFound = errors.New("not published")

func (d *dataSource) fetch(ctx context.Context, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, d.base+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("%s: %w", d.base+path, errNotFound)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: HTTP %d", d.base+path, resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 64<<20))
}

// load returns c's value, downloading path when the cache is empty or older
// than ttl. Callers hold d.mu.
func load[T any](ctx context.Context, d *dataSource, c *cached[T], path string, decode func([]byte) (T, error)) (T, error) {
	if c.ok && d.now().Sub(c.at) < d.ttl {
		return c.v, nil
	}
	b, err := d.fetch(ctx, path)
	var v T
	if err == nil {
		v, err = decode(b)
	}
	if err != nil {
		if c.ok {
			return c.v, nil // stale beats nothing
		}
		return v, err
	}
	*c = cached[T]{v: v, at: d.now(), ok: true}
	return v, nil
}

func (d *dataSource) available(ctx context.Context) (*snapshot, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return load(ctx, d, &d.avail, "available.json", func(b []byte) (*snapshot, error) {
		var a Available
		if err := json.Unmarshal(b, &a); err != nil {
			return nil, fmt.Errorf("available.json: %w", err)
		}
		if a.TLDs == nil {
			a.TLDs = map[string]*TLDList{}
		}
		return newSnapshot(&a), nil
	})
}

func (d *dataSource) statsFile(ctx context.Context) (*Stats, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return load(ctx, d, &d.stats, "stats.json", func(b []byte) (*Stats, error) {
		var s Stats
		if err := json.Unmarshal(b, &s); err != nil {
			return nil, fmt.Errorf("stats.json: %w", err)
		}
		return &s, nil
	})
}

// bitmap returns the 4 character free-bitmap of tld, or nil if it is not
// published.
func (d *dataSource) bitmap(ctx context.Context, tld string) (store.Bitmap, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	c := d.bitmaps[tld]
	if c == nil {
		c = &cached[store.Bitmap]{}
		d.bitmaps[tld] = c
	}
	want := (names.Space(4) + 7) / 8
	bm, err := load(ctx, d, c, "4/"+tld+".bin", func(b []byte) (store.Bitmap, error) {
		if len(b) != want {
			return nil, fmt.Errorf("4/%s.bin: %d bytes, want %d", tld, len(b), want)
		}
		return store.Bitmap(b), nil
	})
	if errors.Is(err, errNotFound) {
		return nil, nil
	}
	return bm, err
}
