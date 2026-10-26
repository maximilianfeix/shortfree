package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/maximilianfeix/shortfree/internal/probe"
	"github.com/maximilianfeix/shortfree/internal/rdap"
	"github.com/maximilianfeix/shortfree/internal/store"
)

// fakeRDAP serves TLD "test": canaryOK decides whether nic.test exists,
// every other name is not found.
func fakeRDAP(t *testing.T, canaryOK bool) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if canaryOK && strings.HasSuffix(r.URL.Path, "/nic.test") {
			w.Write([]byte(`{"status":["active"]}`))
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(srv.Close)
	rdap.Endpoints["test"] = srv.URL + "/"
	t.Cleanup(func() { delete(rdap.Endpoints, "test") })
}

func TestVerifyTrustsHealthyRegistry(t *testing.T) {
	fakeRDAP(t, true)
	now := time.Now().UTC()
	st := &store.State{TLDs: map[string]map[string]*store.Entry{"test": {"q7": {DNS: "free"}}}}
	verify(context.Background(), st, []string{"test"}, 1000, map[string]time.Duration{"test": time.Minute}, 24*time.Hour, now)
	if e := st.TLDs["test"]["q7"]; e.RDAP != string(rdap.Available) || e.Checked != now.Unix() {
		t.Errorf("q7 = %+v, want available", e)
	}
}

// A registry that answers 404 even for its own nic.<tld> must not turn
// queued names into "available".
func TestVerifySkipsRegistryFailingCanary(t *testing.T) {
	fakeRDAP(t, false)
	st := &store.State{TLDs: map[string]map[string]*store.Entry{"test": {"q7": {DNS: "free"}}}}
	verify(context.Background(), st, []string{"test"}, 1000, map[string]time.Duration{"test": time.Minute}, 24*time.Hour, time.Now())
	if e := st.TLDs["test"]["q7"]; e.RDAP != "" || e.Checked != 0 {
		t.Errorf("q7 = %+v, want still unverified", e)
	}
}

func TestCheckBelowMinLen(t *testing.T) {
	c := &checker{
		dns: func(_ context.Context, label, tld string) (probe.Result, error) {
			if label == "bit" {
				return probe.Taken, nil
			}
			return probe.Free, nil
		},
		rdap: func(_ context.Context, label, tld string) (rdap.Info, bool, error) {
			if label == "bit" {
				return rdap.Info{Status: rdap.Registered}, true, nil
			}
			return rdap.Info{Status: rdap.Available}, true, nil
		},
	}
	for domain, want := range map[string]string{
		"ab.so":   "not registrable (below the registry's minimum length of 3)",
		"abc.ly":  "not registrable (below the registry's minimum length of 4)",
		"bit.ly":  "taken", // grandfathered registrations stay taken
		"abc.so":  vNotRegistered,
		"q7.dev":  vNotRegistered,
		"abcd.ly": vNotRegistered,
	} {
		r := c.check(context.Background(), domain)
		if r.Verdict != want || r.Free != (want == vNotRegistered) {
			t.Errorf("%s: verdict %q free %v, want %q", domain, r.Verdict, r.Free, want)
		}
	}
}

// Long labels used to be rejected at random: names.Index overflowed.
func TestSplitDomainLongLabels(t *testing.T) {
	for _, d := range []string{"sf-ci-dac349e53fb6270b.tools", "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz.com", "a.io"} {
		if _, _, ok := splitDomain(d); !ok {
			t.Errorf("splitDomain(%q) rejected", d)
		}
	}
	for _, d := range []string{"a_b.io", "-a.io", strings.Repeat("a", 64) + ".io"} {
		if _, _, ok := splitDomain(d); ok {
			t.Errorf("splitDomain(%q) accepted", d)
		}
	}
}

func TestVerifyZeroBudgetSkipsTLD(t *testing.T) {
	fakeRDAP(t, true)
	st := &store.State{TLDs: map[string]map[string]*store.Entry{"test": {"q7": {DNS: "free"}}}}
	verify(context.Background(), st, []string{"test"}, 1000, map[string]time.Duration{"test": 0}, 24*time.Hour, time.Now())
	if e := st.TLDs["test"]["q7"]; e.RDAP != "" {
		t.Errorf("q7 = %+v, want untouched", e)
	}
}
