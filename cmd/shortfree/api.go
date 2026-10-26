package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"time"

	"github.com/maximilianfeix/shortfree/internal/names"
	"github.com/maximilianfeix/shortfree/internal/rdap"
	"github.com/maximilianfeix/shortfree/internal/store"
)

// apiBase is where GitHub Pages serves the static API.
const apiBase = "https://maximilianfeix.github.io/shortfree/api/v1/"

// csvHeader is the column layout of every CSV file in the API. New columns
// are only ever appended.
var csvHeader = []string{"domain", "tld", "length", "kind", "status", "first_seen", "checked_at", "verification", "evidence_age_days"}

// statusDefinitions is published in index.json so every status says what
// it is (and is not) based on.
var statusDefinitions = map[string]string{
	"available":  "Not registered: not in DNS and the registry's RDAP service answered 404 within max_evidence_age_days. Not a promise that a registrar sells it: reserved, premium and restricted names also answer 404.",
	"dropping":   "Registered, in redemption period or pending delete (RDAP). rdap_expiration is the registry's expiration date, not a drop date.",
	"unverified": "Not in DNS, but no current RDAP answer: check queued, failed, or the last answer is older than max_evidence_age_days (then also listed in stale).",
	"unknown":    "No DNS answer at the last sweep and no current RDAP answer: not checked. Neither free nor taken.",
}

// APIEndpoint describes one file of the static API in v1/index.json.
type APIEndpoint struct {
	Path        string `json:"path"`
	Format      string `json:"format"`
	Description string `json:"description"`
}

// APIIndex is v1/index.json.
type APIIndex struct {
	Updated            time.Time         `json:"updated"`
	Version            string            `json:"version"`
	BaseURL            string            `json:"base_url"`
	Docs               string            `json:"docs"`
	Accuracy           string            `json:"accuracy"`
	OpenAPI            string            `json:"openapi"`
	License            string            `json:"license"`
	TLDs               []string          `json:"tlds"`
	Lengths            []int             `json:"lengths"`
	MaxEvidenceAgeDays float64           `json:"max_evidence_age_days"`
	Statuses           map[string]string `json:"statuses"`
	Endpoints          []APIEndpoint     `json:"endpoints"`
}

// APIList is v1/<tld>.json (Length 0) and v1/<tld>/<n>.json.
type APIList struct {
	Updated time.Time `json:"updated"`
	TLD     string    `json:"tld"`
	Length  int       `json:"length,omitempty"`
	// MaxEvidenceAgeDays: every name in available and dropping was checked
	// via RDAP at most this many days before updated.
	MaxEvidenceAgeDays float64 `json:"max_evidence_age_days"`
	// OldestCheck is the oldest RDAP check behind a name in available.
	OldestCheck *time.Time `json:"oldest_check,omitempty"`
	TLDList
}

// APIDropping is one entry of v1/dropping.json.
type APIDropping struct {
	Domain string `json:"domain"`
	TLD    string `json:"tld"`
	Length int    `json:"length"`
	// Expires is the deprecated name of RDAPExpiration.
	Expires        time.Time `json:"expires,omitzero"`
	RDAPExpiration time.Time `json:"rdap_expiration,omitzero"` // not a drop date
	Phase          string    `json:"phase,omitempty"`          // redemption or pending_delete
	CheckedAt      time.Time `json:"checked_at,omitzero"`
}

// APIDroppingList is v1/dropping.json.
type APIDroppingList struct {
	Updated time.Time     `json:"updated"`
	Count   int           `json:"count"`
	Domains []APIDropping `json:"domains"`
}

// APIBadge is v1/badge.json in the shields.io endpoint format
// (https://shields.io/badges/endpoint-badge).
type APIBadge struct {
	SchemaVersion int    `json:"schemaVersion"`
	Label         string `json:"label"`
	Message       string `json:"message"`
	Color         string `json:"color"`
}

// badgeOf returns the README badge for n RDAP-verified free domains.
func badgeOf(n int) APIBadge {
	return APIBadge{SchemaVersion: 1, Label: "free short domains", Message: num(n) + " verified", Color: "1FC16B"}
}

// APIChanges is v1/changes.json.
type APIChanges struct {
	Updated time.Time `json:"updated"`
	Days    int       `json:"days"`
	Changes []Change  `json:"changes"`
}

func apiEndpoints() []APIEndpoint {
	return []APIEndpoint{
		{"index.json", "json", "This index: update time, TLDs and endpoint list."},
		{"stats.json", "json", "Counts per TLD and label length (available, dropping, unverified, stale, unknown, 4 char free, unscanned, valid) plus coverage."},
		{"coverage.json", "json", "What the counts are based on: last sweep per TLD, DNS unknowns, unscanned 4 character shards, RDAP evidence age, held or missing TLDs."},
		{"changes.json", "json", "Domains freed (RDAP registered/dropping -> 404), newly listed or taken during the last days, newest first."},
		{"dropping.json", "json", "Registered 1-3 character domains in redemption or pending delete on all TLDs: pending delete first, with lifecycle phase and RDAP expiration date (not a drop date)."},
		{"{tld}.json", "json", "1-3 character names on one TLD: available (RDAP 404, not registered), dropping, unverified (not in DNS, no current RDAP answer), stale (part of unverified), unknown (no DNS answer)."},
		{"{tld}/{n}.json", "json", "Same as {tld}.json, limited to labels of length n (1, 2 or 3)."},
		{"free.csv", "csv", "All 1-3 character domains with current RDAP evidence of not being registered, on all TLDs."},
		{"{tld}.csv", "csv", "All listed 1-3 character names on one TLD with status, verification and evidence age."},
		{"badge.json", "json", "shields.io endpoint badge with the number of RDAP-verified (not registered) 1-3 character domains."},
	}
}

// writeAPI writes the versioned static API (JSON + CSV) below dir/v1.
func writeAPI(dir string, st *store.State, order []string, avail *Available, stats *Stats, report *CoverageReport, changes []Change, keepDays int, maxAge time.Duration) error {
	v1 := filepath.Join(dir, "v1")
	now := avail.Updated
	ageLimit := maxAge.Hours() / 24
	idx := APIIndex{
		Updated:            avail.Updated,
		Version:            "v1",
		BaseURL:            apiBase,
		Docs:               "https://github.com/maximilianfeix/shortfree/blob/main/docs/api.md",
		Accuracy:           "https://github.com/maximilianfeix/shortfree/blob/main/docs/accuracy.md",
		OpenAPI:            "https://maximilianfeix.github.io/shortfree/api/openapi.yaml",
		License:            "CC0-1.0",
		TLDs:               order,
		Lengths:            []int{1, 2, 3},
		MaxEvidenceAgeDays: ageLimit,
		Statuses:           statusDefinitions,
		Endpoints:          apiEndpoints(),
	}
	if err := writeCompact(filepath.Join(v1, "index.json"), idx); err != nil {
		return err
	}
	if err := writeCompact(filepath.Join(v1, "stats.json"), stats); err != nil {
		return err
	}
	if err := writeCompact(filepath.Join(v1, "coverage.json"), report); err != nil {
		return err
	}
	if err := writeCompact(filepath.Join(v1, "changes.json"), APIChanges{avail.Updated, keepDays, changes}); err != nil {
		return err
	}

	drop := APIDroppingList{Updated: avail.Updated, Domains: []APIDropping{}}
	var free [][]string
	for _, tld := range order {
		list := avail.TLDs[tld]
		if list == nil {
			continue
		}
		for _, d := range list.Dropping {
			drop.Domains = append(drop.Domains, APIDropping{d.Name + "." + tld, tld, len(d.Name), d.RDAPExpiration, d.RDAPExpiration, d.Phase, d.CheckedAt})
		}
		if err := writeCompact(filepath.Join(v1, tld+".json"), APIList{avail.Updated, tld, 0, ageLimit, oldestCheck(st, tld, list.Available), *list}); err != nil {
			return err
		}
		for n := 1; n <= 3; n++ {
			l := byLength(list, n)
			if err := writeCompact(filepath.Join(v1, tld, fmt.Sprintf("%d.json", n)), APIList{avail.Updated, tld, n, ageLimit, oldestCheck(st, tld, l.Available), l}); err != nil {
				return err
			}
		}

		var rows [][]string
		for _, label := range st.Labels(tld) {
			if len(label) < minLen(tld) {
				continue
			}
			e := st.TLDs[tld][label]
			status := statusOf(e, now, maxAge)
			if status == "" {
				continue
			}
			row := csvRow(label, tld, status, e, now, maxAge)
			rows = append(rows, row)
			if status == "available" {
				free = append(free, row)
			}
		}
		if err := writeCSV(filepath.Join(v1, tld+".csv"), rows); err != nil {
			return err
		}
	}
	// Pending delete is the last phase before deletion; the expiration date
	// says nothing about when a name drops, so it does not decide the order.
	phaseRank := func(p string) int {
		switch p {
		case rdap.PhasePendingDelete:
			return 0
		case rdap.PhaseRedemption:
			return 1
		}
		return 2
	}
	sort.SliceStable(drop.Domains, func(i, j int) bool {
		a, b := drop.Domains[i], drop.Domains[j]
		if phaseRank(a.Phase) != phaseRank(b.Phase) {
			return phaseRank(a.Phase) < phaseRank(b.Phase)
		}
		if len(a.Domain) != len(b.Domain) {
			return len(a.Domain) < len(b.Domain)
		}
		return a.Domain < b.Domain
	})
	drop.Count = len(drop.Domains)
	if err := writeCompact(filepath.Join(v1, "dropping.json"), drop); err != nil {
		return err
	}
	if err := writeCompact(filepath.Join(v1, "badge.json"), badgeOf(len(free))); err != nil {
		return err
	}
	return writeCSV(filepath.Join(v1, "free.csv"), free)
}

// byLength returns the part of l with labels of length n; slices are never nil.
func byLength(l *TLDList, n int) TLDList {
	out := *newTLDList()
	pick := func(src []string, dst *[]string) {
		for _, s := range src {
			if len(s) == n {
				*dst = append(*dst, s)
			}
		}
	}
	pick(l.Available, &out.Available)
	pick(l.Unverified, &out.Unverified)
	pick(l.Stale, &out.Stale)
	pick(l.Unknown, &out.Unknown)
	for _, d := range l.Dropping {
		if len(d.Name) == n {
			out.Dropping = append(out.Dropping, d)
		}
	}
	for k, v := range l.Checked {
		if len(k) == n {
			out.Checked[k] = v
		}
	}
	return out
}

// oldestCheck returns the oldest RDAP check among labels of tld, nil if none.
func oldestCheck(st *store.State, tld string, labels []string) *time.Time {
	var oldest *time.Time
	for _, l := range labels {
		if e := st.TLDs[tld][l]; e != nil && e.Checked != 0 {
			if t := e.CheckedAt(); oldest == nil || t.Before(*oldest) {
				oldest = &t
			}
		}
	}
	return oldest
}

func csvRow(label, tld, status string, e *store.Entry, now time.Time, maxAge time.Duration) []string {
	day := func(sec int64) string {
		if sec == 0 {
			return ""
		}
		return time.Unix(sec, 0).UTC().Format("2006-01-02")
	}
	checked := ""
	if e.Checked != 0 {
		checked = e.CheckedAt().Format(time.RFC3339)
	}
	age := ""
	if d := ageDays(e, now); d >= 0 {
		age = fmt.Sprint(d)
	}
	return []string{label + "." + tld, tld, fmt.Sprint(len(label)), names.Kind(label), status, day(e.Since), checked, verificationOf(e, now, maxAge), age}
}

func writeCSV(path string, rows [][]string) error {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	w.Write(csvHeader)
	w.WriteAll(rows) // flushes
	if err := w.Error(); err != nil {
		return err
	}
	return store.WriteAtomic(path, buf.Bytes())
}

func writeCompact(path string, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return store.WriteAtomic(path, append(b, '\n'))
}
