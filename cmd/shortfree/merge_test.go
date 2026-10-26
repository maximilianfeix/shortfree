package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/maximilianfeix/shortfree/internal/names"
	"github.com/maximilianfeix/shortfree/internal/store"
)

// writePart creates dir/state/state.json.gz with the given TLDs and
// dir/bitmaps/<tld>.bin with one bit set per entry of bits.
func writePart(t *testing.T, dir string, updated time.Time, tldMap map[string]map[string]*store.Entry, bits map[string][]string) {
	t.Helper()
	st := &store.State{Updated: updated, TLDs: tldMap}
	if st.TLDs == nil {
		st.TLDs = map[string]map[string]*store.Entry{}
	}
	if err := st.Save(filepath.Join(dir, "state", "state.json.gz")); err != nil {
		t.Fatal(err)
	}
	for tld, labels := range bits {
		bm := store.NewBitmap(names.Space(4))
		for _, l := range labels {
			bm.Set(names.Index(l), true)
		}
		if err := store.WriteAtomic(filepath.Join(dir, "bitmaps", tld+".bin"), bm); err != nil {
			t.Fatal(err)
		}
	}
}

func free(checked int64) *store.Entry {
	return &store.Entry{DNS: "free", RDAP: "available", Checked: checked}
}

func entryLabels(m map[string]*store.Entry) []string {
	var out []string
	for l := range m {
		out = append(out, l)
	}
	store.SortLabels(out)
	return out
}

func setBits(bm store.Bitmap) []string {
	var out []string
	names.Each(4, 0, 1, func(i int, s string) {
		if bm.Get(i) {
			out = append(out, s)
		}
	})
	return out
}

func TestMergeOwnedReplacesWholeTLD(t *testing.T) {
	root := t.TempDir()
	day1 := time.Date(2026, 10, 26, 3, 0, 0, 0, time.UTC)
	day2 := day1.AddDate(0, 0, 1)
	base := filepath.Join(root, "prev")
	writePart(t, base, day1, map[string]map[string]*store.Entry{
		"dev": {"aa": free(1), "bb": free(1)},
		"io":  {"cc": free(1)},
		"de":  {"dd": free(1)},
	}, map[string][]string{"dev": {"aaaa", "bbbb"}, "io": {"cccc"}})

	// dev: "aa" was registered overnight, the newer check for "bb" carries
	// an older timestamp than base (must still win: the owner is the truth).
	writePart(t, filepath.Join(root, "dev"), day2, map[string]map[string]*store.Entry{
		"dev": {"bb": free(0)},
	}, map[string][]string{"dev": {"bbbb"}})
	// de: every name was registered; the job still owns the (empty) TLD.
	writePart(t, filepath.Join(root, "de"), day2, map[string]map[string]*store.Entry{
		"de": {},
	}, nil)
	// io's job failed: no part, so base survives.

	st, bms, err := mergeOwned(base, []string{filepath.Join(root, "dev"), filepath.Join(root, "de")})
	if err != nil {
		t.Fatal(err)
	}
	if got := entryLabels(st.TLDs["dev"]); !reflect.DeepEqual(got, []string{"bb"}) {
		t.Errorf("dev = %v, want [bb] (aa became taken)", got)
	}
	if got := entryLabels(st.TLDs["de"]); len(got) != 0 {
		t.Errorf("de = %v, want empty", got)
	}
	if got := entryLabels(st.TLDs["io"]); !reflect.DeepEqual(got, []string{"cc"}) {
		t.Errorf("io = %v, want base [cc]", got)
	}
	if got := setBits(bms["dev"]); !reflect.DeepEqual(got, []string{"bbbb"}) {
		t.Errorf("dev bitmap = %v, want [bbbb] (replaced, not OR-ed)", got)
	}
	if got := setBits(bms["io"]); !reflect.DeepEqual(got, []string{"cccc"}) {
		t.Errorf("io bitmap = %v, want base [cccc]", got)
	}
	if !st.Updated.Equal(day2) {
		t.Errorf("updated = %v, want %v", st.Updated, day2)
	}
}

func TestMergeOwnedWithoutBase(t *testing.T) {
	root := t.TempDir()
	writePart(t, filepath.Join(root, "xyz"), time.Now(), map[string]map[string]*store.Entry{"xyz": {"q7": free(1)}}, nil)
	st, bms, err := mergeOwned("", []string{filepath.Join(root, "xyz")})
	if err != nil {
		t.Fatal(err)
	}
	if len(st.TLDs) != 1 || len(st.TLDs["xyz"]) != 1 || len(bms) != 0 {
		t.Errorf("state %v, bitmaps %d", st.TLDs, len(bms))
	}
}

func TestMergeOwnedMissingBaseIsFirstRun(t *testing.T) {
	root := t.TempDir()
	writePart(t, filepath.Join(root, "co"), time.Now(), map[string]map[string]*store.Entry{"co": {"q7": free(1)}}, map[string][]string{"co": {"q7q7"}})
	st, bms, err := mergeOwned(filepath.Join(root, "empty-prev"), []string{filepath.Join(root, "co")})
	if err != nil {
		t.Fatal(err)
	}
	if len(st.TLDs["co"]) != 1 || bms["co"] == nil {
		t.Errorf("state %v, bitmaps %v", st.TLDs, len(bms))
	}
}

func TestMergeOwnedRejectsTwoOwners(t *testing.T) {
	root := t.TempDir()
	writePart(t, filepath.Join(root, "a"), time.Now(), map[string]map[string]*store.Entry{"dev": {}}, nil)
	writePart(t, filepath.Join(root, "b"), time.Now(), map[string]map[string]*store.Entry{"dev": {}}, nil)
	if _, _, err := mergeOwned("", []string{filepath.Join(root, "a"), filepath.Join(root, "b")}); err == nil || !strings.Contains(err.Error(), ".dev state") {
		t.Errorf("err = %v, want duplicate state owner", err)
	}

	writePart(t, filepath.Join(root, "c"), time.Now(), nil, map[string][]string{"io": {"aaaa"}})
	writePart(t, filepath.Join(root, "d"), time.Now(), nil, map[string][]string{"io": {"bbbb"}})
	if _, _, err := mergeOwned("", []string{filepath.Join(root, "c"), filepath.Join(root, "d")}); err == nil || !strings.Contains(err.Error(), ".io bitmap") {
		t.Errorf("err = %v, want duplicate bitmap owner", err)
	}
}

func TestMergeRejectsTruncatedBitmap(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "dev")
	writePart(t, dir, time.Now(), nil, nil)
	if err := store.WriteAtomic(filepath.Join(dir, "bitmaps", "dev.bin"), []byte{1, 2, 3}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := mergeOwned("", []string{dir}); err == nil {
		t.Error("truncated bitmap accepted")
	}
}

func TestMergeOrUnionsBootstrapShards(t *testing.T) {
	root := t.TempDir()
	writePart(t, filepath.Join(root, "short-dev"), time.Now(), map[string]map[string]*store.Entry{
		"dev": {"aa": free(5)},
	}, nil)
	writePart(t, filepath.Join(root, "short-dev2"), time.Now(), map[string]map[string]*store.Entry{
		"dev": {"aa": {DNS: "free", RDAP: "registered", Checked: 3}, "bb": free(1)},
	}, nil)
	writePart(t, filepath.Join(root, "four-0"), time.Now(), nil, map[string][]string{"dev": {"aaaa"}})
	writePart(t, filepath.Join(root, "four-1"), time.Now(), nil, map[string][]string{"dev": {"bbbb"}})
	st, bms, err := mergeOr([]string{
		filepath.Join(root, "short-dev"), filepath.Join(root, "short-dev2"),
		filepath.Join(root, "four-0"), filepath.Join(root, "four-1"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if e := st.TLDs["dev"]["aa"]; e.RDAP != "available" {
		t.Errorf("aa = %+v, want the newer check", e)
	}
	if len(st.TLDs["dev"]) != 2 {
		t.Errorf("dev = %v", entryLabels(st.TLDs["dev"]))
	}
	if got := setBits(bms["dev"]); !reflect.DeepEqual(got, []string{"aaaa", "bbbb"}) {
		t.Errorf("bitmap = %v, want OR of both shards", got)
	}
}

func TestCmdMergeWritesFiles(t *testing.T) {
	root := t.TempDir()
	writePart(t, filepath.Join(root, "prev"), time.Now(), map[string]map[string]*store.Entry{"io": {"cc": free(1)}}, map[string][]string{"io": {"cccc"}})
	writePart(t, filepath.Join(root, "parts", "tld-dev"), time.Now(), map[string]map[string]*store.Entry{"dev": {"q7": free(1)}}, map[string][]string{"dev": {"q7q7"}})
	statePath := filepath.Join(root, "out", "state.json.gz")
	bmDir := filepath.Join(root, "out", "bitmaps")
	err := cmdMerge([]string{"-state", statePath, "-bitmaps", bmDir, "-base", filepath.Join(root, "prev"), filepath.Join(root, "parts", "tld-dev")})
	if err != nil {
		t.Fatal(err)
	}
	st, err := store.Load(statePath)
	if err != nil {
		t.Fatal(err)
	}
	if len(st.TLDs["io"]) != 1 || len(st.TLDs["dev"]) != 1 {
		t.Errorf("merged state = %v", st.TLDs)
	}
	for _, f := range []string{"io.bin", "dev.bin"} {
		if _, err := os.Stat(filepath.Join(bmDir, f)); err != nil {
			t.Error(err)
		}
	}
	if err := cmdMerge([]string{"-or", "-base", "x", "y"}); err == nil {
		t.Error("-or with -base accepted")
	}
}

func TestKeepTLDs(t *testing.T) {
	st := &store.State{TLDs: map[string]map[string]*store.Entry{"dev": {}, "io": {"a": free(1)}, "co": {}}}
	keepTLDs(st, []string{"io"})
	if len(st.TLDs) != 1 || st.TLDs["io"] == nil {
		t.Errorf("TLDs = %v, want only io", st.TLDs)
	}
}

func TestRDAPBudgets(t *testing.T) {
	st := &store.State{TLDs: map[string]map[string]*store.Entry{"dev": {"a": free(1)}, "de": {}}}
	got := rdapBudgets(st, []string{"dev", "de", "xyz"}, 90*time.Minute, 270*time.Minute)
	want := map[string]time.Duration{"dev": 90 * time.Minute, "de": 270 * time.Minute, "xyz": 270 * time.Minute}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("budgets = %v, want %v", got, want)
	}
	if got := rdapBudgets(st, []string{"xyz"}, 90*time.Minute, 0); got["xyz"] != 90*time.Minute {
		t.Errorf("no -rdap-time-new: %v", got)
	}
	if got := rdapBudgets(st, []string{"xyz"}, 0, 270*time.Minute); got["xyz"] != 0 {
		t.Errorf("-rdap-time 0 must skip RDAP even for new TLDs: %v", got)
	}
}

func TestShortLabelsRespectMinLen(t *testing.T) {
	for _, c := range []struct {
		tld      string
		shortest int
	}{{"dev", 1}, {"so", 3}, {"ly", 4}} {
		labels := shortLabels(c.tld, 3)
		for _, l := range labels {
			if len(l) < c.shortest {
				t.Errorf(".%s: swept %q below minimum %d", c.tld, l, c.shortest)
				break
			}
		}
		if c.tld == "ly" && len(labels) != 0 {
			t.Errorf(".ly: %d labels up to 3 chars, want none", len(labels))
		}
	}
}

func TestPruneBelowMinLen(t *testing.T) {
	st := &store.State{TLDs: map[string]map[string]*store.Entry{"so": {"a": free(1), "ab": free(1), "abc": free(1)}}}
	pruneBelowMinLen(st, "so")
	if got := entryLabels(st.TLDs["so"]); !reflect.DeepEqual(got, []string{"abc"}) {
		t.Errorf("so = %v, want [abc]", got)
	}
}

func TestCmdTLDsJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := cmdTLDs(&buf, []string{"-json"}); err != nil {
		t.Fatal(err)
	}
	var got []string
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("%q: %v", buf.String(), err)
	}
	if !reflect.DeepEqual(got, DefaultTLDs) || strings.Count(buf.String(), "\n") != 1 {
		t.Errorf("tlds -json = %q", buf.String())
	}
	buf.Reset()
	if err := cmdTLDs(&buf, nil); err != nil {
		t.Fatal(err)
	}
	if lines := strings.Fields(buf.String()); !reflect.DeepEqual(lines, DefaultTLDs) {
		t.Errorf("tlds = %v", lines)
	}
}

func TestExportSkipsNamesBelowMinLen(t *testing.T) {
	dir := t.TempDir()
	st := &store.State{Updated: time.Now().UTC(), TLDs: map[string]map[string]*store.Entry{
		"so": {"ab": free(1), "abc": free(1)},
	}}
	statePath := filepath.Join(dir, "state.json.gz")
	if err := st.Save(statePath); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "data")
	err := cmdExport([]string{"-tlds", "so", "-state", statePath, "-bitmaps", filepath.Join(dir, "none"),
		"-out", out, "-four-out", "", "-readme", "", "-api-out", ""})
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(out, "txt", "so.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "abc.so\n" {
		t.Errorf("so.txt = %q, want only abc.so", b)
	}
}
