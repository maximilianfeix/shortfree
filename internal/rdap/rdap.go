// Package rdap confirms registration status against each registry's RDAP
// service. A 404 from the authoritative registry means the name is not
// registered; a 200 carries the EPP status, which tells us when a name is
// about to drop.
package rdap

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Endpoints maps each supported TLD to its registry RDAP base URL. Sources:
// IANA bootstrap (data.iana.org/rdap/dns.json); .de and .io are not in the
// bootstrap file and point to DENIC and Identity Digital directly.
var Endpoints = map[string]string{
	"com":   "https://rdap.verisign.com/com/v1/",
	"net":   "https://rdap.verisign.com/net/v1/",
	"de":    "https://rdap.denic.de/",
	"ai":    "https://rdap.identitydigital.services/rdap/",
	"io":    "https://rdap.identitydigital.services/rdap/",
	"tools": "https://rdap.identitydigital.services/rdap/",
	"app":   "https://pubapi.registry.google/rdap/",
	"dev":   "https://pubapi.registry.google/rdap/",
}

// DefaultRate is the sustained requests per second we send to each RDAP host.
// Measured 2026-10-26: Identity Digital served 10/s without errors, Google
// answered 429 at 10/s; the others are kept conservative.
var DefaultRate = map[string]float64{
	"https://rdap.verisign.com/com/v1/":           4,
	"https://rdap.verisign.com/net/v1/":           4,
	"https://rdap.denic.de/":                      2,
	"https://rdap.identitydigital.services/rdap/": 6,
	"https://pubapi.registry.google/rdap/":        1.5,
}

// Status is the registry's view of a name.
type Status string

const (
	Available  Status = "available"  // 404: not registered
	Dropping   Status = "dropping"   // registered but in redemption / pending delete
	Registered Status = "registered" // registered (possibly without DNS)
	Error      Status = "error"      // no definitive answer
)

// Info is the result of one lookup.
type Info struct {
	Status  Status    `json:"status"`
	EPP     []string  `json:"epp,omitempty"`
	Expires time.Time `json:"expires,omitempty"`
}

// Client looks up domains with one rate limiter per RDAP host.
type Client struct {
	HTTP      *http.Client
	UserAgent string
	rps       float64

	mu     sync.Mutex
	limits map[string]*rate.Limiter
}

// New returns a client that sends at most rps requests per second per host;
// rps <= 0 uses DefaultRate.
func New(rps float64, userAgent string) *Client {
	return &Client{
		HTTP:      &http.Client{Timeout: 20 * time.Second},
		UserAgent: userAgent,
		rps:       rps,
		limits:    map[string]*rate.Limiter{},
	}
}

func (c *Client) limiter(base string) *rate.Limiter {
	c.mu.Lock()
	defer c.mu.Unlock()
	l, ok := c.limits[base]
	if !ok {
		l = rate.NewLimiter(rate.Limit(c.Rate(base)), 1)
		c.limits[base] = l
	}
	return l
}

// Rate returns the requests per second used for host.
func (c *Client) Rate(host string) float64 {
	if c.rps > 0 {
		return c.rps
	}
	if r, ok := DefaultRate[host]; ok {
		return r
	}
	return 1
}

type response struct {
	Status []string `json:"status"`
	Events []struct {
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
	url := base + "domain/" + label + "." + tld
	lim := c.limiter(base)

	var lastErr error
	for attempt := 0; attempt < 6; attempt++ {
		if err := lim.Wait(ctx); err != nil {
			return Info{Status: Error}, err
		}
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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

		switch {
		case resp.StatusCode == http.StatusNotFound:
			return Info{Status: Available}, nil
		case resp.StatusCode == http.StatusOK:
			return parse(body), nil
		case resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500:
			lastErr = fmt.Errorf("%s: HTTP %d", url, resp.StatusCode)
			wait := time.Duration(2<<attempt) * time.Second
			if s, err := strconv.Atoi(resp.Header.Get("Retry-After")); err == nil && s > 0 {
				wait = time.Duration(s) * time.Second
			}
			sleep(ctx, wait)
		default:
			return Info{Status: Error}, fmt.Errorf("%s: HTTP %d", url, resp.StatusCode)
		}
	}
	return Info{Status: Error}, lastErr
}

func parse(body []byte) Info {
	var r response
	if err := json.Unmarshal(body, &r); err != nil {
		return Info{Status: Registered}
	}
	info := Info{Status: Registered, EPP: r.Status}
	for _, s := range r.Status {
		s = strings.ToLower(s)
		if strings.Contains(s, "pending delete") || strings.Contains(s, "redemption") {
			info.Status = Dropping
		}
	}
	for _, e := range r.Events {
		if e.Action == "expiration" {
			if t, err := time.Parse(time.RFC3339, e.Date); err == nil {
				info.Expires = t.UTC()
			}
		}
	}
	return info
}

func sleep(ctx context.Context, d time.Duration) {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
	}
}
