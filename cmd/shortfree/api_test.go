package main

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/maximilianfeix/shortfree/internal/store"
)

func TestStatusOf(t *testing.T) {
	cases := []struct {
		e    store.Entry
		want string
	}{
		{store.Entry{DNS: "free", RDAP: "available"}, "available"},
		{store.Entry{DNS: "free", RDAP: "dropping"}, "dropping"},
		{store.Entry{DNS: "free"}, "unverified"},
		{store.Entry{DNS: "free", RDAP: "error"}, "unverified"},
		{store.Entry{DNS: "unknown"}, ""},
		{store.Entry{DNS: "free", RDAP: "registered"}, ""},
	}
	for _, c := range cases {
		if got := statusOf(&c.e); got != c.want {
			t.Errorf("statusOf(%+v) = %q, want %q", c.e, got, c.want)
		}
	}
}

// exportFixture runs a full export of a small state into temp dirs and
// returns the API directory.
func exportFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	day := func(s string) int64 {
		tm, err := time.Parse("2006-01-02", s)
		if err != nil {
			t.Fatal(err)
		}
		return tm.Unix()
	}
	st := &store.State{Updated: time.Date(2026, 10, 26, 3, 0, 0, 0, time.UTC), TLDs: map[string]map[string]*store.Entry{}}
	st.TLD("dev")["q7"] = &store.Entry{DNS: "free", RDAP: "available", Since: day("2026-10-01"), Checked: day("2026-10-25")}
	st.TLD("dev")["abc"] = &store.Entry{DNS: "free", RDAP: "available", Since: day("2026-09-30"), Checked: day("2026-10-24")}
	st.TLD("dev")["x"] = &store.Entry{DNS: "free", RDAP: "dropping", Expires: day("2026-11-05")}
	st.TLD("dev")["z9z"] = &store.Entry{DNS: "free"}
	st.TLD("dev")["zz"] = &store.Entry{DNS: "free", RDAP: "registered"}
	st.TLD("io")["k-1"] = &store.Entry{DNS: "free", RDAP: "available", Since: day("2026-10-20")}
	st.TLD("io")["ab"] = &store.Entry{DNS: "free", RDAP: "dropping", Expires: day("2026-11-01")}
	statePath := filepath.Join(dir, "state.json.gz")
	if err := st.Save(statePath); err != nil {
		t.Fatal(err)
	}
	api := filepath.Join(dir, "api")
	err := cmdExport([]string{
		"-tlds", "dev,io",
		"-state", statePath,
		"-bitmaps", filepath.Join(dir, "bitmaps"),
		"-out", filepath.Join(dir, "data"),
		"-four-out", "",
		"-readme", "",
		"-api-out", api,
	})
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(api, "v1")
}

func TestExportAPIJSON(t *testing.T) {
	v1 := exportFixture(t)
	cases := []struct {
		file  string
		check func(t *testing.T, b []byte)
	}{
		{"index.json", func(t *testing.T, b []byte) {
			var idx APIIndex
			mustJSON(t, b, &idx)
			if !reflect.DeepEqual(idx.TLDs, []string{"dev", "io"}) || len(idx.Endpoints) == 0 || idx.Updated.IsZero() {
				t.Errorf("index = %+v", idx)
			}
		}},
		{"stats.json", func(t *testing.T, b []byte) {
			var s Stats
			mustJSON(t, b, &s)
			if l := s.TLDs["dev"]["2"]; l.Available != 1 || s.TLDs["dev"]["1"].Dropping != 1 || s.TLDs["dev"]["3"].Unverified != 1 {
				t.Errorf("stats dev = %+v", s.TLDs["dev"])
			}
		}},
		{"changes.json", func(t *testing.T, b []byte) {
			var c APIChanges
			mustJSON(t, b, &c)
			if c.Changes == nil || c.Days != 30 {
				t.Errorf("changes = %s", b)
			}
		}},
		{"dropping.json", func(t *testing.T, b []byte) {
			var d APIDroppingList
			mustJSON(t, b, &d)
			if d.Count != 2 || d.Domains[0].Domain != "ab.io" || d.Domains[1].Domain != "x.dev" {
				t.Errorf("dropping = %s (want ab.io before x.dev, soonest expiry first)", b)
			}
		}},
		{"dev.json", func(t *testing.T, b []byte) {
			var l APIList
			mustJSON(t, b, &l)
			want := TLDList{Available: []string{"q7", "abc"}, Dropping: l.Dropping, Unverified: []string{"z9z"}}
			if l.TLD != "dev" || !reflect.DeepEqual(l.TLDList, want) || len(l.Dropping) != 1 {
				t.Errorf("dev.json = %s", b)
			}
		}},
		{"dev/2.json", func(t *testing.T, b []byte) {
			var l APIList
			mustJSON(t, b, &l)
			if l.Length != 2 || !reflect.DeepEqual(l.Available, []string{"q7"}) || len(l.Dropping) != 0 || len(l.Unverified) != 0 {
				t.Errorf("dev/2.json = %s", b)
			}
			if !strings.Contains(string(b), `"dropping":[]`) {
				t.Errorf("empty lists must be [] not null: %s", b)
			}
		}},
		{"badge.json", func(t *testing.T, b []byte) {
			want := `{"schemaVersion":1,"label":"free short domains","message":"3 verified","color":"1FC16B"}` + "\n"
			if string(b) != want {
				t.Errorf("badge.json = %s, want %s", b, want)
			}
		}},
		{"io/3.json", func(t *testing.T, b []byte) {
			var l APIList
			mustJSON(t, b, &l)
			if !reflect.DeepEqual(l.Available, []string{"k-1"}) {
				t.Errorf("io/3.json = %s", b)
			}
		}},
	}
	for _, c := range cases {
		t.Run(c.file, func(t *testing.T) {
			b, err := os.ReadFile(filepath.Join(v1, c.file))
			if err != nil {
				t.Fatal(err)
			}
			if strings.Count(string(b), "\n") != 1 {
				t.Errorf("%s is not compact JSON", c.file)
			}
			c.check(t, b)
		})
	}
}

func TestBadgeOf(t *testing.T) {
	if got := badgeOf(12345).Message; got != "12,345 verified" {
		t.Errorf("badgeOf(12345).Message = %q", got)
	}
	if got := badgeOf(0).Message; got != "0 verified" {
		t.Errorf("badgeOf(0).Message = %q", got)
	}
}

func TestExportAPICSV(t *testing.T) {
	v1 := exportFixture(t)
	header := strings.Join(csvHeader, ",")
	cases := []struct {
		file string
		want []string
	}{
		{"free.csv", []string{
			header,
			"q7.dev,dev,2,mixed,available,2026-10-01,2026-10-25T00:00:00Z",
			"abc.dev,dev,3,letters,available,2026-09-30,2026-10-24T00:00:00Z",
			"k-1.io,io,3,mixed,available,2026-10-20,",
		}},
		{"dev.csv", []string{
			header,
			"x.dev,dev,1,letters,dropping,,",
			"q7.dev,dev,2,mixed,available,2026-10-01,2026-10-25T00:00:00Z",
			"abc.dev,dev,3,letters,available,2026-09-30,2026-10-24T00:00:00Z",
			"z9z.dev,dev,3,mixed,unverified,,",
		}},
		{"io.csv", []string{
			header,
			"ab.io,io,2,letters,dropping,,",
			"k-1.io,io,3,mixed,available,2026-10-20,",
		}},
	}
	for _, c := range cases {
		t.Run(c.file, func(t *testing.T) {
			f, err := os.Open(filepath.Join(v1, c.file))
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			recs, err := csv.NewReader(f).ReadAll()
			if err != nil {
				t.Fatal(err)
			}
			var got []string
			for _, r := range recs {
				got = append(got, strings.Join(r, ","))
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("%s:\n got %q\nwant %q", c.file, got, c.want)
			}
		})
	}
}

func TestExportAPISkipped(t *testing.T) {
	dir := t.TempDir()
	err := cmdExport([]string{
		"-tlds", "dev", "-state", filepath.Join(dir, "none.json"),
		"-bitmaps", dir, "-out", filepath.Join(dir, "data"),
		"-four-out", "", "-readme", "", "-api-out", "",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "api")); !os.IsNotExist(err) {
		t.Fatal("api written although -api-out is empty")
	}
}

func mustJSON(t *testing.T, b []byte, v any) {
	t.Helper()
	if err := json.Unmarshal(b, v); err != nil {
		t.Fatalf("%v: %s", err, b)
	}
}
