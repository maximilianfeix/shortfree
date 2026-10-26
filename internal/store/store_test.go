package store

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/maximilianfeix/shortfree/internal/names"
)

func TestStateRoundTrip(t *testing.T) {
	for _, name := range []string{"state.json", "state.json.gz"} {
		path := filepath.Join(t.TempDir(), name)
		s, err := Load(path)
		if err != nil {
			t.Fatal(err)
		}
		s.Updated = time.Date(2026, 10, 26, 20, 0, 0, 0, time.UTC)
		s.TLD("tools")["q7"] = &Entry{DNS: "free", RDAP: "available", Checked: 1, Since: 2}
		s.TLD("tools")["a"] = &Entry{DNS: "free"}
		if err := s.Save(path); err != nil {
			t.Fatal(err)
		}
		got, err := Load(path)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, s) {
			t.Fatalf("%s: round trip mismatch:\n got %+v\nwant %+v", name, got, s)
		}
		if l := got.Labels("tools"); !reflect.DeepEqual(l, []string{"a", "q7"}) {
			t.Fatalf("Labels = %v", l)
		}
	}
}

func TestBitmap(t *testing.T) {
	path := filepath.Join(t.TempDir(), "x.bin")
	b, _ := LoadBitmap(path, 100)
	for _, i := range []int{0, 7, 8, 63, 99} {
		b.Set(i, true)
	}
	b.Set(7, false)
	if b.Count() != 4 || b.Get(7) || !b.Get(99) {
		t.Fatalf("bitmap state wrong: count %d", b.Count())
	}
	if err := WriteAtomic(path, b); err != nil {
		t.Fatal(err)
	}
	if c, _ := LoadBitmap(path, 100); c.Count() != 4 {
		t.Fatal("bitmap not persisted")
	}
	// A corrupt bitmap must not silently become an empty one (every name
	// would turn "taken").
	if _, err := LoadBitmap(path, 1000); err == nil {
		t.Fatal("size mismatch must be an error")
	}
	if _, err := ReadBitmap(filepath.Join(t.TempDir(), "none.bin"), 100); !os.IsNotExist(err) {
		t.Fatalf("ReadBitmap of a missing file: %v, want not-exist", err)
	}
}

func TestCoverage(t *testing.T) {
	path := CoveragePath(filepath.Join(t.TempDir(), "dev.bin"))
	if filepath.Base(path) != "dev.cov.json" {
		t.Fatalf("CoveragePath = %s", path)
	}
	c, err := LoadCoverage(path, 7)
	if err != nil || c.Complete() || c.Scanned(0) != nil {
		t.Fatalf("missing coverage: %+v, %v", c, err)
	}
	for i := 0; i < 7; i++ {
		c.Record(i, 7, &ShardScan{At: int64(i), Names: 10})
	}
	if !c.Complete() {
		t.Fatal("all shards recorded, want complete")
	}
	if err := c.Save(path); err != nil {
		t.Fatal(err)
	}
	got, err := LoadCoverage(path, 7)
	if err != nil || !got.Complete() || got.Scanned(3).At != 3 {
		t.Fatalf("round trip: %+v, %v", got, err)
	}
	got.Record(0, 5, &ShardScan{At: 9})
	if got.Complete() || got.Scanned(1) != nil || got.Shards != 5 {
		t.Fatalf("new shard count must reset coverage: %+v", got)
	}
}

// A bitmap published before coverage files existed was swept completely by
// every run: it counts as fully swept at the state's update time.
func TestBitmapCoverageLegacy(t *testing.T) {
	dir := t.TempDir()
	at := time.Date(2026, 10, 26, 3, 0, 0, 0, time.UTC)
	bm := filepath.Join(dir, "dev.bin")
	// No bitmap: nothing swept.
	if c, err := BitmapCoverage(bm, 7, at); err != nil || c.Complete() || c.LegacyAt != 0 {
		t.Fatalf("no bitmap: %+v, %v", c, err)
	}
	if err := WriteAtomic(bm, NewBitmap(100)); err != nil {
		t.Fatal(err)
	}
	c, err := BitmapCoverage(bm, 7, at)
	if err != nil || !c.Complete() || c.LegacyAt != at.Unix() || c.Scanned(6).At != at.Unix() {
		t.Fatalf("legacy bitmap: %+v, %v", c, err)
	}
	total := 0
	for i := 0; i < 7; i++ {
		total += c.Scanned(i).Names
	}
	want := 0
	names.Each(4, 0, 1, func(int, string) { want++ })
	if total != want { // every valid 4 character label, once
		t.Errorf("legacy shard names sum to %d, want %d", total, want)
	}
	// Without a time to date it, a legacy bitmap is not assumed swept.
	if c, _ := BitmapCoverage(bm, 7, time.Time{}); c.Complete() {
		t.Error("zero legacyAt must not assume coverage")
	}
	// A coverage file wins over the assumption.
	real := &Coverage{Shards: 7, Scans: map[string]*ShardScan{}}
	real.Record(2, 7, &ShardScan{At: 5})
	if err := real.Save(CoveragePath(bm)); err != nil {
		t.Fatal(err)
	}
	if c, _ := BitmapCoverage(bm, 7, at); c.Complete() || c.LegacyAt != 0 {
		t.Errorf("coverage file ignored: %+v", c)
	}
	// Re-sweeping every shard ends the assumption.
	c = LegacyCoverage(7, at)
	for i := 0; i < 7; i++ {
		if c.LegacyAt == 0 {
			t.Fatalf("assumption ended after %d shards", i)
		}
		c.Record(i, 7, &ShardScan{At: at.Unix() + 86400})
	}
	if c.LegacyAt != 0 {
		t.Error("assumption kept after every shard was swept")
	}
}

func TestScanOf(t *testing.T) {
	s := &State{Updated: time.Unix(100, 0), TLDs: map[string]map[string]*Entry{"dev": {}}}
	if sc := s.ScanOf("dev"); sc == nil || sc.At != 100 {
		t.Fatalf("legacy TLD: %+v, want scanned at Updated", sc)
	}
	if s.ScanOf("io") != nil {
		t.Fatal("unknown TLD must have no scan")
	}
	s.SetScan("io", time.Unix(200, 0))
	if sc := s.ScanOf("io"); sc == nil || sc.At != 200 {
		t.Fatalf("SetScan: %+v", sc)
	}
}
