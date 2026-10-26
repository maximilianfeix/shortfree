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
	if info, _ := c.Lookup(context.Background(), "drop", "test"); info.Phase != PhasePendingDelete {
		t.Errorf("drop: phase %q, want pending_delete", info.Phase)
	}
	if info, _ := c.Lookup(context.Background(), "redemption", "test"); info.Phase != PhaseRedemption {
		t.Errorf("redemption: phase %q, want redemption", info.Phase)
	}
}

// Failure modes of broken or misconfigured endpoints: none of them may turn
// into "available" (or silently into "registered").
func TestLookupRejectsBadResponses(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Path[strings.LastIndexByte(r.URL.Path, '/')+1:]
		switch name {
		case "html404.test": // moved endpoint: the web server's own 404 page
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("<html><body>Not Found</body></html>"))
		case "html200.test": // CDN interstitial or landing page
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte("<html>please wait</html>"))
		case "garbage.test":
			w.Header().Set("Content-Type", "application/rdap+json")
			w.Write([]byte("{not json"))
		case "entity.test":
			w.Header().Set("Content-Type", "application/rdap+json")
			w.Write([]byte(`{"objectClassName":"entity","handle":"X"}`))
		case "other.test":
			w.Header().Set("Content-Type", "application/rdap+json")
			w.Write([]byte(`{"objectClassName":"domain","ldhName":"something-else.test","status":["active"]}`))
		case "ok.test":
			w.Header().Set("Content-Type", "application/rdap+json")
			w.Write([]byte(`{"objectClassName":"domain","ldhName":"OK.TEST.","status":["active"]}`))
		case "json404.test":
			w.Header().Set("Content-Type", "application/rdap+json")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"errorCode":404,"title":"Not Found"}`))
		case "empty404.test":
			w.WriteHeader(http.StatusNotFound)
		case "forbidden.test":
			w.WriteHeader(http.StatusForbidden)
		}
	}))
	defer srv.Close()
	Endpoints["test"] = srv.URL + "/"
	defer delete(Endpoints, "test")
	c := New(1000, "test")
	for label, want := range map[string]Status{
		"html404": Error, "html200": Error, "garbage": Error, "entity": Error, "other": Error, "forbidden": Error,
		"ok": Registered, "json404": Available, "empty404": Available,
	} {
		info, err := c.Lookup(context.Background(), label, "test")
		if info.Status != want {
			t.Errorf("%s: status %s (err %v), want %s", label, info.Status, err, want)
		}
		if (want == Error) != (err != nil) {
			t.Errorf("%s: err = %v", label, err)
		}
		if label != "forbidden" && want == Error && !errors.Is(err, ErrBadResponse) {
			t.Errorf("%s: err %v, want ErrBadResponse", label, err)
		}
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

func TestSharedRate(t *testing.T) {
	id, google := Host("ai"), Host("dev")
	cases := []struct {
		name         string
		host         string
		scanned, all []string
		want         float64
	}{
		{"one of four on Identity Digital", id, []string{"ai"}, tlds.Default, 0.5},
		{"one of two on Google", google, []string{"dev"}, tlds.Default, 0.75},
		{"alone on its host", Host("de"), []string{"de"}, tlds.Default, 2},
		{"all TLDs: full total", id, tlds.Default, tlds.Default, 2},
		{"host not scanned", id, []string{"de"}, tlds.Default, 0},
		{"extra TLD outside the list counts too", id, []string{"ai"}, []string{"io"}, 1},
	}
	for _, c := range cases {
		if got := SharedRate(c.host, c.scanned, c.all); got != c.want {
			t.Errorf("%s: SharedRate = %v, want %v", c.name, got, c.want)
		}
	}
	// The matrix jobs of one host never exceed its total together.
	for host, total := range DefaultRate {
		var sum float64
		for _, tld := range tlds.Default {
			sum += SharedRate(host, []string{tld}, tlds.Default)
		}
		if sum > total+1e-9 {
			t.Errorf("%s: per-TLD jobs add up to %v > total %v", host, sum, total)
		}
	}
}

func TestClientUsesShare(t *testing.T) {
	c := New(0, "")
	c.ShareWith([]string{"io"}, tlds.Default)
	if r := c.Rate(Host("io")); r != 0.5 {
		t.Errorf("rate for .io job = %v, want 0.5", r)
	}
	c = New(9, "")
	c.ShareWith([]string{"io"}, tlds.Default)
	if r := c.Rate(Host("io")); r != 9 {
		t.Errorf("explicit -rdap-rps must win, got %v", r)
	}
}

// TLDs on one host share one limiter.
func TestSharedHostSharesLimiter(t *testing.T) {
	c := New(0, "")
	if c.limiter(Host("xyz")) != c.limiter(Host("fm")) {
		t.Error("xyz and fm use different limiters on the same host")
	}
}
