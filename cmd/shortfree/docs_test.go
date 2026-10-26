package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var (
	mdLink    = regexp.MustCompile(`\]\(([^)\s]+)\)`)
	htmlLink  = regexp.MustCompile(`(?:href|src|srcset)="([^"]+)"`)
	htmlID    = regexp.MustCompile(`<a id="([^"]+)">`)
	heading   = regexp.MustCompile(`(?m)^#{1,6} +(.+)$`)
	slugStrip = regexp.MustCompile(`[^\p{L}\p{N} _-]`)
)

// anchors returns the link targets inside a Markdown file: explicit
// <a id="..."> tags and GitHub's heading slugs.
func anchors(md string) map[string]bool {
	out := map[string]bool{}
	for _, m := range htmlID.FindAllStringSubmatch(md, -1) {
		out[m[1]] = true
	}
	for _, m := range heading.FindAllStringSubmatch(md, -1) {
		s := strings.ToLower(strings.TrimSpace(m[1]))
		s = strings.ReplaceAll(slugStrip.ReplaceAllString(s, ""), " ", "-")
		out[s] = true
	}
	return out
}

// TestDocLinks checks that every relative link and #anchor in the READMEs and
// docs/ points to something that exists. The generated block between the
// shortfree markers is skipped: it links to data/, which only the scan creates.
func TestDocLinks(t *testing.T) {
	root := filepath.Join("..", "..")
	files, _ := filepath.Glob(filepath.Join(root, "*.md"))
	docs, _ := filepath.Glob(filepath.Join(root, "docs", "*.md"))
	files = append(files, docs...)
	if len(files) < 4 {
		t.Fatalf("found only %d Markdown files", len(files))
	}
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		md := string(b)
		if i, j := strings.Index(md, markStart), strings.Index(md, markEnd); i >= 0 && j > i {
			md = md[:i] + md[j:]
		}
		// Code blocks hold examples, not links.
		parts := strings.Split(md, "```")
		for i := 1; i < len(parts); i += 2 {
			parts[i] = ""
		}
		text := strings.Join(parts, "")

		var links []string
		for _, re := range []*regexp.Regexp{mdLink, htmlLink} {
			for _, m := range re.FindAllStringSubmatch(text, -1) {
				links = append(links, m[1])
			}
		}
		for _, l := range links {
			if strings.Contains(l, "://") || strings.HasPrefix(l, "mailto:") {
				continue
			}
			target, frag, _ := strings.Cut(l, "#")
			path := f
			if target != "" {
				path = filepath.Join(filepath.Dir(f), filepath.FromSlash(target))
				if _, err := os.Stat(path); err != nil {
					t.Errorf("%s: broken link %q", f, l)
					continue
				}
			}
			if frag == "" || !strings.HasSuffix(path, ".md") {
				continue
			}
			tb, err := os.ReadFile(path)
			if err != nil {
				t.Errorf("%s: %v", f, err)
				continue
			}
			if !anchors(string(tb))[frag] {
				t.Errorf("%s: link %q: no anchor #%s in %s", f, l, frag, path)
			}
		}
	}
}

func TestAnchors(t *testing.T) {
	a := anchors("<a id=\"try-it\"></a>\n## What it does, and what it can't do\n### In GitHub Actions\n")
	for _, want := range []string{"try-it", "what-it-does-and-what-it-cant-do", "in-github-actions"} {
		if !a[want] {
			t.Errorf("anchor %q missing in %v", want, a)
		}
	}
}
