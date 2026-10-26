package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/maximilianfeix/shortfree/internal/names"
	"github.com/maximilianfeix/shortfree/internal/store"
)

// Dropping is a registered name in redemption or pending delete.
type Dropping struct {
	Name string `json:"name"`
	// Expires is the deprecated name of RDAPExpiration, kept for v1
	// compatibility.
	Expires time.Time `json:"expires,omitzero"`
	// RDAPExpiration is the registry's expiration event. It is NOT a drop
	// date: registries delete names on their own schedules.
	RDAPExpiration time.Time `json:"rdap_expiration,omitzero"`
	Phase          string    `json:"phase,omitempty"`     // redemption or pending_delete
	CheckedAt      time.Time `json:"checked_at,omitzero"` // RDAP check this is based on
}

// TLDList is the public per-TLD result for 1-3 character names.
type TLDList struct {
	Available  []string   `json:"available"`  // RDAP 404 within the evidence age limit
	Dropping   []Dropping `json:"dropping"`   // RDAP redemption / pending delete
	Unverified []string   `json:"unverified"` // not in DNS, no current RDAP evidence
	Stale      []string   `json:"stale"`      // part of Unverified: RDAP evidence expired
	Unknown    []string   `json:"unknown"`    // no DNS answer at the last sweep: not checked
}

func newTLDList() *TLDList {
	return &TLDList{Available: []string{}, Dropping: []Dropping{}, Unverified: []string{}, Stale: []string{}, Unknown: []string{}}
}

// Available is data/available.json.
type Available struct {
	Updated time.Time           `json:"updated"`
	TLDs    map[string]*TLDList `json:"tlds"`
}

// LenStats counts one label length on one TLD.
type LenStats struct {
	Available  int `json:"available,omitempty"`
	Dropping   int `json:"dropping,omitempty"`
	Unverified int `json:"unverified,omitempty"`
	Stale      int `json:"stale,omitempty"`   // part of Unverified: RDAP evidence expired
	Unknown    int `json:"unknown,omitempty"` // no DNS answer at the last sweep
	Free       int `json:"free,omitempty"`    // 4 chars: not delegated (DNS only)
	// Unscanned counts valid names no sweep has covered yet (4 characters:
	// shards never swept). They are neither free nor taken.
	Unscanned int `json:"unscanned,omitempty"`
	// Missing means there is no data for this length at all: the counts
	// are unknown, not zero.
	Missing bool `json:"missing,omitempty"`
	Valid   int  `json:"valid"`
}

// Stats is data/stats.json.
type Stats struct {
	Updated time.Time                       `json:"updated"`
	TLDs    map[string]map[string]*LenStats `json:"tlds"`
	Order   []string                        `json:"order"`
	// Coverage says per TLD what the counts are based on (see coverage.json).
	Coverage map[string]*TLDCoverage `json:"coverage,omitempty"`
	// Missing lists TLDs without any published data (e.g. a new TLD whose
	// first scan failed). They are left out of Order and TLDs instead of
	// being shown with 0 free names.
	Missing []string `json:"missing,omitempty"`
}

// Change is one entry of data/changes.json.
type Change struct {
	Date   string `json:"date"`
	Domain string `json:"domain"`
	// Event is "freed" (RDAP status went from registered or dropping to
	// 404), "newly_listed" (now listed as available without an observed
	// registration before, e.g. its first RDAP check) or "taken".
	Event string `json:"event"`
}

// Change events.
const (
	evFreed       = "freed"
	evNewlyListed = "newly_listed"
	evTaken       = "taken"
)

func validCount(n int) int {
	c := 0
	names.Each(n, 0, 1, func(int, string) { c++ })
	return c
}

// defaultMaxAge is how long an RDAP answer counts as current evidence. The
// scan rechecks after 10 days; names whose evidence is older than this are
// published as unverified (stale), never as verified.
const defaultMaxAge = 14 * 24 * time.Hour

func cmdExport(args []string) error {
	fs := newFlags("export")
	tlds := fs.String("tlds", "", "comma-separated TLDs (default: all supported)")
	statePath := fs.String("state", "state/state.json.gz", "state file (.gz = compressed)")
	bitmaps := fs.String("bitmaps", "bitmaps", "directory with 4 character bitmaps and their .cov.json coverage")
	out := fs.String("out", "data", "output directory for public data")
	four := fs.String("four-out", "dist/4", "output directory for 4 character text lists (empty = skip)")
	readme := fs.String("readme", "README.md", "README to update between shortfree markers (empty = skip)")
	keepDays := fs.Int("changes-days", 30, "days of history kept in changes.json")
	apiOut := fs.String("api-out", "dist/api", "output directory for the static JSON/CSV API (empty = skip)")
	maxAge := fs.Duration("max-age", defaultMaxAge, "oldest RDAP answer still published as verified; older ones become unverified (stale)")
	shards := fs.Int("shards", 7, "number of 4 character shards (for bitmaps without coverage file)")
	fs.Parse(args)

	st, err := store.Load(*statePath)
	if err != nil {
		return err
	}
	now := st.Updated
	var order, missing []string
	for _, tld := range parseTLDs(*tlds) {
		if st.ScanOf(tld) == nil {
			// Never publish "0 free" for a TLD that was never scanned.
			fmt.Fprintf(os.Stderr, "export: .%s has no scan data, left out\n", tld)
			missing = append(missing, tld)
			continue
		}
		order = append(order, tld)
	}
	if len(order) == 0 {
		return fmt.Errorf("export: none of the TLDs has scan data, refusing to publish empty lists")
	}
	avail := Available{Updated: st.Updated, TLDs: map[string]*TLDList{}}
	stats := Stats{Updated: st.Updated, TLDs: map[string]map[string]*LenStats{}, Order: order, Coverage: map[string]*TLDCoverage{}, Missing: missing}
	valid := map[int]int{}
	for n := 1; n <= 4; n++ {
		valid[n] = validCount(n)
	}

	for _, tld := range order {
		list := newTLDList()
		ls := map[string]*LenStats{}
		for n := 1; n <= 4; n++ {
			ls[fmt.Sprint(n)] = &LenStats{Valid: valid[n]}
		}
		for _, label := range st.Labels(tld) {
			if len(label) < minLen(tld) {
				continue // the registry does not register names this short
			}
			e := st.TLDs[tld][label]
			s := ls[fmt.Sprint(len(label))]
			switch statusOf(e, now, *maxAge) {
			case "available":
				list.Available = append(list.Available, label)
				s.Available++
			case "dropping":
				list.Dropping = append(list.Dropping, droppingOf(label, e))
				s.Dropping++
			case "unverified":
				list.Unverified = append(list.Unverified, label)
				s.Unverified++
				if verificationOf(e, now, *maxAge) == "stale" {
					list.Stale = append(list.Stale, label)
					s.Stale++
				}
			case "unknown":
				list.Unknown = append(list.Unknown, label)
				s.Unknown++
			}
		}
		avail.TLDs[tld] = list

		bmPath := filepath.Join(*bitmaps, tld+".bin")
		bm, err := store.ReadBitmap(bmPath, names.Space(4))
		var cov *store.Coverage
		switch {
		case errors.Is(err, os.ErrNotExist):
			bm = nil
			ls["4"].Missing = minLen(tld) <= 4
		case err != nil:
			return err
		default:
			if cov, err = store.LoadCoverage(store.CoveragePath(bmPath), *shards); err != nil {
				return err
			}
			ls["4"].Free = bm.Count()
			ls["4"].Unscanned, ls["4"].Unknown = shardCounts(cov)
		}
		stats.TLDs[tld] = ls
		stats.Coverage[tld] = tldCoverage(st, tld, ls, cov, now, *maxAge)
		if *four != "" && bm != nil {
			if err := writeFour(filepath.Join(*four, tld+".txt"), bm, tld); err != nil {
				return err
			}
		}
		var txt bytes.Buffer
		for _, l := range list.Available {
			fmt.Fprintf(&txt, "%s.%s\n", l, tld)
		}
		if err := store.WriteAtomic(filepath.Join(*out, "txt", tld+".txt"), txt.Bytes()); err != nil {
			return err
		}
	}

	availPath := filepath.Join(*out, "available.json")
	freedSince := func(domain string, since time.Time) bool {
		e := st.TLDs[tldOf(domain)][labelOf(domain)]
		return e != nil && e.Freed > since.Unix()
	}
	changes, err := diffChanges(availPath, filepath.Join(*out, "changes.json"), &avail, *keepDays, freedSince)
	if err != nil {
		return err
	}
	report := coverageReport(&stats, *maxAge)
	if err := writeJSON(availPath, avail); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(*out, "stats.json"), stats); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(*out, "coverage.json"), report); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(*out, "changes.json"), changes); err != nil {
		return err
	}
	if *apiOut != "" {
		if err := writeAPI(*apiOut, st, order, &avail, &stats, report, changes, *keepDays, *maxAge); err != nil {
			return err
		}
	}
	if *readme != "" {
		return updateReadme(*readme, &avail, &stats, changes)
	}
	return nil
}

func droppingOf(label string, e *store.Entry) Dropping {
	return Dropping{Name: label, Expires: e.ExpiresAt(), RDAPExpiration: e.ExpiresAt(), Phase: e.Phase, CheckedAt: checkedAt(e)}
}

func checkedAt(e *store.Entry) time.Time {
	if e.Checked == 0 {
		return time.Time{}
	}
	return e.CheckedAt()
}

// shardCounts returns how many valid 4 character names lie in shards that
// were never swept, and how many had no DNS answer in their last sweep.
func shardCounts(cov *store.Coverage) (unscanned, unknown int) {
	for i := 0; i < cov.Shards; i++ {
		if s := cov.Scanned(i); s != nil {
			unknown += s.Unknown
		} else {
			unscanned += shardSize(i, cov.Shards)
		}
	}
	return unscanned, unknown
}

// shardSize is the number of valid 4 character labels in shard of shards.
func shardSize(shard, shards int) int {
	n := 0
	names.Each(4, shard, shards, func(int, string) { n++ })
	return n
}

func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		return err
	}
	return store.WriteAtomic(path, append(b, '\n'))
}

func writeFour(path string, bm store.Bitmap, tld string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	names.Each(4, 0, 1, func(i int, s string) {
		if bm.Get(i) {
			w.WriteString(s + "." + tld + "\n")
		}
	})
	if err := w.Flush(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// diffChanges compares the previous available.json with the new one and
// prepends today's events to the rolling change log. A name that is newly
// available is "freed" only if the registry was seen to hold it before:
// it was dropping in the previous export, or freedSince reports an observed
// RDAP registered/dropping -> 404 transition after the previous export.
// Otherwise it is "newly_listed" (e.g. its first RDAP check). A name that
// fell back to unverified, stale or unknown is not reported as taken.
func diffChanges(prevPath, logPath string, cur *Available, keepDays int, freedSince func(domain string, since time.Time) bool) ([]Change, error) {
	log := []Change{} // never null: the site iterates over it
	if b, err := os.ReadFile(logPath); err == nil {
		if err := json.Unmarshal(b, &log); err != nil {
			return nil, err
		}
	}
	b, err := os.ReadFile(prevPath)
	if err != nil {
		return log, nil // first export: nothing to compare against
	}
	var prev Available
	if err := json.Unmarshal(b, &prev); err != nil {
		return nil, err
	}
	set := func(a *Available) map[string]bool {
		m := map[string]bool{}
		for tld, l := range a.TLDs {
			for _, n := range l.Available {
				m[n+"."+tld] = true
			}
		}
		return m
	}
	wasDropping := map[string]bool{}
	for tld, l := range prev.TLDs {
		for _, d := range l.Dropping {
			wasDropping[d.Name+"."+tld] = true
		}
	}
	before, after := set(&prev), set(cur)
	day := cur.Updated.Format("2006-01-02")
	var fresh []Change
	for d := range after {
		if before[d] || prev.TLDs[tldOf(d)] == nil {
			continue
		}
		ev := evNewlyListed
		if wasDropping[d] || (freedSince != nil && freedSince(d, prev.Updated)) {
			ev = evFreed
		}
		fresh = append(fresh, Change{day, d, ev})
	}
	for d := range before {
		l := cur.TLDs[tldOf(d)]
		if after[d] || l == nil || contains(l.Unverified, labelOf(d)) || contains(l.Unknown, labelOf(d)) {
			continue
		}
		fresh = append(fresh, Change{day, d, evTaken})
	}
	sort.Slice(fresh, func(i, j int) bool {
		if len(fresh[i].Domain) != len(fresh[j].Domain) {
			return len(fresh[i].Domain) < len(fresh[j].Domain)
		}
		return fresh[i].Domain < fresh[j].Domain
	})
	cutoff := cur.Updated.AddDate(0, 0, -keepDays).Format("2006-01-02")
	out := fresh
	for _, c := range log {
		if c.Date >= cutoff {
			out = append(out, c)
		}
	}
	if out == nil {
		out = []Change{}
	}
	return out, nil
}

func tldOf(d string) string   { return d[strings.LastIndexByte(d, '.')+1:] }
func labelOf(d string) string { return d[:strings.LastIndexByte(d, '.')] }

func contains(l []string, s string) bool {
	for _, x := range l {
		if x == s {
			return true
		}
	}
	return false
}

const (
	markStart = "<!-- shortfree:start -->"
	markEnd   = "<!-- shortfree:end -->"
)

func updateReadme(path string, a *Available, s *Stats, changes []Change) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	i, j := bytes.Index(b, []byte(markStart)), bytes.Index(b, []byte(markEnd))
	if i < 0 || j < i {
		return fmt.Errorf("%s: shortfree markers not found", path)
	}
	var w bytes.Buffer
	w.WriteString(markStart + "\n")
	fmt.Fprintf(&w, "\n_Last scan: **%s UTC**. Updated automatically by GitHub Actions._\n\n", a.Updated.Format("2006-01-02 15:04"))
	w.WriteString("| TLD | 1 char | 2 chars | 3 chars | 4 chars¹ | dropping² | list |\n")
	w.WriteString("|---|--:|--:|--:|--:|--:|---|\n")
	for _, tld := range s.Order {
		ls := s.TLDs[tld]
		fmt.Fprintf(&w, "| `.%s` | %s | %s | %s | %s | %d | [txt](data/txt/%s.txt) |\n", tld,
			cell(ls["1"]), cell(ls["2"]), cell(ls["3"]), cell4(ls["4"]), len(a.TLDs[tld].Dropping), tld)
	}
	w.WriteString("\n¹ not in the zone (DNS check only, not RDAP-verified). Full lists: [latest data release](../../releases/tag/data).  \n")
	w.WriteString("² registered, but in redemption or pending delete. The RDAP expiration date is not a drop date.  \n")
	w.WriteString("`?` = not delegated in DNS, but no current RDAP answer (check queued, or the last one is older than 14 days).  \n")
	w.WriteString("Counted names are not registered (RDAP 404). That does not prove a registrar will sell them: registries reserve or premium-price some names. See [accuracy](docs/accuracy.md).\n")
	if len(s.Missing) > 0 {
		fmt.Fprintf(&w, "\nNo data yet for .%s (first scan pending or failed).\n", strings.Join(s.Missing, " ."))
	}

	var freed []string
	for _, c := range changes {
		if c.Event == evFreed && len(freed) < 30 {
			freed = append(freed, "`"+c.Domain+"`")
		}
	}
	if len(freed) > 0 {
		w.WriteString("\n### Recently freed\n\nNames the registry reported as registered or dropping that are now not registered (RDAP 404).\n\n" + strings.Join(freed, " · ") + "\n")
	}

	w.WriteString("\n### Free 1–2 character domains\n\n")
	found := false
	for _, tld := range s.Order {
		var short []string
		for _, n := range a.TLDs[tld].Available {
			if len(n) <= 2 {
				short = append(short, "`"+n+"."+tld+"`")
			}
		}
		if len(short) == 0 {
			continue
		}
		found = true
		if len(short) > 60 {
			fmt.Fprintf(&w, "<details><summary><b>.%s</b> – %d names</summary>\n\n%s\n\n</details>\n\n", tld, len(short), strings.Join(short, " "))
		} else {
			fmt.Fprintf(&w, "**.%s** – %s\n\n", tld, strings.Join(short, " "))
		}
	}
	if !found {
		w.WriteString("None right now. Every 1–2 character name on these TLDs is registered or reserved.\n\n")
	}
	w.WriteString(markEnd)
	out := append([]byte{}, b[:i]...)
	out = append(out, w.Bytes()...)
	out = append(out, b[j+len(markEnd):]...)
	return store.WriteAtomic(path, out)
}

func cell(l *LenStats) string {
	if l.Available == 0 && l.Unverified > 0 {
		return fmt.Sprintf("%s?", num(l.Unverified))
	}
	if l.Unverified > 0 {
		return fmt.Sprintf("%s (+%s?)", num(l.Available), num(l.Unverified))
	}
	return num(l.Available)
}

// cell4 shows the 4 character count; "–" when there is no data, "partial"
// while some shards were never swept (their names are not counted).
func cell4(l *LenStats) string {
	switch {
	case l.Missing:
		return "–"
	case l.Unscanned > 0:
		return num(l.Free) + " (partial)"
	}
	return num(l.Free)
}

func num(n int) string {
	s := fmt.Sprint(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}
