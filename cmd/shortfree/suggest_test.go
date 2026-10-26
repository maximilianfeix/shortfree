package main

import (
	"reflect"
	"testing"
)

func TestSplitWords(t *testing.T) {
	cases := map[string][]string{
		"Deep Mind Labs":   {"deep", "mind", "labs"},
		"openAI_labs":      {"open", "ai", "labs"},
		"Café Ölberg":      {"cafe", "olberg"},
		"rock&roll":        {"rock", "and", "roll"},
		"  web3 ---tools ": {"web3", "tools"},
		"!!!":              nil,
	}
	for in, want := range cases {
		if got := splitWords(in); len(got)+len(want) > 0 && !reflect.DeepEqual(got, want) {
			t.Errorf("splitWords(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSkeleton(t *testing.T) {
	for in, want := range map[string]string{"flicker": "flckr", "tumbler": "tmblr", "apple": "apl", "x": "x"} {
		if got := skeleton(in); got != want {
			t.Errorf("skeleton(%q) = %q, want %q", in, got, want)
		}
	}
}

func labelsOf(c []candidate) []string {
	var out []string
	for _, x := range c {
		out = append(out, x.Label)
	}
	return out
}

func TestSuggestCandidates(t *testing.T) {
	got := suggestCandidates("Flicker", 4)
	want := []string{"flic", "flck", "fli", "flc", "fl", "flc1", "fli1", "fl1", "flc2"}
	if l := labelsOf(got); !reflect.DeepEqual(l[:len(want)], want) {
		t.Errorf("Flicker: %v, want prefix %v", l, want)
	}

	got = suggestCandidates("The Deep Mind Labs", 4)
	if got[0].Label != "dml" || got[0].Reason != "initials" || got[0].Score != 95 {
		t.Errorf("first candidate = %+v, want initials dml", got[0])
	}
	seen := map[string]candidate{}
	for _, c := range got {
		if len(c.Label) > 4 {
			t.Errorf("%q longer than 4", c.Label)
		}
		if _, dup := seen[c.Label]; dup {
			t.Errorf("duplicate %q", c.Label)
		}
		seen[c.Label] = c
	}
	for _, l := range []string{"deep", "mind", "demi", "dmin", "deem", "dml1", "dmlx", "dm1"} {
		if _, ok := seen[l]; !ok {
			t.Errorf("deep mind labs: missing %q in %v", l, labelsOf(got))
		}
	}
	if _, ok := seen["t"]; ok {
		t.Error("stop word 'the' was used")
	}

	short := suggestCandidates("ox", 4)
	if short[0].Label != "ox" || short[0].Score != 100 {
		t.Errorf("ox: %+v", short[0])
	}
	if l := labelsOf(suggestCandidates("Flicker", 2)); !reflect.DeepEqual(l, []string{"fl"}) {
		t.Errorf("max length 2: %v", l)
	}
	// Deterministic.
	if !reflect.DeepEqual(suggestCandidates("Deep Mind Labs", 4), suggestCandidates("Deep Mind Labs", 4)) {
		t.Error("not deterministic")
	}
	if suggestCandidates("", 4) != nil {
		t.Error("empty idea produced candidates")
	}
}
