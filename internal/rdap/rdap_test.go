package rdap

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/maximilianfeix/shortfree/internal/tlds"
)

func TestLookup(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/free.test"):
			http.NotFound(w, r)
		case strings.HasSuffix(r.URL.Path, "/drop.test"):
			w.Write([]byte(`{"status":["pending delete","inactive"],"events":[{"eventAction":"expiration","eventDate":"2026-08-08T22:28:23Z"}]}`))
		case strings.HasSuffix(r.URL.Path, "/redemption.test"):
			w.Write([]byte(`{"status":["redemption period"]}`))
		case strings.HasSuffix(r.URL.Path, "/busy.test"):
			if hits.Add(1) == 1 {
				w.Header().Set("Retry-After", "1")
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			w.Write([]byte(`{"status":["active"]}`))
		default:
			w.Write([]byte(`{"status":["client hold"]}`))
		}
	}))
	defer srv.Close()
	Endpoints["test"] = srv.URL + "/"
	defer delete(Endpoints, "test")

	c := New(100, "test")
	for label, want := range map[string]Status{
		"free": Available, "drop": Dropping, "redemption": Dropping, "held": Registered, "busy": Registered,
	} {
		info, err := c.Lookup(context.Background(), label, "test")
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
		if info.Status != want {
			t.Errorf("%s: status %s, want %s", label, info.Status, want)
		}
		if label == "drop" && !info.Expires.Equal(time.Date(2026, 8, 8, 22, 28, 23, 0, time.UTC)) {
			t.Errorf("drop: expires %v", info.Expires)
		}
	}
	if _, err := c.Lookup(context.Background(), "x", "unknown-tld"); err == nil {
		t.Error("want error for TLD without endpoint")
	}
}

func TestRate(t *testing.T) {
	if r := New(0, "").Rate(Host("dev")); r != 1.5 {
		t.Errorf("default rate for .dev = %v, want 1.5", r)
	}
	if r := New(9, "").Rate(Host("dev")); r != 9 {
		t.Errorf("override rate = %v", r)
	}
	if r := New(0, "").Rate("rdap.example"); r != 1 {
		t.Errorf("rate for unknown host = %v, want 1", r)
	}
}

func TestHost(t *testing.T) {
	for tld, want := range map[string]string{
		"com": "rdap.verisign.com", "net": "rdap.verisign.com",
		"xyz": "rdap.centralnic.com", "inc": "rdap.centralnic.com",
		"cc": "tld-rdap.verisign.com", "nope": "",
	} {
		if got := Host(tld); got != want {
			t.Errorf("Host(%q) = %q, want %q", tld, got, want)
		}
	}
}

// Every scanned TLD needs an endpoint, and every endpoint host an explicit,
// conservative rate.
func TestEveryDefaultTLDHasEndpointAndRate(t *testing.T) {
	for _, tld := range tlds.Default {
		h := Host(tld)
		if h == "" {
			t.Errorf(".%s: no RDAP endpoint", tld)
			continue
		}
		r, ok := DefaultRate[h]
		if !ok || r <= 0 || r > 10 {
			t.Errorf(".%s: host %s has rate %v (set %v)", tld, h, r, ok)
		}
	}
}

// A Retry-After far beyond MaxWait is a ban: fail fast and stop sending.
func TestLongRetryAfterBlocksHost(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Header().Set("Retry-After", "83359")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()
	Endpoints["test"] = srv.URL + "/"
	defer delete(Endpoints, "test")

	c := New(100, "test")
	start := time.Now()
	for i := 0; i < 3; i++ {
		info, err := c.Lookup(context.Background(), "x", "test")
		if !errors.Is(err, ErrBlocked) || info.Status != Error {
			t.Fatalf("lookup %d: %v, %v; want ErrBlocked", i, info.Status, err)
		}
	}
	if time.Since(start) > 2*time.Second {
		t.Errorf("blocked lookups took %s", time.Since(start))
	}
	if n := hits.Load(); n != 1 {
		t.Errorf("server saw %d requests, want 1", n)
	}
}

// TLDs on one host share one limiter.
func TestSharedHostSharesLimiter(t *testing.T) {
	c := New(0, "")
	if c.limiter(Host("xyz")) != c.limiter(Host("fm")) {
		t.Error("xyz and fm use different limiters on the same host")
	}
}
