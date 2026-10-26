package main

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/maximilianfeix/shortfree/internal/names"
	"github.com/maximilianfeix/shortfree/internal/probe"
	"github.com/maximilianfeix/shortfree/internal/rdap"
)

// checkResult is the live status of one domain, shared by "shortfree check"
// and the check_domains MCP tool.
type checkResult struct {
	Domain        string     `json:"domain"`
	DNS           string     `json:"dns" jsonschema:"zone status at the TLD's authoritative name servers: free (NXDOMAIN), taken (delegated), unknown, or - if not checked"`
	RDAP          string     `json:"rdap" jsonschema:"registry RDAP status: available (404), registered, dropping (redemption or pending delete), error, or - if the TLD has no RDAP endpoint here"`
	Verdict       string     `json:"verdict" jsonschema:"FREE, taken, dropping soon, probably free (...), unknown (...) or invalid (...)"`
	Free          bool       `json:"free" jsonschema:"true only when the registry confirmed the name is not registered (verdict FREE)"`
	PremiumLikely bool       `json:"premium_likely" jsonschema:"1-2 character name outside .com/.net/.de: often registry-reserved or premium-priced even when RDAP says available"`
	Expires       *time.Time `json:"expires,omitempty" jsonschema:"expiration date from RDAP, mainly useful for dropping names"`
	EPP           []string   `json:"epp_status,omitempty" jsonschema:"raw EPP status values from RDAP"`
	Error         string     `json:"error,omitempty"`
}

// premiumLikely mirrors the site: 1-2 character names on new TLDs are
// usually reserved or premium-priced, which RDAP reports as unregistered.
func premiumLikely(label, tld string) bool {
	return len(label) <= 2 && tld != "com" && tld != "net" && tld != "de"
}

// checker runs live checks. dns and rdap are swappable for tests.
type checker struct {
	dns  func(ctx context.Context, label, tld string) (probe.Result, error)
	rdap func(ctx context.Context, label, tld string) (info rdap.Info, known bool, err error)

	mu      sync.Mutex
	probers map[string]*probe.Prober
}

func newChecker() *checker {
	c := &checker{probers: map[string]*probe.Prober{}}
	rc := rdap.New(0, userAgent)
	c.dns = c.liveDNS
	c.rdap = func(ctx context.Context, label, tld string) (rdap.Info, bool, error) {
		if _, ok := rdap.Endpoints[tld]; !ok {
			return rdap.Info{}, false, nil
		}
		info, err := rc.Lookup(ctx, label, tld)
		return info, true, err
	}
	return c
}

func (c *checker) liveDNS(ctx context.Context, label, tld string) (probe.Result, error) {
	c.mu.Lock()
	p, ok := c.probers[tld]
	if !ok {
		var err error
		if p, err = probe.New(ctx, tld, 20); err != nil {
			c.mu.Unlock()
			return probe.Unknown, err
		}
		c.probers[tld] = p
	}
	c.mu.Unlock()
	return p.Check(ctx, label), nil
}

// normalizeDomain accepts "Q7.Tools.", " q7.tools " or "https://q7.tools/x".
func normalizeDomain(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.TrimPrefix(strings.TrimPrefix(s, "https://"), "http://")
	if i := strings.IndexAny(s, "/?#"); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSuffix(s, ".")
}

// splitDomain returns label and TLD of a second-level domain.
func splitDomain(d string) (label, tld string, ok bool) {
	label, tld, ok = strings.Cut(d, ".")
	if !ok || tld == "" || strings.Contains(tld, ".") || !validLabel(label) || !validLabel(tld) {
		return "", "", false
	}
	return label, tld, true
}

// validLabel accepts any LDH label up to 63 characters (IDN "xn--" included).
func validLabel(s string) bool {
	return s != "" && len(s) <= 63 && names.Index(s) >= 0 && s[0] != '-' && s[len(s)-1] != '-'
}

// check looks up one domain live: DNS at the TLD's own servers, then RDAP.
func (c *checker) check(ctx context.Context, raw string) checkResult {
	d := normalizeDomain(raw)
	r := checkResult{Domain: d, DNS: "-", RDAP: "-"}
	label, tld, ok := splitDomain(d)
	if !ok {
		r.Verdict = "invalid (want label.tld)"
		return r
	}
	r.PremiumLikely = premiumLikely(label, tld)
	dns, err := c.dns(ctx, label, tld)
	if err != nil {
		r.Verdict = "error"
		r.Error = err.Error()
		return r
	}
	r.DNS = dns.String()
	info, known, err := c.rdap(ctx, label, tld)
	switch {
	case known:
		r.RDAP = string(info.Status)
		if err != nil {
			r.RDAP = string(rdap.Error)
			r.Error = err.Error()
			info.Status = rdap.Error
		}
		r.Verdict = verdictOf(dns, info.Status)
		r.EPP = info.EPP
		if !info.Expires.IsZero() {
			t := info.Expires
			r.Expires = &t
		}
	case dns == probe.Free:
		r.Verdict = "probably free (no RDAP for this TLD)"
	case dns == probe.Taken:
		r.Verdict = "taken"
	default:
		r.Verdict = "unknown (DNS failed, no RDAP for this TLD)"
	}
	r.Free = r.Verdict == "FREE"
	return r
}

// checkAll checks domains concurrently and returns results in input order.
// RDAP and DNS rate limits are enforced per registry inside the clients.
func (c *checker) checkAll(ctx context.Context, domains []string) []checkResult {
	out := make([]checkResult, len(domains))
	jobs := make(chan int)
	var wg sync.WaitGroup
	for w := 0; w < 8 && w < len(domains); w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				out[i] = c.check(ctx, domains[i])
			}
		}()
	}
	for i := range domains {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	return out
}
