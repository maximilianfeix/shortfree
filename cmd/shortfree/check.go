package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/maximilianfeix/shortfree/internal/names"
	"github.com/maximilianfeix/shortfree/internal/probe"
	"github.com/maximilianfeix/shortfree/internal/rdap"
)

func cmdCheck(ctx context.Context, args []string) error {
	fs := newFlags("check")
	asJSON := fs.Bool("json", false, "print the results as a JSON array")
	fs.Parse(args)
	if fs.NArg() == 0 {
		return fmt.Errorf("usage: shortfree check [-json] <domain>...")
	}
	return writeCheck(os.Stdout, newChecker().checkAll(ctx, fs.Args()), *asJSON)
}

// writeCheck prints check results as a table or, with asJSON, as a JSON
// array in input order (the fields of checkResult).
func writeCheck(w io.Writer, results []checkResult, asJSON bool) error {
	if asJSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(results)
	}
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "DOMAIN\tDNS\tRDAP\tVERDICT")
	for _, r := range results {
		verdict := r.Verdict
		if r.DNS == "-" && r.Error != "" {
			verdict = r.Error
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", r.Domain, r.DNS, r.RDAP, verdict)
	}
	return tw.Flush()
}

// checkResult is the live status of one domain, shared by "shortfree check"
// and the check_domains MCP tool.
type checkResult struct {
	Domain  string `json:"domain"`
	DNS     string `json:"dns" jsonschema:"zone status at the TLD's authoritative name servers: free (NXDOMAIN), taken (delegated), unknown, or - if not checked"`
	RDAP    string `json:"rdap" jsonschema:"registry RDAP status: available (404 = no registration), registered, dropping (redemption or pending delete), error, or - if the TLD has no RDAP endpoint here"`
	Verdict string `json:"verdict" jsonschema:"not registered (RDAP 404), taken, dropping (...), not in DNS (...), not registrable (...), unknown (...) or invalid (...)"`
	// Registration and Registrable keep two questions apart: RDAP can prove
	// that nobody holds a registration, never that a registrar will sell
	// the name (reserved, premium or restricted names also answer 404).
	Registration  string     `json:"registration" jsonschema:"registration evidence: not_registered (RDAP 404), registered (RDAP or DNS delegation), dropping (RDAP redemption or pending delete), not_delegated (not in DNS, no RDAP answer), unknown, or invalid"`
	Registrable   string     `json:"registrable" jsonschema:"no (registered, dropping, below the registry's minimum length, or invalid) or unconfirmed: shortfree never asks a registrar, so registry reservations, premium prices and eligibility rules stay unknown even when the name is not registered"`
	Free          bool       `json:"free" jsonschema:"true only when the registry's RDAP service answered 404 (registration: not_registered). It does not mean the name can be bought: see registrable"`
	PremiumLikely bool       `json:"premium_likely" jsonschema:"1-2 character name outside .com/.net/.de: often registry-reserved or premium-priced even when RDAP says not registered"`
	Expires       *time.Time `json:"expires,omitempty" jsonschema:"RDAP expiration date. Not a drop date: registries delete names on their own schedules"`
	Phase         string     `json:"phase,omitempty" jsonschema:"dropping names only: redemption or pending_delete"`
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
	// A person or a model is waiting: two tries with short timeouts, then
	// report "RDAP failed" instead of blocking for minutes.
	rc.Attempts = 2
	rc.HTTP.Timeout = 8 * time.Second
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

// validLabel accepts any ASCII LDH label up to 63 characters: a-z, 0-9 and
// inner hyphens. (names.Index cannot be used as a character check here: its
// base-37 value overflows for labels longer than 12 characters.)
func validLabel(s string) bool {
	if s == "" || len(s) > 63 || s[0] == '-' || s[len(s)-1] == '-' {
		return false
	}
	for _, c := range []byte(s) {
		if strings.IndexByte(names.Alphabet, c) < 0 {
			return false
		}
	}
	return true
}

// outOfScope explains why d is outside what shortfree covers, or returns "".
// shortfree covers ASCII LDH labels only: internationalized names (Unicode
// or their xn-- A-labels) need IDNA validation and registry-specific
// language tables, which it does not implement. Labels with "--" at
// positions 3-4 are reserved for such encodings (RFC 5891).
func outOfScope(d string) string {
	for _, r := range d {
		if r > 0x7f {
			return "invalid (internationalized domain names are not supported: shortfree checks ASCII labels of a-z, 0-9 and hyphens only)"
		}
	}
	for _, part := range strings.Split(d, ".") {
		if strings.HasPrefix(part, "xn--") {
			return "invalid (IDN/punycode xn-- labels are not supported: shortfree checks ASCII labels of a-z, 0-9 and hyphens only)"
		}
		if len(part) >= 4 && part[2:4] == "--" {
			return "invalid (hyphens at positions 3 and 4 are reserved for encoded labels such as xn--)"
		}
	}
	return ""
}

// check looks up one domain live: DNS at the TLD's own servers, then RDAP.
func (c *checker) check(ctx context.Context, raw string) checkResult {
	d := normalizeDomain(raw)
	r := checkResult{Domain: d, DNS: "-", RDAP: "-", Registration: "invalid", Registrable: "no"}
	if why := outOfScope(d); why != "" {
		r.Verdict = why
		return r
	}
	label, tld, ok := splitDomain(d)
	if !ok {
		r.Verdict = "invalid (want label.tld with a-z, 0-9 and inner hyphens)"
		return r
	}
	r.PremiumLikely = premiumLikely(label, tld)
	r.Registration, r.Registrable = "unknown", "unconfirmed"
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
		r.Registration = registrationOf(dns, info.Status)
		r.EPP = info.EPP
		r.Phase = info.Phase
		if !info.Expires.IsZero() {
			t := info.Expires
			r.Expires = &t
		}
	case dns == probe.Free:
		r.Verdict = "not in DNS (no RDAP for this TLD)"
		r.Registration = "not_delegated"
	case dns == probe.Taken:
		r.Verdict = vTaken
		r.Registration = "registered"
	default:
		r.Verdict = "unknown (DNS failed, no RDAP for this TLD)"
	}
	if r.Registration == "registered" || r.Registration == "dropping" {
		r.Registrable = "no"
	}
	r.Free = r.Registration == "not_registered"
	// Registries answer "not found" for names they would never accept, too.
	// Registered short names (grandfathered) still show as taken.
	if n := minLen(tld); len(label) < n && r.Registrable != "no" {
		r.Verdict = fmt.Sprintf("not registrable (below the registry's minimum length of %d)", n)
		r.Registrable = "no"
		r.Free = false
	}
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
