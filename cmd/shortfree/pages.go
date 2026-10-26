package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/maximilianfeix/shortfree/internal/names"
	"github.com/maximilianfeix/shortfree/internal/store"
)

// siteBase is the public URL of the GitHub Pages site.
const siteBase = "https://maximilianfeix.github.io/shortfree/"

const (
	repoURL      = "https://github.com/maximilianfeix/shortfree"
	releaseURL   = repoURL + "/releases/tag/data"
	downloadURL  = repoURL + "/releases/download/data/"
	liveCheckURL = repoURL + "#check-a-domain-yourself"
	cc0URL       = "https://creativecommons.org/publicdomain/zero/1.0/"
	pageSampleN  = 500 // pronounceable 4 letter names shown on a 4-letter page
)

// reservedDirs are top-level paths of the site that a TLD page must not use.
var reservedDirs = map[string]bool{"api": true, "data": true, "page": true}

var tldPattern = regexp.MustCompile(`^[a-z0-9-]{2,63}$`)

// cmdPages writes static, crawlable pages per TLD and length plus
// sitemap.xml and robots.txt into the assembled site directory.
func cmdPages(args []string) error {
	fs := newFlags("pages")
	data := fs.String("data", "data", "directory with stats.json and available.json")
	bitmaps := fs.String("bitmaps", "bitmaps", "directory with 4 character bitmaps (missing = no 4 letter samples)")
	out := fs.String("out", "_site", "site directory to write into")
	base := fs.String("base", siteBase, "public base URL of the site, with trailing slash")
	perPage := fs.Int("per-page", 1000, "domains per list page")
	fs.Parse(args)

	var stats Stats
	var avail Available
	if err := readJSON(filepath.Join(*data, "stats.json"), &stats); err != nil {
		return err
	}
	if err := readJSON(filepath.Join(*data, "available.json"), &avail); err != nil {
		return err
	}
	bms := map[string]store.Bitmap{}
	for _, tld := range stats.Order {
		bm, err := store.LoadBitmap(filepath.Join(*bitmaps, tld+".bin"), names.Space(4))
		if err != nil {
			return err
		}
		bms[tld] = bm
	}
	files, err := buildPages(pageInput{Stats: &stats, Avail: &avail, Bitmaps: bms, Base: *base, PerPage: *perPage})
	if err != nil {
		return err
	}
	for path, b := range files {
		if err := store.WriteAtomic(filepath.Join(*out, filepath.FromSlash(path)), b); err != nil {
			return err
		}
	}
	fmt.Fprintf(os.Stderr, "pages: wrote %d files to %s\n", len(files), *out)
	return nil
}

func readJSON(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, v); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}

// pageInput is everything the generator needs; tests build it directly.
type pageInput struct {
	Stats   *Stats
	Avail   *Available
	Bitmaps map[string]store.Bitmap // tld -> 4 character free bitmap; nil = unknown
	Base    string
	PerPage int
}

type crumb struct {
	Name, Href, URL string
}

type domainGroup struct {
	Title string
	Note  string
	Names []string
}

type lenRow struct {
	N          int
	Href       string
	Free       int
	Unverified int
	Valid      int
	FourOnly   bool
}

type letterCount struct {
	Letter string
	Count  int
}

type pagerLink struct {
	Label   string
	Href    string
	Current bool
	Gap     bool
}

// pageData feeds the page template. Kind is "tld", "len" or "four".
type pageData struct {
	Kind        string
	Root        string // relative path from the page to the site root
	URL         string // canonical URL
	Title       string
	Description string
	H1          string
	Lede        string
	TLD         string
	N           int
	Updated     time.Time
	Noindex     bool
	Crumbs      []crumb
	JSONLD      template.JS
	Image       string

	// facts
	Count      int
	Unverified int
	Dropping   int
	Four       int
	Letters4   int

	// lists
	Groups  []domainGroup
	Pager   []pagerLink
	PageNo  int
	Pages   int
	Lengths []lenRow
	Short   []string
	Drops   []Dropping
	First   []letterCount
	Sample  []string
	Others  []string

	// downloads
	TXT, JSON, CSV, FourTXT string
}

func (p *pageData) Scan() string { return p.Updated.UTC().Format("2006-01-02 15:04") + " UTC" }
func (p *pageData) Day() string  { return p.Updated.UTC().Format("2006-01-02") }

// buildPages renders every page and returns site-relative path -> content.
func buildPages(in pageInput) (map[string][]byte, error) {
	if in.Stats == nil || in.Avail == nil {
		return nil, errors.New("pages: missing stats or available data")
	}
	if in.PerPage <= 0 {
		in.PerPage = 1000
	}
	if in.Base == "" {
		in.Base = siteBase
	}
	if !strings.HasSuffix(in.Base, "/") {
		in.Base += "/"
	}
	files := map[string][]byte{}
	var sitemap []string
	sitemap = append(sitemap, in.Base)
	updated := in.Stats.Updated
	if updated.IsZero() {
		updated = in.Avail.Updated
	}

	var tlds []string
	for _, tld := range in.Stats.Order {
		if !tldPattern.MatchString(tld) || reservedDirs[tld] || in.Avail.TLDs[tld] == nil || in.Stats.TLDs[tld] == nil {
			continue
		}
		tlds = append(tlds, tld)
	}

	emit := func(path string, p *pageData) error {
		var buf bytes.Buffer
		if err := pageTmpl.Execute(&buf, p); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		files[path] = buf.Bytes()
		if !p.Noindex {
			sitemap = append(sitemap, p.URL)
		}
		return nil
	}

	for _, tld := range tlds {
		list := in.Avail.TLDs[tld]
		ls := in.Stats.TLDs[tld]
		others := make([]string, 0, len(tlds)-1)
		for _, t := range tlds {
			if t != tld {
				others = append(others, t)
			}
		}
		tldCrumb := crumb{"." + tld, "", in.Base + tld + "/"}
		home := crumb{"shortfree", "", in.Base}
		dl := func(p *pageData) {
			p.TXT = p.Root + "data/txt/" + tld + ".txt"
			p.CSV = p.Root + "api/v1/" + tld + ".csv"
			p.FourTXT = downloadURL + "four-" + tld + ".txt.gz"
		}

		// /<tld>/
		bm := in.Bitmaps[tld]
		free4 := statInt(ls, 4, func(s *LenStats) int { return s.Free })
		letters4 := 0
		if bm != nil {
			letters4 = countLetters4(bm, nil)
		}
		verified := 0
		for n := 1; n <= 3; n++ {
			verified += statInt(ls, n, func(s *LenStats) int { return s.Available })
		}
		p := &pageData{
			Kind: "tld", Root: "../", URL: tldCrumb.URL, TLD: tld, Updated: updated, Others: others,
			Title:       fmt.Sprintf("Free short .%s domains – %s verified 1–3 character names | shortfree", tld, num(verified)),
			Description: fmt.Sprintf("Every unregistered 1–4 character .%s domain: %s RDAP-verified 1–3 character names and %s 4 character names not in the zone. Rescanned daily, open data.", tld, num(verified), num(free4)),
			H1:          fmt.Sprintf("Free short .%s domains", tld),
			Lede:        fmt.Sprintf("Every 1–4 character .%s name that was unregistered at the last scan, from single characters to four. Lists are rebuilt from the registry every day.", tld),
			Count:       verified, Dropping: len(list.Dropping), Four: free4, Letters4: letters4,
			Crumbs: []crumb{{home.Name, "../", home.URL}, {tldCrumb.Name, "", tldCrumb.URL}},
			Drops:  list.Dropping,
		}
		dl(p)
		p.JSON = p.Root + "api/v1/" + tld + ".json"
		for n := 1; n <= 4; n++ {
			s := ls[fmt.Sprint(n)]
			if s == nil {
				s = &LenStats{}
			}
			r := lenRow{N: n, Href: fmt.Sprintf("%d-letter/", n), Free: s.Available, Unverified: s.Unverified, Valid: s.Valid}
			if n == 4 {
				r.Free, r.Unverified, r.FourOnly = s.Free, 0, true
			}
			p.Lengths = append(p.Lengths, r)
		}
		for _, l := range list.Available {
			if len(l) <= 2 && names.Kind(l) == "letters" && len(p.Short) < 120 {
				p.Short = append(p.Short, l+"."+tld)
			}
		}
		p.JSONLD = jsonLD(datasetLD(p, fmt.Sprintf("Free short .%s domains", tld), p.Description, []dist{
			{"text/plain", in.Base + "data/txt/" + tld + ".txt"},
			{"application/json", in.Base + "api/v1/" + tld + ".json"},
			{"text/csv", in.Base + "api/v1/" + tld + ".csv"},
		}), breadcrumbLD(p.Crumbs))
		if err := emit(tld+"/index.html", p); err != nil {
			return nil, err
		}

		// /<tld>/<n>-letter/ (n = 1..3), paginated
		for n := 1; n <= 3; n++ {
			var letters, other []string
			for _, l := range list.Available {
				if len(l) != n {
					continue
				}
				if names.Kind(l) == "letters" {
					letters = append(letters, l)
				} else {
					other = append(other, l)
				}
			}
			sort.Strings(letters)
			sort.Strings(other)
			unv := 0
			for _, l := range list.Unverified {
				if len(l) == n {
					unv++
				}
			}
			all := append(append([]string{}, letters...), other...)
			pages := max(1, (len(all)+in.PerPage-1)/in.PerPage)
			dir := fmt.Sprintf("%s/%d-letter/", tld, n)
			for pg := 1; pg <= pages; pg++ {
				path, root := dir, "../../"
				if pg > 1 {
					path, root = fmt.Sprintf("%spage/%d/", dir, pg), "../../../../"
				}
				lo, hi := (pg-1)*in.PerPage, min(pg*in.PerPage, len(all))
				p := &pageData{
					Kind: "len", Root: root, URL: in.Base + path, TLD: tld, N: n, Updated: updated, Others: others,
					Count: len(all), Unverified: unv, PageNo: pg, Pages: pages, Noindex: len(all) == 0,
					H1:     fmt.Sprintf("Free %d-letter .%s domains", n, tld),
					Crumbs: []crumb{{home.Name, root, home.URL}, {tldCrumb.Name, root + tld + "/", tldCrumb.URL}, {fmt.Sprintf("%d-letter", n), "", in.Base + dir}},
				}
				dl(p)
				p.JSON = root + fmt.Sprintf("api/v1/%s/%d.json", tld, n)
				if len(all) == 0 {
					p.Title = fmt.Sprintf("Free %d-letter .%s domains – none right now | shortfree", n, tld)
					p.Description = fmt.Sprintf("No %d character .%s domain was unregistered at the last scan on %s. The list is rechecked daily.", n, tld, p.Day())
					p.Lede = fmt.Sprintf("All %s valid %d character .%s names were registered or reserved at the last scan.", num(statInt(ls, n, func(s *LenStats) int { return s.Valid })), n, tld)
				} else {
					p.Title = fmt.Sprintf("Free %d-letter .%s domains – %s available | shortfree", n, tld, num(len(all)))
					p.Description = fmt.Sprintf("%s unregistered %d character .%s domains, confirmed via RDAP on %s. Full list of letters, digits and hyphen names, rescanned daily.", num(len(all)), n, tld, p.Day())
					p.Lede = fmt.Sprintf("%s .%s names with %d characters (letters, digits or an inner hyphen) that the registry reported as not registered at the last scan.", num(len(all)), tld, n)
				}
				if pg > 1 {
					p.Title = fmt.Sprintf("Free %d-letter .%s domains – page %d of %d | shortfree", n, tld, pg, pages)
					p.H1 += fmt.Sprintf(", page %d", pg)
				}
				p.Groups = groupSlice(letters, other, lo, hi, tld)
				p.Pager = pager(pg, pages, func(k int) string {
					if k == 1 {
						return root + dir
					}
					return root + fmt.Sprintf("%spage/%d/", dir, k)
				})
				if pg > 1 {
					p.Crumbs[2].Href = root + dir
					p.Crumbs = append(p.Crumbs, crumb{fmt.Sprintf("page %d", pg), "", p.URL})
				}
				p.JSONLD = jsonLD(datasetLD(p, fmt.Sprintf("Free %d character .%s domains", n, tld), p.Description, []dist{
					{"application/json", in.Base + fmt.Sprintf("api/v1/%s/%d.json", tld, n)},
					{"text/csv", in.Base + "api/v1/" + tld + ".csv"},
				}), breadcrumbLD(p.Crumbs))
				if err := emit(path+"index.html", p); err != nil {
					return nil, err
				}
			}
		}

		// /<tld>/4-letter/: letters-only summary, links to the full list
		root := "../../"
		p4 := &pageData{
			Kind: "four", Root: root, URL: in.Base + tld + "/4-letter/", TLD: tld, N: 4, Updated: updated, Others: others,
			Four: free4, Letters4: letters4, Noindex: letters4 == 0,
			H1:     fmt.Sprintf("Free 4-letter .%s domains", tld),
			Crumbs: []crumb{{home.Name, root, home.URL}, {tldCrumb.Name, root + tld + "/", tldCrumb.URL}, {"4-letter", "", in.Base + tld + "/4-letter/"}},
		}
		dl(p4)
		p4.Title = fmt.Sprintf("Free 4-letter .%s domains – %s letters-only names | shortfree", tld, num(letters4))
		p4.Description = fmt.Sprintf("%s four letter .%s names (a–z only) are not in the .%s zone, %s in total with digits and hyphens. Per-letter counts, pronounceable picks and the full list.", num(letters4), tld, tld, num(free4))
		p4.Lede = fmt.Sprintf("%s of the 456,976 possible a–z four letter .%s names were not in the zone at the last scan. This page summarises them; the complete list is a download.", num(letters4), tld)
		if bm != nil {
			p4.First = make([]letterCount, 26)
			for i := range p4.First {
				p4.First[i].Letter = string(rune('a' + i))
			}
			countLetters4(bm, func(s string) { p4.First[s[0]-'a'].Count++ })
			p4.Sample = pronounceable(bm, tld, pageSampleN)
		}
		p4.JSONLD = jsonLD(datasetLD(p4, fmt.Sprintf("Free 4 character .%s domains", tld), p4.Description, []dist{
			{"application/gzip", p4.FourTXT},
		}), breadcrumbLD(p4.Crumbs))
		if err := emit(tld+"/4-letter/index.html", p4); err != nil {
			return nil, err
		}
	}

	files["sitemap.xml"] = sitemapXML(sitemap, updated)
	files["robots.txt"] = []byte("User-agent: *\nAllow: /\n\nSitemap: " + in.Base + "sitemap.xml\n")
	return files, nil
}

func statInt(ls map[string]*LenStats, n int, f func(*LenStats) int) int {
	if s := ls[fmt.Sprint(n)]; s != nil {
		return f(s)
	}
	return 0
}

// groupSlice cuts the window [lo, hi) out of letters followed by other and
// returns it as titled groups.
func groupSlice(letters, other []string, lo, hi int, tld string) []domainGroup {
	var g []domainGroup
	add := func(title, note string, l []string, from, to int) {
		from, to = max(from, 0), min(to, len(l))
		if from >= to {
			return
		}
		d := make([]string, 0, to-from)
		for _, s := range l[from:to] {
			d = append(d, s+"."+tld)
		}
		if from > 0 {
			title += " (continued)"
		}
		g = append(g, domainGroup{title, note, d})
	}
	add(fmt.Sprintf("Letters only · %s", num(len(letters))), "a–z, sorted alphabetically.", letters, lo, hi)
	add(fmt.Sprintf("With digits or a hyphen · %s", num(len(other))), "Contain at least one digit or an inner hyphen.", other, lo-len(letters), hi-len(letters))
	return g
}

// pager returns page links: all pages when there are few, otherwise the
// first, the last and two on each side of the current one.
func pager(cur, total int, href func(int) string) []pagerLink {
	if total <= 1 {
		return nil
	}
	var out []pagerLink
	gap := false
	for k := 1; k <= total; k++ {
		if total > 9 && k != 1 && k != total && (k < cur-2 || k > cur+2) {
			if !gap {
				out = append(out, pagerLink{Gap: true, Label: "…"})
				gap = true
			}
			continue
		}
		gap = false
		out = append(out, pagerLink{Label: fmt.Sprint(k), Href: href(k), Current: k == cur})
	}
	return out
}

// countLetters4 counts free 4 character a-z names in bm, calling fn for each.
func countLetters4(bm store.Bitmap, fn func(string)) int {
	n := 0
	var b [4]byte
	for i0 := 0; i0 < 26; i0++ {
		for i1 := 0; i1 < 26; i1++ {
			for i2 := 0; i2 < 26; i2++ {
				for i3 := 0; i3 < 26; i3++ {
					if !bm.Get(((i0*names.Base+i1)*names.Base+i2)*names.Base + i3) {
						continue
					}
					n++
					if fn != nil {
						b = [4]byte{byte('a' + i0), byte('a' + i1), byte('a' + i2), byte('a' + i3)}
						fn(string(b[:]))
					}
				}
			}
		}
	}
	return n
}

// pronounceable returns up to limit free consonant-vowel-consonant-vowel
// names (like "toka"), alphabetically.
func pronounceable(bm store.Bitmap, tld string, limit int) []string {
	const cons, vow = "bcdfghjklmnprstvwz", "aeiou"
	var out []string
	for _, a := range cons {
		for _, b := range vow {
			for _, c := range cons {
				for _, d := range vow {
					s := string([]rune{a, b, c, d})
					if bm.Get(names.Index(s)) {
						out = append(out, s+"."+tld)
						if len(out) == limit {
							return out
						}
					}
				}
			}
		}
	}
	return out
}

type dist struct{ Format, URL string }

func datasetLD(p *pageData, name, desc string, ds []dist) map[string]any {
	var d []map[string]any
	for _, x := range ds {
		d = append(d, map[string]any{"@type": "DataDownload", "encodingFormat": x.Format, "contentUrl": x.URL})
	}
	return map[string]any{
		"@context":            "https://schema.org",
		"@type":               "Dataset",
		"name":                name,
		"description":         desc,
		"url":                 p.URL,
		"dateModified":        p.Updated.UTC().Format(time.RFC3339),
		"license":             cc0URL,
		"isAccessibleForFree": true,
		"keywords":            []string{"domain names", "short domains", "." + p.TLD, "domain availability", "RDAP"},
		"creator":             map[string]any{"@type": "Organization", "name": "shortfree", "url": repoURL},
		"distribution":        d,
	}
}

func breadcrumbLD(cs []crumb) map[string]any {
	var items []map[string]any
	for i, c := range cs {
		items = append(items, map[string]any{"@type": "ListItem", "position": i + 1, "name": c.Name, "item": c.URL})
	}
	return map[string]any{"@context": "https://schema.org", "@type": "BreadcrumbList", "itemListElement": items}
}

// jsonLD marshals the objects for a <script type="application/ld+json">.
// encoding/json escapes <, > and &, so the result cannot close the script.
func jsonLD(objs ...map[string]any) template.JS {
	b, err := json.Marshal(objs)
	if err != nil {
		panic(err) // only maps of strings, numbers and slices
	}
	return template.JS(b)
}

func sitemapXML(urls []string, updated time.Time) []byte {
	var b bytes.Buffer
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
	day := updated.UTC().Format("2006-01-02")
	for _, u := range urls {
		var esc bytes.Buffer
		template.HTMLEscape(&esc, []byte(u))
		fmt.Fprintf(&b, "  <url><loc>%s</loc><lastmod>%s</lastmod></url>\n", esc.String(), day)
	}
	b.WriteString("</urlset>\n")
	return b.Bytes()
}

var pageTmpl = template.Must(template.New("page").Funcs(template.FuncMap{
	"num":     num,
	"lockup":  func() template.HTML { return lockupSVG },
	"favicon": func() template.HTML { return faviconSVG },
	"last":    func(i int, l []crumb) bool { return i == len(l)-1 },
	"pct": func(a, b int) string {
		if b == 0 {
			return "0"
		}
		return fmt.Sprintf("%.1f", 100*float64(a)/float64(b))
	},
}).Parse(pageHTML))

// lockupSVG is brand/lockup-*.svg inline: the wordmark takes the text colour
// and the dark-background edge comes from CSS, so one copy serves both themes.
const lockupSVG template.HTML = `<svg class="lockup" viewBox="0 0 480 96" width="200" height="40" aria-hidden="true" focusable="false"><g transform="translate(8 8) scale(0.15625)"><rect width="512" height="512" rx="112" fill="#0B0C0E"/><circle cx="160" cy="256" r="32" fill="#1FC16B"/><path d="M256 256H416V288H256Z" fill="#FAFAF7"/></g><rect class="lk-edge" x="8.5" y="8.5" width="79" height="79" rx="17" fill="none" stroke-width="1"/><g class="lk-word" transform="translate(112 16)" fill-rule="evenodd"><path d="M8 16H32V24H12Q8 24 8 28V32Q8 36 12 36H24Q32 36 32 44V56Q32 64 24 64H0V56H20Q24 56 24 52V48Q24 44 20 44H8Q0 44 0 36V24Q0 16 8 16Z"/><path transform="translate(40 0)" d="M0 0H8V16H24Q32 16 32 24V64H24V28Q24 24 20 24H8V64H0Z"/><path transform="translate(80 0)" d="M10 16H22Q32 16 32 26V54Q32 64 22 64H10Q0 64 0 54V26Q0 16 10 16ZM12 24Q8 24 8 28V52Q8 56 12 56H20Q24 56 24 52V28Q24 24 20 24Z"/><path transform="translate(120 0)" d="M0 16H8V20Q12 16 20 16H32V24H20Q8 24 8 36V64H0Z"/><path transform="translate(160 0)" d="M12 4H20V16H32V24H20V52Q20 56 24 56H32V64H24Q12 64 12 52V24H0V16H12Z"/><path transform="translate(200 0)" d="M8 12Q8 0 20 0H32V8H20Q16 8 16 12V16H32V24H16V64H8V24H0V16H8Z"/><path transform="translate(240 0)" d="M0 16H8V20Q12 16 20 16H32V24H20Q8 24 8 36V64H0Z"/><path transform="translate(280 0)" d="M8 16H24Q32 16 32 24V44H8V52Q8 56 12 56H32V64H8Q0 64 0 56V24Q0 16 8 16ZM8 28V36H24V28Q24 24 20 24H12Q8 24 8 28Z"/><path transform="translate(320 0)" d="M8 16H24Q32 16 32 24V44H8V52Q8 56 12 56H32V64H8Q0 64 0 56V24Q0 16 8 16ZM8 28V36H24V28Q24 24 20 24H12Q8 24 8 28Z"/></g></svg>`

const faviconSVG template.HTML = `<svg class="foot-mark" viewBox="0 0 32 32" width="24" height="24" aria-hidden="true" focusable="false"><rect width="32" height="32" rx="7" fill="#0B0C0E"/><circle cx="10" cy="16" r="2" fill="#1FC16B"/><path d="M16 16H26V18H16Z" fill="#FAFAF7"/></svg>`

const pageHTML = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Title}}</title>
<meta name="description" content="{{.Description}}">
<link rel="canonical" href="{{.URL}}">
{{- if .Noindex}}
<meta name="robots" content="noindex, follow">
{{- end}}
<meta name="theme-color" content="#0B0C0E" media="(prefers-color-scheme: dark)">
<meta name="theme-color" content="#FAFAF7" media="(prefers-color-scheme: light)">
<meta name="color-scheme" content="dark light">
<meta property="og:site_name" content="shortfree">
<meta property="og:type" content="website">
<meta property="og:title" content="{{.H1}}">
<meta property="og:description" content="{{.Description}}">
<meta property="og:url" content="{{.URL}}">
<meta property="og:image" content="https://maximilianfeix.github.io/shortfree/social-preview.png">
<meta property="og:image:width" content="1280">
<meta property="og:image:height" content="640">
<meta name="twitter:card" content="summary_large_image">
<meta name="twitter:title" content="{{.H1}}">
<meta name="twitter:description" content="{{.Description}}">
<meta name="twitter:image" content="https://maximilianfeix.github.io/shortfree/social-preview.png">
<link rel="icon" href="{{.Root}}favicon.svg" type="image/svg+xml">
<link rel="stylesheet" href="{{.Root}}style.css">
<script type="application/ld+json">{{.JSONLD}}</script>
</head>
<body>
<a class="skip" href="#main">Skip to content</a>
<header class="top">
  <div class="wrap top-in">
    <a class="brand" href="{{.Root}}" aria-label="shortfree home">{{lockup}}</a>
    <nav class="nav" aria-label="Primary">
      <a class="nav-opt" href="{{.Root}}#browse">Browse</a>
      <a class="nav-opt" href="{{.Root}}#use">API &amp; MCP</a>
      <a href="https://github.com/maximilianfeix/shortfree">GitHub</a>
    </nav>
  </div>
</header>
<main id="main">
  <div class="wrap">
    <nav class="crumbs" aria-label="Breadcrumb"><ol>
      {{- range $i, $c := .Crumbs}}
      <li>{{if last $i $.Crumbs}}<span aria-current="page">{{$c.Name}}</span>{{else}}<a href="{{$c.Href}}">{{$c.Name}}</a>{{end}}</li>
      {{- end}}
    </ol></nav>
    <div class="page-head">
      <h1>{{.H1}}</h1>
      <p class="lede">{{.Lede}}</p>
    </div>
    <dl class="page-facts">
      {{- if eq .Kind "tld"}}
      <div><dt>Verified free, 1–3 chars</dt><dd>{{num .Count}}</dd></div>
      <div><dt>4 chars, not in zone</dt><dd>{{num .Four}}</dd></div>
      <div><dt>Dropping</dt><dd>{{num .Dropping}}</dd></div>
      {{- else if eq .Kind "len"}}
      <div><dt>Verified free</dt><dd>{{num .Count}}</dd></div>
      {{- if .Unverified}}<div><dt>Awaiting RDAP</dt><dd>{{num .Unverified}}</dd></div>{{end}}
      {{- else}}
      <div><dt>Letters only, a–z</dt><dd>{{num .Letters4}}</dd></div>
      <div><dt>All 4 character names</dt><dd>{{num .Four}}</dd></div>
      {{- end}}
      <div><dt>Last scan</dt><dd><time datetime="{{.Updated.UTC.Format "2006-01-02T15:04:05Z07:00"}}">{{.Scan}}</time></dd></div>
    </dl>

    {{- if eq .Kind "tld"}}
    <section class="section" aria-labelledby="len-h">
      <div class="section-head"><h2 id="len-h">By length</h2><p class="note">1–3 characters are RDAP-verified; 4 characters are a DNS check.</p></div>
      <div class="lens-wrap"><table class="lens">
        <thead><tr><th scope="col">Length</th><th scope="col">Free</th><th scope="col">Awaiting RDAP</th><th scope="col">Of valid names</th></tr></thead>
        <tbody>
        {{- range .Lengths}}
          <tr><td><a href="{{.Href}}">{{.N}}-letter .{{$.TLD}}</a></td><td>{{num .Free}}</td><td>{{if .FourOnly}}–{{else}}{{num .Unverified}}{{end}}</td><td>{{pct .Free .Valid}}% of {{num .Valid}}</td></tr>
        {{- end}}
        </tbody>
      </table></div>
    </section>
    {{- if .Short}}
    <section class="section" aria-labelledby="short-h">
      <h2 id="short-h">Free 1–2 letter .{{.TLD}} names</h2>
      <p class="note">Letters only. Registries often reserve these or charge premium prices; check at a registrar before you plan around one.</p>
      <ul class="domain-grid" style="margin-top:16px">{{range .Short}}<li><span>{{.}}</span></li>{{end}}</ul>
    </section>
    {{- end}}
    <section class="section" aria-labelledby="drop-h">
      <h2 id="drop-h">Dropping soon</h2>
      <p class="note">Registered, but in redemption or pending delete at the last scan.</p>
      {{- if .Drops}}
      <ul class="feed">{{range .Drops}}<li><span>{{.Name}}.{{$.TLD}}</span><span class="when"><span class="status" data-state="dropping">dropping</span>{{if not .Expires.IsZero}} exp {{.Expires.UTC.Format "2006-01-02"}}{{end}}</span></li>{{end}}</ul>
      {{- else}}
      <p class="msg">No short .{{.TLD}} names in redemption right now.</p>
      {{- end}}
    </section>
    {{- end}}

    {{- if eq .Kind "len"}}
    <section class="section" aria-labelledby="list-h">
      <div class="section-head">
        <h2 id="list-h">{{if .Count}}{{num .Count}} free {{.N}}-letter .{{.TLD}} domains{{else}}No free {{.N}}-letter .{{.TLD}} domains{{end}}</h2>
        {{- if gt .Pages 1}}<p class="note">Page {{.PageNo}} of {{.Pages}}</p>{{end}}
      </div>
      {{- if .Count}}
      {{- range .Groups}}
      <h3 class="sub">{{.Title}}</h3>
      <p class="note">{{.Note}}</p>
      <ul class="domain-grid">{{range .Names}}<li><span>{{.}}</span></li>{{end}}</ul>
      {{- end}}
      {{- if .Pager}}
      <nav class="pager" aria-label="Pages">{{range .Pager}}{{if .Gap}}<span class="gap">…</span>{{else if .Current}}<span aria-current="page">{{.Label}}</span>{{else}}<a href="{{.Href}}">{{.Label}}</a>{{end}}{{end}}</nav>
      {{- end}}
      {{- else}}
      <p class="msg">Every valid {{.N}} character .{{.TLD}} name was registered or reserved at the last scan. Names do come back: see what was <a href="{{.Root}}#freed-h">recently freed</a> or <a href="{{.Root}}{{.TLD}}/">dropping soon on .{{.TLD}}</a>.</p>
      {{- end}}
      {{- if .Unverified}}
      <p class="note" style="margin-top:16px">{{num .Unverified}} more {{.N}} character names are not in DNS but still wait for the RDAP check. They show up in the <a href="{{.Root}}#browse">interactive browser</a> marked with “?”.</p>
      {{- end}}
    </section>
    {{- end}}

    {{- if eq .Kind "four"}}
    {{- if .First}}
    <section class="section" aria-labelledby="first-h">
      <h2 id="first-h">Letters-only names by first letter</h2>
      <p class="note" style="margin:4px 0 16px">Free a–z four letter .{{.TLD}} names, counted by their first letter.</p>
      <ul class="letters">{{range .First}}<li><span>{{.Letter}}…</span><span>{{num .Count}}</span></li>{{end}}</ul>
    </section>
    {{- end}}
    {{- if .Sample}}
    <section class="section" aria-labelledby="sample-h">
      <h2 id="sample-h">Pronounceable picks</h2>
      <p class="note" style="margin:4px 0 16px">Consonant–vowel–consonant–vowel names that are free, alphabetically{{if ge (len .Sample) 500}}, first 500{{end}}.</p>
      <ul class="domain-grid">{{range .Sample}}<li><span>{{.}}</span></li>{{end}}</ul>
    </section>
    {{- end}}
    {{- end}}

    <div class="cta">
      {{- if eq .Kind "four"}}
      <a class="btn" href="{{.FourTXT}}">Full 4 character list (.txt.gz)</a>
      {{- else}}
      <a class="btn" href="{{.JSON}}">JSON</a>
      <a class="btn" href="{{.CSV}}">CSV</a>
      <a class="btn" href="{{.TXT}}">Plain text</a>
      {{- end}}
      <a class="btn btn-quiet" href="{{.Root}}#browse">Filter in the browser</a>
    </div>

    <section class="section method" aria-labelledby="method-h">
      <h2 id="method-h">How this list is made</h2>
      {{- if eq .Kind "four"}}
      <p>Each of the 1.8 million possible four character .{{.TLD}} labels is asked for at the TLD's authoritative name servers, in seven daily shards. A name is listed when the answer is <code>NXDOMAIN</code>: it is not in the zone. That is a DNS check only; a registered name without name servers would also show up, so confirm at a registrar.</p>
      {{- else}}
      <p>Every valid label is asked for at the .{{.TLD}} name servers. Names that are not in the zone are then looked up in the registry's RDAP service, and only names it reports as not registered (<code>404</code>) are listed as free. Names in redemption or pending delete are listed as dropping.</p>
      {{- end}}
      <p>“Free” means not registered at the last scan on {{.Scan}}. Registries may still reserve a name or price it as premium, and names change state between scans. For a live answer run <a href="` + liveCheckURL + `"><code>shortfree check</code></a> or use the <a href="https://github.com/maximilianfeix/shortfree/blob/main/docs/mcp.md">MCP server</a>.</p>
    </section>

    {{- if .Others}}
    <section class="section" aria-labelledby="other-h">
      <h2 id="other-h">Other TLDs</h2>
      <ul class="tld-index" style="margin-top:16px">{{range .Others}}<li><a href="{{$.Root}}{{.}}/">.{{.}}</a>{{if $.N}}<a href="{{$.Root}}{{.}}/{{$.N}}-letter/" aria-label="{{$.N}}-letter .{{.}}">{{$.N}}-letter</a>{{end}}</li>{{end}}</ul>
    </section>
    {{- end}}
  </div>
</main>
<footer class="foot">
  <div class="wrap foot-in">
    {{favicon}}
    <p>Code MIT · data CC0 · no cookies, no analytics</p>
    <nav aria-label="Footer">
      <a href="https://github.com/maximilianfeix/shortfree">Source</a>
      <a href="{{.Root}}api/v1/index.json">API</a>
      <a href="https://github.com/maximilianfeix/shortfree/blob/main/docs/mcp.md">MCP</a>
      <a href="` + releaseURL + `">All lists</a>
      <a href="{{.Root}}sitemap.xml">Sitemap</a>
    </nav>
  </div>
</footer>
</body>
</html>
`
