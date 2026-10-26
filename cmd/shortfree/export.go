package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/maximilianfeix/shortfree/internal/names"
	"github.com/maximilianfeix/shortfree/internal/rdap"
	"github.com/maximilianfeix/shortfree/internal/store"
)

// Dropping is a registered name in redemption or pending delete.
type Dropping struct {
	Name    string    `json:"name"`
	Expires time.Time `json:"expires,omitempty"`
}

// TLDList is the public per-TLD result for 1-3 character names.
type TLDList struct {
	Available  []string   `json:"available"`
	Dropping   []Dropping `json:"dropping"`
	Unverified []string   `json:"unverified"`
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
	Free       int `json:"free,omitempty"` // 4 chars: not delegated (DNS only)
	Valid      int `json:"valid"`
}

// Stats is data/stats.json.
type Stats struct {
	Updated time.Time                       `json:"updated"`
	TLDs    map[string]map[string]*LenStats `json:"tlds"`
	Order   []string                        `json:"order"`
}

// Change is one entry of data/changes.json.
type Change struct {
	Date   string `json:"date"`
	Domain string `json:"domain"`
	Event  string `json:"event"` // "freed" or "taken"
}

func validCount(n int) int {
	c := 0
	names.Each(n, 0, 1, func(int, string) { c++ })
	return c
}

func cmdExport(args []string) error {
	fs := newFlags("export")
	tlds := fs.String("tlds", "", "comma-separated TLDs (default: all supported)")
	statePath := fs.String("state", "state/state.json.gz", "state file (.gz = compressed)")
	bitmaps := fs.String("bitmaps", "bitmaps", "directory with 4 character bitmaps")
	out := fs.String("out", "data", "output directory for public data")
	four := fs.String("four-out", "dist/4", "output directory for 4 character text lists (empty = skip)")
	readme := fs.String("readme", "README.md", "README to update between shortfree markers (empty = skip)")
	keepDays := fs.Int("changes-days", 30, "days of history kept in changes.json")
	fs.Parse(args)

	st, err := store.Load(*statePath)
	if err != nil {
		return err
	}
	order := parseTLDs(*tlds)
	avail := Available{Updated: st.Updated, TLDs: map[string]*TLDList{}}
	stats := Stats{Updated: st.Updated, TLDs: map[string]map[string]*LenStats{}, Order: order}
	valid := map[int]int{}
	for n := 1; n <= 4; n++ {
		valid[n] = validCount(n)
	}

	for _, tld := range order {
		list := &TLDList{Available: []string{}, Dropping: []Dropping{}, Unverified: []string{}}
		ls := map[string]*LenStats{}
		for n := 1; n <= 4; n++ {
			ls[fmt.Sprint(n)] = &LenStats{Valid: valid[n]}
		}
		for _, label := range st.Labels(tld) {
			e := st.TLDs[tld][label]
			s := ls[fmt.Sprint(len(label))]
			switch {
			case e.RDAP == string(rdap.Available):
				list.Available = append(list.Available, label)
				s.Available++
			case e.RDAP == string(rdap.Dropping):
				list.Dropping = append(list.Dropping, Dropping{label, e.ExpiresAt()})
				s.Dropping++
			case e.RDAP == "" || e.RDAP == string(rdap.Error):
				if e.DNS == "free" {
					list.Unverified = append(list.Unverified, label)
					s.Unverified++
				}
			}
		}
		avail.TLDs[tld] = list

		bm, err := store.LoadBitmap(filepath.Join(*bitmaps, tld+".bin"), names.Space(4))
		if err != nil {
			return err
		}
		ls["4"].Free = bm.Count()
		stats.TLDs[tld] = ls
		if *four != "" {
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
	changes, err := diffChanges(availPath, filepath.Join(*out, "changes.json"), &avail, *keepDays)
	if err != nil {
		return err
	}
	if err := writeJSON(availPath, avail); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(*out, "stats.json"), stats); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(*out, "changes.json"), changes); err != nil {
		return err
	}
	if *readme != "" {
		return updateReadme(*readme, &avail, &stats, changes)
	}
	return nil
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
// prepends today's freed/taken events to the rolling change log.
func diffChanges(prevPath, logPath string, cur *Available, keepDays int) ([]Change, error) {
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
	before, after := set(&prev), set(cur)
	day := cur.Updated.Format("2006-01-02")
	var fresh []Change
	for d := range after {
		if !before[d] && prev.TLDs[tldOf(d)] != nil {
			fresh = append(fresh, Change{day, d, "freed"})
		}
	}
	for d := range before {
		// A name that only fell back to "unverified" is not a registration.
		if l := cur.TLDs[tldOf(d)]; !after[d] && l != nil && !contains(l.Unverified, labelOf(d)) {
			fresh = append(fresh, Change{day, d, "taken"})
		}
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
			cell(ls["1"]), cell(ls["2"]), cell(ls["3"]), num(ls["4"].Free), len(a.TLDs[tld].Dropping), tld)
	}
	w.WriteString("\n¹ not in the zone (DNS check only, not RDAP-verified). Full lists: [latest data release](../../releases/tag/data).  \n")
	w.WriteString("² registered, but in redemption or pending delete – may become available soon.  \n")
	w.WriteString("`?` = not delegated in DNS, RDAP verification still queued.\n")

	var freed []string
	for _, c := range changes {
		if c.Event == "freed" && len(freed) < 30 {
			freed = append(freed, "`"+c.Domain+"`")
		}
	}
	if len(freed) > 0 {
		w.WriteString("\n### Recently freed\n\n" + strings.Join(freed, " · ") + "\n")
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

func num(n int) string {
	s := fmt.Sprint(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}
