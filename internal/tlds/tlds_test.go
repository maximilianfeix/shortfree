package tlds

import "testing"

func TestAllowed(t *testing.T) {
	cases := []struct {
		tld, label string
		want       bool
	}{
		{"com", "a", true},
		{"com", "ab", true},
		{"ly", "abc", false},
		{"ly", "abcd", true},
		{"so", "ab", false},
		{"so", "abc", true},
		{"to", "a", false},
		{"to", "q7x", true},
		{"unknown", "x", true},
	}
	for _, c := range cases {
		if got := Allowed(c.tld, c.label); got != c.want {
			t.Errorf("Allowed(%q, %q) = %v, want %v", c.tld, c.label, got, c.want)
		}
	}
}

func TestGetDefaultsToOne(t *testing.T) {
	if p := Get("dev"); p.MinLen != 1 || p.Canary != "nic" {
		t.Errorf("Get(dev) = %+v, want MinLen 1, Canary nic", p)
	}
}

func TestDefaultListIsClean(t *testing.T) {
	seen := map[string]bool{}
	for _, tld := range Default {
		if tld == "" || seen[tld] {
			t.Errorf("empty or duplicate TLD %q", tld)
		}
		seen[tld] = true
	}
	for tld, p := range policies {
		if !seen[tld] {
			t.Errorf("policy for %q, which is not in Default", tld)
		}
		if p.MinLen < 1 || p.MinLen > 4 {
			t.Errorf("%s: MinLen %d outside 1-4", tld, p.MinLen)
		}
	}
}
