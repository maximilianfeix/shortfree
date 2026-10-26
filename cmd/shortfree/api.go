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

// csvHeader is the column layout of every CSV file in the API.
var csvHeader = []string{"domain", "tld", "length", "kind", "status", "first_seen", "checked_at"}

// statusOf classifies a 1-3 character state entry for the public lists:
// "available" (RDAP 404), "dropping" (redemption/pending delete),
// "unverified" (not in DNS, RDAP pending) or "" (not listed).
func statusOf(e *store.Entry) string {
	switch {
	case e.RDAP == string(rdap.Available):
		return "available"
	case e.RDAP == string(rdap.Dropping):
		return "dropping"
	case (e.RDAP == "" || e.RDAP == string(rdap.Error)) && e.DNS == "free":
		return "unverified"
	}
	return ""
}

// APIEndpoint describes one file of the static API in v1/index.json.
type APIEndpoint struct {
	Path        string `json:"path"`
	Format      string `json:"format"`
	Description string `json:"description"`
}

// APIIndex is v1/index.json.
type APIIndex struct {
	Updated   time.Time     `json:"updated"`
	Version   string        `json:"version"`
	BaseURL   string        `json:"base_url"`
	Docs      string        `json:"docs"`
	OpenAPI   string        `json:"openapi"`
	License   string        `json:"license"`
	TLDs      []string      `json:"tlds"`
	Lengths   []int         `json:"lengths"`
	Endpoints []APIEndpoint `json:"endpoints"`
}

// APIList is v1/<tld>.json (Length 0) and v1/<tld>/<n>.json.
type APIList struct {
	Updated time.Time `json:"updated"`
	TLD     string    `json:"tld"`
	Length  int       `json:"length,omitempty"`
	TLDList
}

// APIDropping is one entry of v1/dropping.json.
type APIDropping struct {
	Domain  string    `json:"domain"`
	TLD     string    `json:"tld"`
	Length  int       `json:"length"`
	Expires time.Time `json:"expires,omitzero"`
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
		{"stats.json", "json", "Counts per TLD and label length (available, dropping, unverified, 4 char free, valid)."},
		{"changes.json", "json", "Domains freed or taken during the last days, newest first."},
		{"dropping.json", "json", "Registered 1-3 character domains in redemption or pending delete on all TLDs, with expiry, soonest first."},
		{"{tld}.json", "json", "1-3 character names on one TLD: available (RDAP-verified), dropping (with expiry), unverified (not in DNS, RDAP pending)."},
		{"{tld}/{n}.json", "json", "Same as {tld}.json, limited to labels of length n (1, 2 or 3)."},
		{"free.csv", "csv", "All RDAP-verified free 1-3 character domains on all TLDs."},
		{"{tld}.csv", "csv", "All listed 1-3 character names on one TLD with status available, dropping or unverified."},
		{"badge.json", "json", "shields.io endpoint badge with the number of RDAP-verified free 1-3 character domains."},
	}
}

// writeAPI writes the versioned static API (JSON + CSV) below dir/v1.
func writeAPI(dir string, st *store.State, order []string, avail *Available, stats *Stats, changes []Change, keepDays int) error {
	v1 := filepath.Join(dir, "v1")
	idx := APIIndex{
		Updated:   avail.Updated,
		Version:   "v1",
		BaseURL:   apiBase,
		Docs:      "https://github.com/maximilianfeix/shortfree/blob/main/docs/api.md",
		OpenAPI:   "https://maximilianfeix.github.io/shortfree/api/openapi.yaml",
		License:   "CC0-1.0",
		TLDs:      order,
		Lengths:   []int{1, 2, 3},
		Endpoints: apiEndpoints(),
	}
	if err := writeCompact(filepath.Join(v1, "index.json"), idx); err != nil {
		return err
	}
	if err := writeCompact(filepath.Join(v1, "stats.json"), stats); err != nil {
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
			drop.Domains = append(drop.Domains, APIDropping{d.Name + "." + tld, tld, len(d.Name), d.Expires})
		}
		if err := writeCompact(filepath.Join(v1, tld+".json"), APIList{avail.Updated, tld, 0, *list}); err != nil {
			return err
		}
		for n := 1; n <= 3; n++ {
			if err := writeCompact(filepath.Join(v1, tld, fmt.Sprintf("%d.json", n)), APIList{avail.Updated, tld, n, byLength(list, n)}); err != nil {
				return err
			}
		}

		var rows [][]string
		for _, label := range st.Labels(tld) {
			e := st.TLDs[tld][label]
			status := statusOf(e)
			if status == "" {
				continue
			}
			row := csvRow(label, tld, status, e)
			rows = append(rows, row)
			if status == "available" {
				free = append(free, row)
			}
		}
		if err := writeCSV(filepath.Join(v1, tld+".csv"), rows); err != nil {
			return err
		}
	}
	sort.SliceStable(drop.Domains, func(i, j int) bool {
		a, b := drop.Domains[i].Expires, drop.Domains[j].Expires
		if a.IsZero() != b.IsZero() {
			return b.IsZero() // unknown expiry last
		}
		if !a.Equal(b) {
			return a.Before(b)
		}
		return drop.Domains[i].Domain < drop.Domains[j].Domain
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
	out := TLDList{Available: []string{}, Dropping: []Dropping{}, Unverified: []string{}}
	for _, s := range l.Available {
		if len(s) == n {
			out.Available = append(out.Available, s)
		}
	}
	for _, d := range l.Dropping {
		if len(d.Name) == n {
			out.Dropping = append(out.Dropping, d)
		}
	}
	for _, s := range l.Unverified {
		if len(s) == n {
			out.Unverified = append(out.Unverified, s)
		}
	}
	return out
}

func csvRow(label, tld, status string, e *store.Entry) []string {
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
	return []string{label + "." + tld, tld, fmt.Sprint(len(label)), names.Kind(label), status, day(e.Since), checked}
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
