package main

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/maximilianfeix/shortfree/internal/rdap"
	"github.com/maximilianfeix/shortfree/internal/store"
)

// Evidence and coverage: what a published status is based on and how old
// that evidence is. See docs/accuracy.md.

// fresh reports whether e's last RDAP answer is recent enough to publish as
// evidence at time now. maxAge <= 0 means answers never expire.
func fresh(e *store.Entry, now time.Time, maxAge time.Duration) bool {
	return e.Checked != 0 && (maxAge <= 0 || now.Sub(e.CheckedAt()) <= maxAge)
}

// statusOf classifies a 1-3 character state entry for the public lists at
// time now:
//
//	available   RDAP 404 (not registered) no older than maxAge
//	dropping    RDAP redemption / pending delete no older than maxAge
//	unverified  not in DNS, no current RDAP evidence (queued, failed or expired)
//	unknown     no DNS answer at the last sweep and no current RDAP evidence
//	""          not listed (registered)
//
// "available" never means "registrable": registries answer 404 for reserved
// and premium names too.
func statusOf(e *store.Entry, now time.Time, maxAge time.Duration) string {
	f := fresh(e, now, maxAge)
	switch {
	case e.RDAP == string(rdap.Available) && f:
		return "available"
	case e.RDAP == string(rdap.Dropping) && f:
		return "dropping"
	case e.RDAP == string(rdap.Registered):
		return ""
	case e.DNS == "free":
		return "unverified"
	case e.DNS == "unknown":
		return "unknown"
	}
	return ""
}

// verificationOf describes the RDAP evidence behind e: "rdap_verified"
// (current answer), "stale" (the last answer is older than maxAge) or
// "pending" (no answer yet, or the last lookup failed).
func verificationOf(e *store.Entry, now time.Time, maxAge time.Duration) string {
	switch e.RDAP {
	case string(rdap.Available), string(rdap.Dropping), string(rdap.Registered):
		if fresh(e, now, maxAge) {
			return "rdap_verified"
		}
		return "stale"
	}
	return "pending"
}

// ageDays is the age of e's RDAP evidence at now in whole days, or -1 if
// there is none.
func ageDays(e *store.Entry, now time.Time) int {
	if e.Checked == 0 || e.RDAP == "" || e.RDAP == string(rdap.Error) {
		return -1
	}
	d := int(now.Sub(e.CheckedAt()) / (24 * time.Hour))
	return max(d, 0)
}

// Coverage status of a TLD, from best to worst.
const (
	covComplete = "complete" // every registrable name swept, all 4 character shards swept
	covPartial  = "partial"  // some names have no DNS answer or some shards were never swept
	covStale    = "stale"    // the latest scan of this TLD failed: older data is published
	covHeld     = "held"     // the latest scan output was rejected as implausible: older data is published
)

// TLDCoverage says what one TLD's published counts are based on.
type TLDCoverage struct {
	Status string `json:"status"` // complete, partial, stale or held
	// ScannedAt is the last complete DNS sweep of the 1-3 character names.
	ScannedAt time.Time               `json:"scanned_at"`
	Held      string                  `json:"held,omitempty"` // why the latest output was rejected
	MinLength int                     `json:"min_length"`     // shortest label the registry accepts
	Lengths   map[string]*LenCoverage `json:"lengths"`        // "1".."4"
}

// LenCoverage covers one label length.
type LenCoverage struct {
	Valid int `json:"valid"` // syntactically valid labels
	// Swept counts labels covered by the last DNS sweep (0 below the
	// registry's minimum length); DNSUnknown of them got no answer.
	Swept      int `json:"swept"`
	DNSUnknown int `json:"dns_unknown"`
	// Unscanned counts registrable labels no sweep has covered: they are
	// neither free nor taken.
	Unscanned int `json:"unscanned"`
	// 1-3 characters: RDAP evidence behind the listed names.
	RDAPVerified  int        `json:"rdap_verified,omitempty"` // available + dropping with current evidence
	RDAPPending   int        `json:"rdap_pending,omitempty"`  // not in DNS, no RDAP answer yet
	RDAPStale     int        `json:"rdap_stale,omitempty"`    // not in DNS, last RDAP answer expired
	OldestCheck   *time.Time `json:"oldest_check,omitempty"`  // oldest evidence behind a verified name
	MedianAgeDays *float64   `json:"median_age_days,omitempty"`
	// 4 characters: one entry per shard; scanned_at is omitted for shards
	// that were never swept.
	Shards []ShardCoverage `json:"shards,omitempty"`
	// Missing is true when there is no data for this length at all.
	Missing bool `json:"missing,omitempty"`
}

// ShardCoverage is one 4 character shard.
type ShardCoverage struct {
	Shard      int        `json:"shard"`
	Names      int        `json:"names"`
	ScannedAt  *time.Time `json:"scanned_at,omitempty"`
	DNSUnknown int        `json:"dns_unknown"`
}

// tldCoverage builds the coverage of tld. ls are the export's counts, cov
// the 4 character shard coverage (nil if there is no bitmap).
func tldCoverage(st *store.State, tld string, ls map[string]*LenStats, cov *store.Coverage, now time.Time, maxAge time.Duration) *TLDCoverage {
	scan := st.ScanOf(tld)
	c := &TLDCoverage{Status: covComplete, ScannedAt: scan.ScannedAt(), Held: scan.Held, MinLength: minLen(tld), Lengths: map[string]*LenCoverage{}}
	ages := map[int][]time.Time{}
	for label, e := range st.TLDs[tld] {
		if len(label) < minLen(tld) || len(label) > 3 {
			continue
		}
		switch statusOf(e, now, maxAge) {
		case "available", "dropping":
			ages[len(label)] = append(ages[len(label)], e.CheckedAt())
		}
	}
	partial := false
	for n := 1; n <= 3; n++ {
		s := ls[fmt.Sprint(n)]
		lc := &LenCoverage{Valid: s.Valid, DNSUnknown: s.Unknown, RDAPVerified: s.Available + s.Dropping, RDAPPending: s.Unverified - s.Stale, RDAPStale: s.Stale}
		if n >= minLen(tld) {
			lc.Swept = s.Valid
		}
		if t := ages[n]; len(t) > 0 {
			sort.Slice(t, func(i, j int) bool { return t[i].Before(t[j]) })
			oldest := t[0]
			lc.OldestCheck = &oldest
			med := now.Sub(t[len(t)/2]).Hours() / 24
			if len(t)%2 == 0 {
				med = (now.Sub(t[len(t)/2-1]).Hours()/24 + med) / 2
			}
			med = float64(int(med*10+0.5)) / 10
			lc.MedianAgeDays = &med
		}
		partial = partial || lc.DNSUnknown > 0
		c.Lengths[fmt.Sprint(n)] = lc
	}
	s4 := ls["4"]
	l4 := &LenCoverage{Valid: s4.Valid, Missing: s4.Missing}
	switch {
	case minLen(tld) > 4:
		// not registrable at this length: nothing to sweep
	case cov == nil:
		l4.Unscanned = s4.Valid
		partial = true
	default:
		for i := 0; i < cov.Shards; i++ {
			sc := ShardCoverage{Shard: i, Names: shardSize(i, cov.Shards)}
			if s := cov.Scanned(i); s != nil {
				t := time.Unix(s.At, 0).UTC()
				sc.ScannedAt, sc.DNSUnknown = &t, s.Unknown
				l4.Swept += sc.Names
				l4.DNSUnknown += s.Unknown
			} else {
				l4.Unscanned += sc.Names
			}
			l4.Shards = append(l4.Shards, sc)
		}
		partial = partial || l4.Unscanned > 0 || l4.DNSUnknown > 0
		if cov.Held != "" && c.Held == "" {
			c.Held = "4 characters: " + cov.Held
		}
	}
	c.Lengths["4"] = l4
	switch {
	case c.Held != "":
		c.Status = covHeld
	case now.Sub(c.ScannedAt) > 36*time.Hour:
		c.Status = covStale
	case partial:
		c.Status = covPartial
	}
	return c
}

// CoverageReport is data/coverage.json and api/v1/coverage.json.
type CoverageReport struct {
	Updated time.Time `json:"updated"`
	// MaxEvidenceAgeDays: RDAP answers older than this are published as
	// unverified (stale), never as verified.
	MaxEvidenceAgeDays float64                 `json:"max_evidence_age_days"`
	Scope              string                  `json:"scope"`
	Method             string                  `json:"method"`
	Source             *Source                 `json:"source,omitempty"`
	Order              []string                `json:"order"`
	TLDs               map[string]*TLDCoverage `json:"tlds"`
	Missing            []string                `json:"missing"` // TLDs without any data, not published
}

// Source identifies the run that produced the data (GitHub Actions only).
type Source struct {
	Repository string `json:"repository,omitempty"`
	Commit     string `json:"commit,omitempty"`
	Run        string `json:"run,omitempty"`
}

const (
	scopeText  = "ASCII LDH labels only (a-z, 0-9, inner hyphens, no -- at positions 3-4) of length 1-4 directly below each TLD. Internationalized (IDN, xn--) names are not covered."
	methodText = "1-3 characters: DNS NS query at the TLD's authoritative servers, then RDAP at the registry for names not in DNS (404 = not registered). 4 characters: DNS only, one of seven shards per day. Not registered does not mean registrable: reserved and premium names also answer 404. See https://github.com/maximilianfeix/shortfree/blob/main/docs/accuracy.md"
)

func coverageReport(s *Stats, maxAge time.Duration) *CoverageReport {
	r := &CoverageReport{
		Updated:            s.Updated,
		MaxEvidenceAgeDays: maxAge.Hours() / 24,
		Scope:              scopeText,
		Method:             methodText,
		Order:              s.Order,
		TLDs:               s.Coverage,
		Missing:            s.Missing,
	}
	if r.Missing == nil {
		r.Missing = []string{}
	}
	if repo := os.Getenv("GITHUB_REPOSITORY"); repo != "" {
		r.Source = &Source{Repository: repo, Commit: os.Getenv("GITHUB_SHA")}
		if id := os.Getenv("GITHUB_RUN_ID"); id != "" {
			r.Source.Run = os.Getenv("GITHUB_SERVER_URL") + "/" + repo + "/actions/runs/" + id
		}
	}
	return r
}
