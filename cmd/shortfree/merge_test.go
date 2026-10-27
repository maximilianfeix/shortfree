package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
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

var noJump = jumpLimit{}

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

	st, bms, _, err := mergeOwned(base, []string{filepath.Join(root, "dev"), filepath.Join(root, "de")}, noJump)
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
	st, bms, _, err := mergeOwned("", []string{filepath.Join(root, "xyz")}, noJump)
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
	st, bms, _, err := mergeOwned(filepath.Join(root, "empty-prev"), []string{filepath.Join(root, "co")}, noJump)
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
	if _, _, _, err := mergeOwned("", []string{filepath.Join(root, "a"), filepath.Join(root, "b")}, noJump); err == nil || !strings.Contains(err.Error(), ".dev state") {
		t.Errorf("err = %v, want duplicate state owner", err)
	}

	writePart(t, filepath.Join(root, "c"), time.Now(), nil, map[string][]string{"io": {"aaaa"}})
	writePart(t, filepath.Join(root, "d"), time.Now(), nil, map[string][]string{"io": {"bbbb"}})
	if _, _, _, err := mergeOwned("", []string{filepath.Join(root, "c"), filepath.Join(root, "d")}, noJump); err == nil || !strings.Contains(err.Error(), ".io bitmap") {
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
	if _, _, _, err := mergeOwned("", []string{dir}, noJump); err == nil {
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
	st, bms, _, err := mergeOr([]string{
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
	now := time.Now().UTC()
	st := &store.State{Updated: now, TLDs: map[string]map[string]*store.Entry{
		"so": {"ab": free(now.Unix()), "abc": free(now.Unix())},
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

// writeScanPart writes a part like a scan job does: state with a scan record
// for every TLD at at, and per TLD in bits a bitmap whose coverage says shard
// 0 was swept at at.
func writeScanPart(t *testing.T, dir string, at time.Time, tldMap map[string]map[string]*store.Entry, bits map[string][]string) {
	t.Helper()
	writePart(t, dir, at, tldMap, bits)
	st, err := store.Load(filepath.Join(dir, "state", "state.json.gz"))
	if err != nil {
		t.Fatal(err)
	}
	for tld := range tldMap {
		st.SetScan(tld, at)
	}
	if err := st.Save(filepath.Join(dir, "state", "state.json.gz")); err != nil {
		t.Fatal(err)
	}
	for tld := range bits {
		c := &store.Coverage{Shards: 7, Scans: map[string]*store.ShardScan{"0": {At: at.Unix(), Names: 1}}}
		if err := c.Save(filepath.Join(dir, "bitmaps", tld+".cov.json")); err != nil {
			t.Fatal(err)
		}
	}
}

// A part left in the data release must never roll back newer data that
// is already merged into the base.
func TestMergeOwnedIgnoresOlderPart(t *testing.T) {
	root := t.TempDir()
	day1 := time.Date(2026, 10, 26, 3, 0, 0, 0, time.UTC)
	day2 := day1.AddDate(0, 0, 1)
	base := filepath.Join(root, "prev")
	writeScanPart(t, base, day2, map[string]map[string]*store.Entry{"dev": {"bb": free(2)}}, map[string][]string{"dev": {"bbbb"}})
	old := filepath.Join(root, "part-dev")
	writeScanPart(t, old, day1, map[string]map[string]*store.Entry{"dev": {"aa": free(1)}}, map[string][]string{"dev": {"aaaa"}})

	st, bms, covs, err := mergeOwned(base, []string{old}, noJump)
	if err != nil {
		t.Fatal(err)
	}
	if got := entryLabels(st.TLDs["dev"]); !reflect.DeepEqual(got, []string{"bb"}) {
		t.Errorf("dev = %v, want base [bb]", got)
	}
	if sc := st.ScanOf("dev"); sc.At != day2.Unix() || sc.Held != "" {
		t.Errorf("dev scan = %+v, want base scan, not held", sc)
	}
	if got := setBits(bms["dev"]); !reflect.DeepEqual(got, []string{"bbbb"}) {
		t.Errorf("dev bitmap = %v, want base [bbbb]", got)
	}
	if c := covs["dev"]; newestShard(c) != day2.Unix() || c.Held != "" {
		t.Errorf("dev coverage = %+v, want base coverage", c)
	}

	// The same sweep time is not older: a job's second part (after its 4
	// character step) carries the scan record of its first part.
	same := filepath.Join(root, "part-dev-4")
	writeScanPart(t, same, day2, map[string]map[string]*store.Entry{"dev": {"cc": free(2)}}, map[string][]string{"dev": {"cccc"}})
	st, bms, _, err = mergeOwned(base, []string{same}, noJump)
	if err != nil {
		t.Fatal(err)
	}
	if got := entryLabels(st.TLDs["dev"]); !reflect.DeepEqual(got, []string{"cc"}) {
		t.Errorf("dev = %v, want part [cc]", got)
	}
	if got := setBits(bms["dev"]); !reflect.DeepEqual(got, []string{"cccc"}) {
		t.Errorf("dev bitmap = %v, want part [cccc]", got)
	}
}

// mergeInto runs the merge command like the pages workflow: base is the
// published data, the result goes to out/state and out/bitmaps.
func mergeInto(t *testing.T, base, out string, extra []string, parts ...string) {
	t.Helper()
	args := append(extra, "-state", filepath.Join(out, "state", "state.json.gz"), "-bitmaps", filepath.Join(out, "bitmaps"), "-base", base)
	if err := cmdMerge(append(args, parts...)); err != nil {
		t.Fatal(err)
	}
}

// readTree returns every file below dir (state decoded), keyed by relative
// path.
func readTree(t *testing.T, dir string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(dir, path)
		if strings.HasSuffix(path, ".gz") {
			st, err := store.Load(path)
			if err != nil {
				return err
			}
			b, err := json.Marshal(st)
			out[rel] = string(b)
			return err
		}
		b, err := os.ReadFile(path)
		out[rel] = string(b)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// Incremental publishing: each finished TLD is merged into the published
// data on its own (the first run without any base), every later run merges
// the same parts again, and the final merge of all parts gives the same
// result as a single merge would.
func TestCmdMergeIncrementalPublish(t *testing.T) {
	root := t.TempDir()
	day := time.Date(2026, 10, 27, 3, 0, 0, 0, time.UTC)
	dev := filepath.Join(root, "parts", "dev")
	io := filepath.Join(root, "parts", "io")
	writeScanPart(t, dev, day, map[string]map[string]*store.Entry{"dev": {"q7": free(1)}}, nil)

	// Run 1: no data release yet; only .dev is done (1-3 characters).
	mergeInto(t, filepath.Join(root, "none"), filepath.Join(root, "r1"), nil, dev)
	st, err := store.Load(filepath.Join(root, "r1", "state", "state.json.gz"))
	if err != nil {
		t.Fatal(err)
	}
	if len(st.TLDs) != 1 || st.ScanOf("io") != nil {
		t.Errorf("run 1 state = %v, want only .dev (.io stays missing)", st.TLDs)
	}

	// Run 2: .io is done; .dev's part is still in the release.
	writeScanPart(t, io, day.Add(time.Hour), map[string]map[string]*store.Entry{"io": {"x1": free(1)}}, map[string][]string{"io": {"x1x1"}})
	mergeInto(t, filepath.Join(root, "r1"), filepath.Join(root, "r2"), nil, dev, io)

	// Run 3: .dev's 4 character step is done; its part now has a bitmap.
	writeScanPart(t, dev, day, map[string]map[string]*store.Entry{"dev": {"q7": free(1)}}, map[string][]string{"dev": {"q7q7"}})
	mergeInto(t, filepath.Join(root, "r2"), filepath.Join(root, "r3"), nil, dev, io)

	// Run 4 (final) merges the same parts again: nothing changes, and the
	// result equals one merge of all parts.
	mergeInto(t, filepath.Join(root, "r3"), filepath.Join(root, "r4"), nil, dev, io)
	mergeInto(t, filepath.Join(root, "none"), filepath.Join(root, "once"), nil, dev, io)
	r3, r4, once := readTree(t, filepath.Join(root, "r3")), readTree(t, filepath.Join(root, "r4")), readTree(t, filepath.Join(root, "once"))
	if !reflect.DeepEqual(r3, r4) {
		t.Errorf("merging the same parts again changed the data:\n%v\n%v", r3, r4)
	}
	if !reflect.DeepEqual(r4, once) {
		t.Errorf("incremental result differs from a single merge:\n%v\n%v", r4, once)
	}
	if _, ok := r4[filepath.Join("bitmaps", "dev.bin")]; !ok {
		t.Errorf("dev bitmap missing: %v", r4)
	}
}

// The jump safeguard still compares against the data before the run: a
// held TLD stays held (with its old data) when its part is merged again.
func TestMergeOwnedHeldStaysHeldOnRemerge(t *testing.T) {
	root := t.TempDir()
	day1 := time.Date(2026, 10, 26, 3, 0, 0, 0, time.UTC)
	many := map[string]*store.Entry{}
	names.Each(2, 0, 1, func(_ int, s string) {
		if len(many) < 300 {
			many[s] = free(1)
		}
	})
	writeScanPart(t, filepath.Join(root, "prev"), day1, map[string]map[string]*store.Entry{"dev": many}, nil)
	part := filepath.Join(root, "dev")
	writeScanPart(t, part, day1.AddDate(0, 0, 1), map[string]map[string]*store.Entry{"dev": {}}, nil)

	base := filepath.Join(root, "prev")
	for run := 1; run <= 2; run++ {
		out := filepath.Join(root, "r"+strconv.Itoa(run))
		mergeInto(t, base, out, []string{"-max-jump", "0.5", "-min-jump", "200"}, part)
		st, err := store.Load(filepath.Join(out, "state", "state.json.gz"))
		if err != nil {
			t.Fatal(err)
		}
		if sc := st.ScanOf("dev"); len(st.TLDs["dev"]) != 300 || sc.Held == "" || sc.At != day1.Unix() {
			t.Errorf("run %d: dev has %d entries, scan %+v; want the old 300, held", run, len(st.TLDs["dev"]), sc)
		}
		base = out
	}
}
