package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/maximilianfeix/shortfree/internal/names"
	"github.com/maximilianfeix/shortfree/internal/store"
)

// pagesFixture has .dev with a few names of every length, .com with nothing
// free and a TLD that would collide with the site's api/ directory.
func pagesFixture() pageInput {
	updated := time.Date(2026, 10, 26, 6, 30, 0, 0, time.UTC)
	ls := func(a1, a2, a3, u3, f4 int) map[string]*LenStats {
		return map[string]*LenStats{
			"1": {Available: a1, Valid: 36},
			"2": {Available: a2, Valid: 1296},
			"3": {Available: a3, Unverified: u3, Valid: 47952},
			"4": {Free: f4, Valid: 1772928},
		}
	}
	bm := store.NewBitmap(names.Space(4))
	for _, s := range []string{"toka", "zzzz", "ab1c", "bela"} {
		bm.Set(names.Index(s), true)
	}
	// 2,500 free 3 character .dev names so the list paginates at 1,000.
	var three []string
	names.Each(3, 0, 1, func(_ int, s string) {
		if len(three) < 2500 {
			three = append(three, s)
		}
	})
	dev := &TLDList{
		Available:  append([]string{"x", "q7", "ab"}, three...),
		Dropping:   []Dropping{{Name: "zz", Expires: updated.AddDate(0, 0, 10)}},
		Unverified: []string{"k9x"},
	}
	return pageInput{
		Stats: &Stats{Updated: updated, Order: []string{"dev", "com", "api"}, TLDs: map[string]map[string]*LenStats{
			"dev": ls(1, 2, 2500, 1, 4), "com": ls(0, 0, 0, 0, 0), "api": ls(0, 0, 0, 0, 0),
		}},
		Avail: &Available{Updated: updated, TLDs: map[string]*TLDList{
			"dev": dev,
			"com": {Available: []string{}, Dropping: []Dropping{}, Unverified: []string{}},
			"api": {Available: []string{"zz"}},
		}},
		Bitmaps: map[string]store.Bitmap{"dev": bm},
		Base:    "https://example.test/shortfree",
		PerPage: 1000,
	}
}

func TestBuildPages(t *testing.T) {
	files, err := buildPages(pagesFixture())
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		path    string
		want    []string
		notWant []string
	}{
		{"dev/index.html", []string{
			"<h1>Free short .dev domains</h1>",
			`<link rel="canonical" href="https://example.test/shortfree/dev/">`,
			`href="../style.css"`,
			`<a href="2-letter/">2-letter .dev</a>`,
			"<span>x.dev</span>", "<span>ab.dev</span>",
			"zz.dev", `data-state="dropping"`,
			"2026-10-26 06:30 UTC",
			`<a href="../com/">.com</a>`,
		}, []string{"<span>q7.dev</span>", "noindex", `href="../api/"`}},
		{"dev/2-letter/index.html", []string{
			"<h1>Free 2-letter .dev domains</h1>",
			"Letters only · 1", "With digits or a hyphen · 1",
			"<span>ab.dev</span>", "<span>q7.dev</span>",
			`href="../../style.css"`,
			`"@type":"BreadcrumbList"`, `"@type":"Dataset"`,
			`og:image" content="https://maximilianfeix.github.io/shortfree/social-preview.png"`,
		}, []string{"noindex", "<span>x.dev</span>", `class="pager"`}},
		{"dev/3-letter/index.html", []string{
			"2,500 free 3-letter .dev domains", "Page 1 of 3",
			`<span aria-current="page">1</span>`, `<a href="../../dev/3-letter/page/2/">2</a>`,
			"1 more 3 character names are not in DNS",
		}, nil},
		{"dev/3-letter/page/3/index.html", []string{
			"<h1>Free 3-letter .dev domains, page 3</h1>",
			`href="../../../../style.css"`,
			`<link rel="canonical" href="https://example.test/shortfree/dev/3-letter/page/3/">`,
			"(continued)",
		}, []string{"Page 1 of"}},
		{"dev/4-letter/index.html", []string{
			"<h1>Free 4-letter .dev domains</h1>",
			"<span>toka.dev</span>", "<span>bela.dev</span>",
			"four-dev.txt.gz",
		}, []string{"zzzz.dev</span>", "ab1c", "noindex"}},
		{"com/1-letter/index.html", []string{
			`<meta name="robots" content="noindex, follow">`,
			"No free 1-letter .com domains", "All 36 valid",
		}, []string{`class="domain-grid"`}},
		{"robots.txt", []string{"Sitemap: https://example.test/shortfree/sitemap.xml"}, nil},
		{"sitemap.xml", []string{
			"<loc>https://example.test/shortfree/</loc>",
			"<loc>https://example.test/shortfree/dev/3-letter/page/3/</loc>",
			"<lastmod>2026-10-26</lastmod>",
		}, []string{"com/1-letter/", "/api/"}},
	}
	for _, c := range cases {
		t.Run(c.path, func(t *testing.T) {
			b, ok := files[c.path]
			if !ok {
				t.Fatalf("%s not generated", c.path)
			}
			s := string(b)
			for _, w := range c.want {
				if !strings.Contains(s, w) {
					t.Errorf("missing %q", w)
				}
			}
			for _, w := range c.notWant {
				if strings.Contains(s, w) {
					t.Errorf("unexpected %q", w)
				}
			}
		})
	}
	if _, ok := files["api/index.html"]; ok {
		t.Error("a TLD named api must not overwrite the API directory")
	}
	if _, ok := files["dev/3-letter/page/4/index.html"]; ok {
		t.Error("2,500 names at 1,000 per page are 3 pages")
	}
}

func TestPagesStructuredData(t *testing.T) {
	files, err := buildPages(pagesFixture())
	if err != nil {
		t.Fatal(err)
	}
	ld := regexp.MustCompile(`(?s)<script type="application/ld\+json">(.*?)</script>`)
	h1 := regexp.MustCompile(`<h1>`)
	n := 0
	for path, b := range files {
		if !strings.HasSuffix(path, ".html") {
			continue
		}
		n++
		if c := len(h1.FindAll(b, -1)); c != 1 {
			t.Errorf("%s: %d h1 elements, want 1", path, c)
		}
		m := ld.FindSubmatch(b)
		if m == nil {
			t.Errorf("%s: no JSON-LD", path)
			continue
		}
		var objs []map[string]any
		if err := json.Unmarshal(m[1], &objs); err != nil {
			t.Errorf("%s: JSON-LD does not parse: %v", path, err)
			continue
		}
		types := map[string]bool{}
		for _, o := range objs {
			types[fmt.Sprint(o["@type"])] = true
		}
		if !types["Dataset"] || !types["BreadcrumbList"] || types["FAQPage"] {
			t.Errorf("%s: JSON-LD types %v", path, types)
		}
	}
	if n < 10 {
		t.Errorf("only %d pages", n)
	}
	var sm struct {
		URLs []struct {
			Loc string `xml:"loc"`
		} `xml:"url"`
	}
	if err := xml.Unmarshal(files["sitemap.xml"], &sm); err != nil {
		t.Fatalf("sitemap: %v", err)
	}
	if len(sm.URLs) == 0 {
		t.Fatal("empty sitemap")
	}
}

func TestPager(t *testing.T) {
	href := func(k int) string { return fmt.Sprint(k) }
	cases := []struct {
		cur, total int
		want       string
	}{
		{1, 1, ""},
		{2, 3, "1 [2] 3"},
		{1, 20, "[1] 2 3 … 20"},
		{10, 20, "1 … 8 9 [10] 11 12 … 20"},
		{20, 20, "1 … 18 19 [20]"},
	}
	for _, c := range cases {
		var parts []string
		for _, l := range pager(c.cur, c.total, href) {
			switch {
			case l.Gap:
				parts = append(parts, "…")
			case l.Current:
				parts = append(parts, "["+l.Label+"]")
			default:
				parts = append(parts, l.Label)
			}
		}
		if got := strings.Join(parts, " "); got != c.want {
			t.Errorf("pager(%d, %d) = %q, want %q", c.cur, c.total, got, c.want)
		}
	}
}

func TestBuildPagesNeedsData(t *testing.T) {
	if _, err := buildPages(pageInput{}); err == nil {
		t.Error("want error without stats")
	}
}
