package store

import (
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
	if c, _ := LoadBitmap(path, 1000); c.Count() != 0 {
		t.Fatal("size mismatch must yield an empty bitmap")
	}
}
