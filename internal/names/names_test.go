package names

import "testing"

func TestNameIndexRoundTrip(t *testing.T) {
	for n := 1; n <= 3; n++ {
		for i := 0; i < Space(n); i++ {
			if got := Index(Name(i, n)); got != i {
				t.Fatalf("Index(Name(%d,%d)) = %d", i, n, got)
			}
		}
	}
}

func TestValid(t *testing.T) {
	for s, want := range map[string]bool{
		"a": true, "q7": true, "a-b": true, "ab-c": true,
		"-a": false, "a-": false, "-": false, "a--b": false, "": false,
	} {
		if got := Valid(s); got != want {
			t.Errorf("Valid(%q) = %v, want %v", s, got, want)
		}
	}
}

func TestEachCounts(t *testing.T) {
	// 36 + 36² + (36³ + 36·36 hyphenated middles) = 36 + 1296 + 47952.
	want := map[int]int{1: 36, 2: 1296, 3: 47952}
	for n, w := range want {
		c := 0
		Each(n, 0, 1, func(int, string) { c++ })
		if c != w {
			t.Errorf("Each(%d) visited %d labels, want %d", n, c, w)
		}
	}
}

func TestEachShardsPartition(t *testing.T) {
	total := 0
	Each(3, 0, 1, func(int, string) { total++ })
	sum := 0
	seen := map[string]bool{}
	for s := 0; s < 7; s++ {
		Each(3, s, 7, func(_ int, l string) {
			if seen[l] {
				t.Fatalf("%q in two shards", l)
			}
			seen[l] = true
			sum++
		})
	}
	if sum != total {
		t.Fatalf("shards cover %d labels, want %d", sum, total)
	}
}

func TestKind(t *testing.T) {
	for s, want := range map[string]string{"abc": "letters", "123": "digits", "a1": "mixed", "a-b": "mixed"} {
		if got := Kind(s); got != want {
			t.Errorf("Kind(%q) = %q, want %q", s, got, want)
		}
	}
}
