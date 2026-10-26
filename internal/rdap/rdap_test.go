package rdap

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
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
	if r := New(0, "").Rate(Endpoints["dev"]); r != DefaultRate[Endpoints["dev"]] {
		t.Errorf("default rate for .dev = %v", r)
	}
	if r := New(9, "").Rate(Endpoints["dev"]); r != 9 {
		t.Errorf("override rate = %v", r)
	}
}
