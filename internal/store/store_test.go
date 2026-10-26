package store

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
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
