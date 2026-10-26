// Package rdap confirms registration status against each registry's RDAP
// service. A 404 from the authoritative registry means the name is not
// registered; a 200 carries the EPP status, which tells us when a name is
// about to drop.
package rdap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Endpoints maps each supported TLD to its registry RDAP base URL. Sources:
// IANA bootstrap (data.iana.org/rdap/dns.json); TLDs missing from the
// bootstrap file point to their registry's service directly. docs/tlds.md
// lists the source and verification date of each entry.
var Endpoints = map[string]string{
	"com":   "https://rdap.verisign.com/com/v1/",
	"net":   "https://rdap.verisign.com/net/v1/",
	"de":    "https://rdap.denic.de/",
	"ai":    "https://rdap.identitydigital.services/rdap/",
	"io":    "https://rdap.identitydigital.services/rdap/",
	"tools": "https://rdap.identitydigital.services/rdap/",
	"run":   "https://rdap.identitydigital.services/rdap/",
	"app":   "https://pubapi.registry.google/rdap/",
	"dev":   "https://pubapi.registry.google/rdap/",
	"org":   "https://rdap.publicinterestregistry.org/rdap/",
	"co":    "https://rdap.registry.co/co/",
	"xyz":   "https://rdap.centralnic.com/xyz/",
	"fm":    "https://rdap.centralnic.com/fm/",
	"inc":   "https://rdap.centralnic.com/inc/",
	"cc":    "https://tld-rdap.verisign.com/cc/v1/",
	"tv":    "https://rdap.nic.tv/",
	"to":    "https://rdap.tonicregistry.to/rdap/",
	"is":    "https://rdap.isnic.is/rdap/",
	"ly":    "https://rdap.nic.ly/",
	"so":    "https://rdap.nic.so/",
}

// DefaultRate is the TOTAL requests per second shortfree sends to each RDAP
// host name, however many runners scan in parallel: TLDs on one host
// (Identity Digital, Google, CentralNic, Verisign) split it, see SharedRate.
// Measured 2026-10-26: Identity Digital served 10/s without errors but later
// that day banned the client IP for ~23 h (Cloudflare 1015, Retry-After
// 83359), so it gets 2/s in total; Google answered 429 at 10/s; hosts added
// later start at 1-2/s.
var DefaultRate = map[string]float64{
	"rdap.verisign.com":               4,
	"rdap.denic.de":                   2,
	"rdap.identitydigital.services":   2,
	"pubapi.registry.google":          1.5,
	"rdap.publicinterestregistry.org": 2,
	"rdap.registry.co":                2,
	"rdap.centralnic.com":             2,
	"tld-rdap.verisign.com":           2,
	"rdap.nic.tv":                     1,
	"rdap.tonicregistry.to":           1,
	"rdap.isnic.is":                   1,
	"rdap.nic.ly":                     1,
	"rdap.nic.so":                     1,
}

// Host returns the host name of tld's RDAP service, or "" if it has none.
// Rate limits and time budgets are per host.
func Host(tld string) string {
	base, ok := Endpoints[tld]
	if !ok {
		return ""
	}
	u, err := url.Parse(base)
	if err != nil {
		return base
	}
	return u.Hostname()
}

// SharedRate returns the requests per second a process scanning the TLDs in
// scanned may send to host, out of the host's total in DefaultRate: the total
// is split evenly over every TLD on that host in all (the full TLD list) and
// scanned, and this process gets the shares of the TLDs it scans. A matrix
// job scanning only .ai gets a quarter of Identity Digital's total when .ai,
// .io, .tools and .run share it; a local run of all TLDs gets the whole.
func SharedRate(host string, scanned, all []string) float64 {
	total, ok := DefaultRate[host]
	if !ok {
		total = 1
	}
	onHost := map[string]bool{}
	mine := map[string]bool{}
	for _, t := range all {
		if Host(t) == host {
			onHost[t] = true
		}
	}
	for _, t := range scanned {
		if Host(t) == host {
			onHost[t] = true
			mine[t] = true
		}
	}
	if len(mine) == 0 {
		return 0
	}
	return total * float64(len(mine)) / float64(len(onHost))
}

// Status is the registry's view of a name.
type Status string

const (
	Available  Status = "available"  // 404: not registered
	Dropping   Status = "dropping"   // registered but in redemption / pending delete
	Registered Status = "registered" // registered (possibly without DNS)
	Error      Status = "error"      // no definitive answer
)

// Lifecycle phases of a Dropping name, from its EPP status.
const (
	PhaseRedemption    = "redemption"     // redemption period: the old owner can still restore it
	PhasePendingDelete = "pending_delete" // pending delete: restore no longer possible
)

// Info is the result of one lookup.
type Info struct {
	Status Status   `json:"status"`
	EPP    []string `json:"epp,omitempty"`
	// Expires is the RDAP "expiration" event. It is not a drop date: names
	// are usually renewed or deleted on registry-specific schedules.
	Expires time.Time `json:"expires,omitempty"`
	Phase   string    `json:"phase,omitempty"` // Dropping only, see PhaseRedemption
}

// ErrBadResponse means the registry answered, but not with something that
// looks like RDAP (an HTML error page, malformed JSON, another object).
// Such answers must never count as "available" or "registered".
var ErrBadResponse = errors.New("not an RDAP domain response")

// Client looks up domains with one rate limiter per RDAP host.
type Client struct {
	HTTP      *http.Client
	UserAgent string
	// Attempts is the number of tries per lookup. Batch scans can wait out a
	// registry's 429s; interactive callers should give up early.
	Attempts int
	// MaxWait is the longest Retry-After the client waits out. A longer one
	// marks the host as blocked until then; lookups fail fast with
	// ErrBlocked meanwhile.
	MaxWait time.Duration
	rps     float64

	scanned, all []string // set by ShareWith

	mu      sync.Mutex
	limits  map[string]*rate.Limiter
	blocked map[string]time.Time
}

// ErrBlocked means the RDAP host asked us to stay away for longer than
// Client.MaxWait.
var ErrBlocked = errors.New("RDAP host blocked us")

// New returns a client that sends at most rps requests per second per host;
// rps <= 0 uses DefaultRate.
func New(rps float64, userAgent string) *Client {
	return &Client{
		HTTP:      &http.Client{Timeout: 20 * time.Second},
		UserAgent: userAgent,
		Attempts:  6,
		MaxWait:   10 * time.Minute,
		rps:       rps,
		limits:    map[string]*rate.Limiter{},
		blocked:   map[string]time.Time{},
	}
}

func (c *Client) block(host string, until time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if until.After(c.blocked[host]) {
		c.blocked[host] = until
	}
}

func (c *Client) blockedUntil(host string) time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.blocked[host]
}

func (c *Client) limiter(host string) *rate.Limiter {
	c.mu.Lock()
	defer c.mu.Unlock()
	l, ok := c.limits[host]
	if !ok {
		l = rate.NewLimiter(rate.Limit(c.Rate(host)), 1)
		c.limits[host] = l
	}
	return l
}

// ShareWith makes the client use only the share of each host's total rate
// that belongs to the scanned TLDs (see SharedRate), for scans that run in
// parallel with others covering the rest of all. Call it before the first
// lookup.
func (c *Client) ShareWith(scanned, all []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.scanned, c.all = scanned, all
}

// Rate returns the requests per second used for host (see Host).
func (c *Client) Rate(host string) float64 {
	if c.rps > 0 {
		return c.rps
	}
	if c.scanned != nil {
		if r := SharedRate(host, c.scanned, c.all); r > 0 {
			return r
		}
	}
	if r, ok := DefaultRate[host]; ok {
		return r
	}
	return 1
}

type response struct {
	ObjectClass string   `json:"objectClassName"`
	LDHName     string   `json:"ldhName"`
	Status      []string `json:"status"`
	Events      []struct {
		Action string `json:"eventAction"`
		Date   string `json:"eventDate"`
	} `json:"events"`
}

// Lookup queries the registry for label.tld.
func (c *Client) Lookup(ctx context.Context, label, tld string) (Info, error) {
	base, ok := Endpoints[tld]
	if !ok {
		return Info{Status: Error}, fmt.Errorf("no RDAP endpoint for .%s", tld)
	}
	target := base + "domain/" + label + "." + tld
	host := Host(tld)
	if until := c.blockedUntil(host); time.Now().Before(until) {
		return Info{Status: Error}, fmt.Errorf("%s until %s: %w", host, until.UTC().Format(time.RFC3339), ErrBlocked)
	}
	lim := c.limiter(host)

	var lastErr error
	for attempt := 0; attempt < c.Attempts; attempt++ {
		if err := lim.Wait(ctx); err != nil {
			return Info{Status: Error}, err
		}
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
		req.Header.Set("Accept", "application/rdap+json")
		req.Header.Set("User-Agent", c.UserAgent)
		resp, err := c.HTTP.Do(req)
		if err != nil {
			lastErr = err
			sleep(ctx, time.Duration(1<<attempt)*time.Second)
			continue
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		resp.Body.Close()

		html := strings.Contains(strings.ToLower(resp.Header.Get("Content-Type")), "html")
		switch {
		case (resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusOK) && html:
			// A web server's error or landing page (moved endpoint, CDN
			// interstitial), not the registry's RDAP service.
			return Info{Status: Error}, fmt.Errorf("%s: HTTP %d with an HTML body: %w", target, resp.StatusCode, ErrBadResponse)
		case resp.StatusCode == http.StatusNotFound:
			return Info{Status: Available}, nil
		case resp.StatusCode == http.StatusOK:
			info, err := parse(body, label+"."+tld)
			if err != nil {
				return Info{Status: Error}, fmt.Errorf("%s: %w", target, err)
			}
			return info, nil
		case resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500:
			lastErr = fmt.Errorf("%s: HTTP %d", target, resp.StatusCode)
			wait := time.Duration(2<<attempt) * time.Second
			if s, err := strconv.Atoi(resp.Header.Get("Retry-After")); err == nil && s > 0 {
				wait = time.Duration(s) * time.Second
			}
			if wait > c.MaxWait {
				// A ban, not a hiccup: stop asking this host until it ends
				// instead of hammering it with every queued lookup.
				c.block(host, time.Now().Add(wait))
				return Info{Status: Error}, fmt.Errorf("%s: HTTP %d, retry after %s: %w", target, resp.StatusCode, wait, ErrBlocked)
			}
			sleep(ctx, wait)
		default:
			return Info{Status: Error}, fmt.Errorf("%s: HTTP %d", target, resp.StatusCode)
		}
	}
	return Info{Status: Error}, lastErr
}

// parse reads a 200 answer for domain. A body that is not a JSON object, or
// that names another object class or another domain, is an error: it says
// nothing about domain's registration.
func parse(body []byte, domain string) (Info, error) {
	var r response
	if err := json.Unmarshal(body, &r); err != nil {
		return Info{Status: Error}, fmt.Errorf("%w: %v", ErrBadResponse, err)
	}
	if r.ObjectClass != "" && !strings.EqualFold(r.ObjectClass, "domain") {
		return Info{Status: Error}, fmt.Errorf("%w: objectClassName %q", ErrBadResponse, r.ObjectClass)
	}
	if r.LDHName != "" && !strings.EqualFold(strings.TrimSuffix(r.LDHName, "."), domain) {
		return Info{Status: Error}, fmt.Errorf("%w: ldhName %q, asked for %q", ErrBadResponse, r.LDHName, domain)
	}
	info := Info{Status: Registered, EPP: r.Status}
	for _, s := range r.Status {
		s = strings.ToLower(s)
		switch {
		case strings.Contains(s, "pending delete"):
			info.Status, info.Phase = Dropping, PhasePendingDelete
		case strings.Contains(s, "redemption") && info.Phase == "":
			info.Status, info.Phase = Dropping, PhaseRedemption
		}
	}
	for _, e := range r.Events {
		if e.Action == "expiration" {
			if t, err := time.Parse(time.RFC3339, e.Date); err == nil {
				info.Expires = t.UTC()
			}
		}
	}
	return info, nil
}

func sleep(ctx context.Context, d time.Duration) {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
	}
}
