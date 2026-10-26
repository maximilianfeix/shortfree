package main

import (
	"sort"
	"strings"
	"unicode"

	"github.com/maximilianfeix/shortfree/internal/names"
)

// candidate is one short label derived from a word or brand idea.
type candidate struct {
	Label  string `json:"label"`
	Reason string `json:"reason"`
	Score  int    `json:"score"`
}

var translit = strings.NewReplacer(
	"ä", "a", "ö", "o", "ü", "u", "ß", "ss", "à", "a", "á", "a", "â", "a", "å", "a",
	"ç", "c", "è", "e", "é", "e", "ê", "e", "ë", "e", "ì", "i", "í", "i", "î", "i", "ï", "i",
	"ñ", "n", "ò", "o", "ó", "o", "ô", "o", "ø", "o", "ù", "u", "ú", "u", "û", "u", "ý", "y", "ÿ", "y",
	"&", " and ", "+", " plus ",
)

var stopwords = map[string]bool{"the": true, "a": true, "an": true, "of": true, "and": true, "for": true, "to": true, "in": true, "on": true, "my": true}

// splitWords turns "OpenAI Labs", "open-ai_labs" or "Café Ölberg" into
// lowercase ASCII words; camelCase boundaries split too.
func splitWords(s string) []string {
	var b strings.Builder
	rs := []rune(s)
	for i, r := range rs {
		if i > 0 && unicode.IsUpper(r) && unicode.IsLower(rs[i-1]) {
			b.WriteRune(' ')
		}
		b.WriteRune(unicode.ToLower(r))
	}
	t := translit.Replace(b.String())
	return strings.FieldsFunc(t, func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9')
	})
}

// skeleton keeps the first letter and the consonants after it, collapsing
// repeats: "flicker" -> "flckr", "tumbler" -> "tmblr".
func skeleton(w string) string {
	if w == "" {
		return ""
	}
	b := []byte{w[0]}
	for i := 1; i < len(w); i++ {
		c := w[i]
		if strings.IndexByte("aeiou", c) >= 0 || c == b[len(b)-1] {
			continue
		}
		b = append(b, c)
	}
	return string(b)
}

func prefix(s string, n int) string {
	if len(s) < n {
		return s
	}
	return s[:n]
}

// suggestCandidates derives short labels (at most maxLen characters) from
// idea, best first. The heuristic is deterministic:
//
//	100 the word itself, if short enough
//	 95 initials of a multi-word idea ("deep mind labs" -> dml)
//	 90 a single short word of the idea
//	 85 initials of the first two words (dm)
//	 80 first 4 / 3 / 2 letters (75, 70)
//	 78 consonant skeleton, 4 / 3 letters (73)
//	 72 blends of two words: 2+2, 1+3, 3+1 letters (68), three words 2+1+1 (66)
//	 55 a base plus one digit, 1-9 in order (55 down to 47)
//	 50 a base plus "x", 2 letters plus "hq" or "ly" (48), "go"/"my" plus 2 letters (45)
//
// Bases for the digit and affix patterns are the initials (all words and
// first two), the first 2-3 letters and the 3 letter skeleton.
func suggestCandidates(idea string, maxLen int) []candidate {
	if maxLen < 1 || maxLen > 4 {
		maxLen = 4
	}
	words := splitWords(idea)
	if len(words) == 0 {
		return nil
	}
	if len(words) > 1 {
		var kept []string
		for _, w := range words {
			if !stopwords[w] {
				kept = append(kept, w)
			}
		}
		if len(kept) > 0 {
			words = kept
		}
	}
	joined := strings.Join(words, "")

	best := map[string]candidate{}
	add := func(label, reason string, score int) {
		if label == "" || len(label) > maxLen || !names.Valid(label) || names.Index(label) < 0 {
			return
		}
		if c, ok := best[label]; ok && c.Score >= score {
			return
		}
		best[label] = candidate{label, reason, score}
	}

	add(joined, "the word itself", 100)
	var initials string
	if len(words) > 1 {
		for _, w := range words {
			initials += w[:1]
		}
		add(initials, "initials", 95)
		if len(words) > 2 {
			add(initials[:2], "initials of the first two words", 85)
		}
		for _, w := range words {
			add(w, "one word of the idea", 90)
		}
	}
	add(prefix(joined, 4), "first 4 letters", 80)
	add(prefix(joined, 3), "first 3 letters", 75)
	add(prefix(joined, 2), "first 2 letters", 70)
	sk := skeleton(joined)
	add(prefix(sk, 4), "consonant skeleton (vowels dropped)", 78)
	add(prefix(sk, 3), "consonant skeleton (vowels dropped)", 73)
	if len(words) >= 2 {
		w1, w2 := words[0], words[1]
		add(prefix(w1, 2)+prefix(w2, 2), "first two letters of each word", 72)
		add(prefix(w1, 1)+prefix(w2, 3), "blend of both words", 68)
		add(prefix(w1, 3)+prefix(w2, 1), "blend of both words", 68)
	}
	if len(words) >= 3 {
		add(prefix(words[0], 2)+words[1][:1]+words[2][:1], "blend of three words", 66)
	}

	var bases []string
	for _, b := range []string{initials, prefix(initials, 2), prefix(joined, 3), prefix(joined, 2), prefix(sk, 3)} {
		if b != "" && !contains(bases, b) {
			bases = append(bases, b)
		}
	}
	for _, b := range bases {
		for d := 1; d <= 9; d++ {
			add(b+string(rune('0'+d)), "startup pattern: add a digit", 56-d)
		}
		add(b+"x", "startup pattern: add x", 50)
		if len(b) == 2 {
			add(b+"hq", "startup pattern: add hq", 48)
			add(b+"ly", "startup pattern: add ly", 48)
			add("go"+b, "startup pattern: go + name", 45)
			add("my"+b, "startup pattern: my + name", 45)
		}
	}

	out := make([]candidate, 0, len(best))
	for _, c := range best {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.Score != b.Score {
			return a.Score > b.Score
		}
		if len(a.Label) != len(b.Label) {
			return len(a.Label) > len(b.Label) // longer is more readable
		}
		return a.Label < b.Label
	})
	return out
}
