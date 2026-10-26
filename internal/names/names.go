// Package names enumerates every valid short DNS label.
//
// Labels are indexed in base 37 over the alphabet a-z, 0-9 and "-". An index
// maps to exactly one string of a given length; indices whose string is not a
// valid LDH label (leading/trailing hyphen, "--") are skipped by Valid.
package names

import "strings"

// Alphabet is the ordered character set used for indexing.
const Alphabet = "abcdefghijklmnopqrstuvwxyz0123456789-"

// Base is len(Alphabet).
const Base = len(Alphabet)

// Space returns the size of the index space for labels of length n (37^n).
func Space(n int) int {
	s := 1
	for i := 0; i < n; i++ {
		s *= Base
	}
	return s
}

// Name returns the label of length n at index i.
func Name(i, n int) string {
	b := make([]byte, n)
	for p := n - 1; p >= 0; p-- {
		b[p] = Alphabet[i%Base]
		i /= Base
	}
	return string(b)
}

// Index returns the index of label s, or -1 if s contains a character outside
// the alphabet.
func Index(s string) int {
	i := 0
	for _, c := range []byte(s) {
		d := strings.IndexByte(Alphabet, c)
		if d < 0 {
			return -1
		}
		i = i*Base + d
	}
	return i
}

// Valid reports whether s is a registrable LDH label: no leading or trailing
// hyphen and no consecutive hyphens.
func Valid(s string) bool {
	if s == "" || s[0] == '-' || s[len(s)-1] == '-' {
		return false
	}
	return !strings.Contains(s, "--")
}

// Each calls fn for every valid label of length n whose first character falls
// into shard (first-character index modulo shards). shards <= 1 means all.
func Each(n, shard, shards int, fn func(i int, s string)) {
	space := Space(n)
	for i := 0; i < space; i++ {
		s := Name(i, n)
		if !Valid(s) {
			continue
		}
		if shards > 1 && strings.IndexByte(Alphabet, s[0])%shards != shard {
			continue
		}
		fn(i, s)
	}
}

// Kind classifies a label for filtering: "letters", "digits" or "mixed".
func Kind(s string) string {
	letters, digits := 0, 0
	for _, c := range []byte(s) {
		switch {
		case c >= 'a' && c <= 'z':
			letters++
		case c >= '0' && c <= '9':
			digits++
		}
	}
	switch {
	case letters == len(s):
		return "letters"
	case digits == len(s):
		return "digits"
	}
	return "mixed"
}
