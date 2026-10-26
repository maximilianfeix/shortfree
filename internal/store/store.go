// Package store persists scan state: a JSON record per non-delegated 1-3
// character name, and a free-bitmap per TLD for 4 character names.
package store

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Entry is what we know about one name that is not delegated in DNS. Keys are
// short and times are Unix seconds because the state holds ~200k entries.
type Entry struct {
	DNS     string `json:"d"`           // "free" or "unknown" (no DNS answer at the last sweep)
	RDAP    string `json:"r,omitempty"` // rdap.Status, empty until verified
	Checked int64  `json:"c,omitempty"` // last RDAP check
	Since   int64  `json:"s,omitempty"` // first seen non-delegated
	Expires int64  `json:"e,omitempty"` // RDAP expiration event (not a drop date)
	Phase   string `json:"l,omitempty"` // dropping only: "redemption" or "pending_delete"
	// Freed is the time of an observed transition from RDAP registered or
	// dropping to RDAP 404. Only then is a name reported as "freed".
	Freed int64 `json:"f,omitempty"`
}

// Scan records the last complete DNS sweep of one TLD's 1-3 character names.
type Scan struct {
	At int64 `json:"at"` // when the sweep finished (Unix seconds)
	// Held is set when the latest scan output for this TLD was rejected
	// (e.g. implausible jump in the count of free names) and this older
	// data was kept instead.
	Held string `json:"held,omitempty"`
}

// ScannedAt returns At as a time.
func (s *Scan) ScannedAt() time.Time { return time.Unix(s.At, 0).UTC() }

// CheckedAt returns Checked as a time.
func (e *Entry) CheckedAt() time.Time { return time.Unix(e.Checked, 0).UTC() }

// ExpiresAt returns Expires as a time, zero if unknown.
func (e *Entry) ExpiresAt() time.Time {
	if e.Expires == 0 {
		return time.Time{}
	}
	return time.Unix(e.Expires, 0).UTC()
}

// State holds 1-3 character results, keyed by TLD then label.
type State struct {
	Updated time.Time                    `json:"updated"`
	TLDs    map[string]map[string]*Entry `json:"tlds"`
	// Scans holds the last complete sweep per TLD. A TLD with entries but
	// no Scans record comes from a state written before scans were
	// recorded; it counts as scanned at Updated.
	Scans map[string]*Scan `json:"scans,omitempty"`
}

// ScanOf returns the sweep record of tld, or nil if tld was never scanned.
func (s *State) ScanOf(tld string) *Scan {
	if sc := s.Scans[tld]; sc != nil {
		return sc
	}
	if _, ok := s.TLDs[tld]; ok {
		return &Scan{At: s.Updated.Unix()} // legacy state
	}
	return nil
}

// SetScan records a complete sweep of tld at t.
func (s *State) SetScan(tld string, t time.Time) { s.PutScan(tld, &Scan{At: t.Unix()}) }

// PutScan stores the scan record of tld.
func (s *State) PutScan(tld string, sc *Scan) {
	if s.Scans == nil {
		s.Scans = map[string]*Scan{}
	}
	s.Scans[tld] = sc
}

// Load reads path (gzip-compressed if it ends in .gz), returning an empty
// state if it does not exist.
func Load(path string) (*State, error) {
	s := &State{TLDs: map[string]map[string]*Entry{}}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if strings.HasSuffix(path, ".gz") {
		zr, err := gzip.NewReader(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
		if b, err = io.ReadAll(zr); err != nil {
			return nil, err
		}
	}
	if err := json.Unmarshal(b, s); err != nil {
		return nil, err
	}
	if s.TLDs == nil {
		s.TLDs = map[string]map[string]*Entry{}
	}
	return s, nil
}

// Save writes the state atomically, gzip-compressed if path ends in .gz.
func (s *State) Save(path string) error {
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	if strings.HasSuffix(path, ".gz") {
		var buf bytes.Buffer
		zw, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
		zw.Write(b)
		if err := zw.Close(); err != nil {
			return err
		}
		b = buf.Bytes()
	}
	return WriteAtomic(path, b)
}

// TLD returns the entry map for tld, creating it.
func (s *State) TLD(tld string) map[string]*Entry {
	m, ok := s.TLDs[tld]
	if !ok {
		m = map[string]*Entry{}
		s.TLDs[tld] = m
	}
	return m
}

// Labels returns the labels of tld sorted by length, then alphabetically.
func (s *State) Labels(tld string) []string {
	m := s.TLDs[tld]
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	SortLabels(out)
	return out
}

// SortLabels sorts by length, then alphabetically.
func SortLabels(l []string) {
	sort.Slice(l, func(i, j int) bool {
		if len(l[i]) != len(l[j]) {
			return len(l[i]) < len(l[j])
		}
		return l[i] < l[j]
	})
}

// Bitmap is a fixed-size bitset indexed by names.Index.
type Bitmap []byte

// NewBitmap returns a zeroed bitmap with room for n bits.
func NewBitmap(n int) Bitmap { return make(Bitmap, (n+7)/8) }

func (b Bitmap) Get(i int) bool { return b[i>>3]&(1<<(i&7)) != 0 }

func (b Bitmap) Set(i int, v bool) {
	if v {
		b[i>>3] |= 1 << (i & 7)
	} else {
		b[i>>3] &^= 1 << (i & 7)
	}
}

// Count returns the number of set bits.
func (b Bitmap) Count() int {
	n := 0
	for _, x := range b {
		for ; x != 0; x &= x - 1 {
			n++
		}
	}
	return n
}

// ReadBitmap reads a bitmap of n bits. A missing file returns an error
// wrapping os.ErrNotExist; a file of the wrong size is an error, never an
// empty bitmap (that would turn every name into "taken").
func ReadBitmap(path string, n int) (Bitmap, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if want := (n + 7) / 8; len(b) != want {
		return nil, fmt.Errorf("%s: %d bytes, want %d", path, len(b), want)
	}
	return Bitmap(b), nil
}

// LoadBitmap reads a bitmap of n bits, or returns an empty one if the file
// does not exist yet. A file of the wrong size is an error.
func LoadBitmap(path string, n int) (Bitmap, error) {
	b, err := ReadBitmap(path, n)
	if errors.Is(err, os.ErrNotExist) {
		return NewBitmap(n), nil
	}
	return b, err
}

// ShardScan is one DNS sweep of one 4 character shard.
type ShardScan struct {
	At      int64 `json:"at"`      // when the sweep finished (Unix seconds)
	Names   int   `json:"names"`   // names in the shard
	Unknown int   `json:"unknown"` // names without a DNS answer (bit kept from before)
}

// Coverage records which shards of a 4 character bitmap have been swept. It
// is stored next to the bitmap as <tld>.cov.json. A zero bit in a shard that
// was never swept means "not checked", not "taken".
type Coverage struct {
	Shards int                   `json:"shards"`
	Scans  map[string]*ShardScan `json:"scans"` // key: shard number
	// Held is set when the latest bitmap was rejected and this older one
	// kept (see Scan.Held).
	Held string `json:"held,omitempty"`
}

// CoveragePath returns the coverage file belonging to a bitmap path.
func CoveragePath(bitmapPath string) string {
	return strings.TrimSuffix(bitmapPath, ".bin") + ".cov.json"
}

// LoadCoverage reads a coverage file. A missing file yields an empty
// coverage with the given shard count (nothing known to be swept).
func LoadCoverage(path string, shards int) (*Coverage, error) {
	c := &Coverage{Shards: shards, Scans: map[string]*ShardScan{}}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return c, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, c); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if c.Scans == nil {
		c.Scans = map[string]*ShardScan{}
	}
	if c.Shards < 1 {
		return nil, fmt.Errorf("%s: shards = %d", path, c.Shards)
	}
	return c, nil
}

// Save writes c atomically.
func (c *Coverage) Save(path string) error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return WriteAtomic(path, append(b, '\n'))
}

// Scanned returns the sweep of shard, or nil if it was never swept.
func (c *Coverage) Scanned(shard int) *ShardScan {
	if c == nil {
		return nil
	}
	return c.Scans[fmt.Sprint(shard)]
}

// Record notes a sweep of shard. A different shard count than before
// invalidates every earlier record, because shard numbers no longer match.
func (c *Coverage) Record(shard, shards int, s *ShardScan) {
	if c.Shards != shards {
		c.Shards, c.Scans = shards, map[string]*ShardScan{}
	}
	c.Scans[fmt.Sprint(shard)] = s
	c.Held = ""
}

// Complete reports whether every shard has been swept at least once.
func (c *Coverage) Complete() bool {
	if c == nil || c.Shards < 1 {
		return false
	}
	for i := 0; i < c.Shards; i++ {
		if c.Scanned(i) == nil {
			return false
		}
	}
	return true
}

// WriteAtomic writes data to path via a temp file and rename.
func WriteAtomic(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
