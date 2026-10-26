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

	"github.com/maximilianfeix/shortfree/internal/names"
	"github.com/maximilianfeix/shortfree/internal/store"
)

func TestStatusOf(t *testing.T) {
	now := time.Date(2026, 10, 26, 3, 0, 0, 0, time.UTC)
	recent := now.Add(-24 * time.Hour).Unix()
	old := now.Add(-15 * 24 * time.Hour).Unix() // older than the 14 day limit
	cases := []struct {
		e            store.Entry
		want, verify string
	}{
		{store.Entry{DNS: "free", RDAP: "available", Checked: recent}, "available", "rdap_verified"},
		{store.Entry{DNS: "free", RDAP: "dropping", Checked: recent}, "dropping", "rdap_verified"},
		{store.Entry{DNS: "free"}, "unverified", "pending"},
		{store.Entry{DNS: "free", RDAP: "error", Checked: recent}, "unverified", "pending"},
		{store.Entry{DNS: "unknown"}, "unknown", "pending"},
		{store.Entry{DNS: "free", RDAP: "registered", Checked: recent}, "", "rdap_verified"},
		// Old evidence is never published as verified.
		{store.Entry{DNS: "free", RDAP: "available", Checked: old}, "unverified", "stale"},
		{store.Entry{DNS: "free", RDAP: "dropping", Checked: old}, "unverified", "stale"},
		{store.Entry{DNS: "unknown", RDAP: "available", Checked: old}, "unknown", "stale"},
		{store.Entry{DNS: "unknown", RDAP: "available", Checked: recent}, "available", "rdap_verified"},
		{store.Entry{DNS: "free", RDAP: "registered", Checked: old}, "", "stale"},
		// An RDAP status without a check time is no evidence.
		{store.Entry{DNS: "free", RDAP: "available"}, "unverified", "stale"},
	}
	for _, c := range cases {
		if got := statusOf(&c.e, now, defaultMaxAge); got != c.want {
			t.Errorf("statusOf(%+v) = %q, want %q", c.e, got, c.want)
		}
		if got := verificationOf(&c.e, now, defaultMaxAge); got != c.verify {
			t.Errorf("verificationOf(%+v) = %q, want %q", c.e, got, c.verify)
		}
	}
	// maxAge <= 0 disables expiry.
	e := store.Entry{DNS: "free", RDAP: "available", Checked: old}
	if got := statusOf(&e, now, 0); got != "available" {
		t.Errorf("statusOf without age limit = %q", got)
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
	st.TLD("dev")["x"] = &store.Entry{DNS: "free", RDAP: "dropping", Expires: day("2026-11-05"), Checked: day("2026-10-25"), Phase: "redemption"}
	st.TLD("dev")["z9z"] = &store.Entry{DNS: "free"}
	st.TLD("dev")["zz"] = &store.Entry{DNS: "free", RDAP: "registered", Checked: day("2026-10-25")}
	st.TLD("dev")["old"] = &store.Entry{DNS: "free", RDAP: "available", Since: day("2026-09-01"), Checked: day("2026-10-01")}
	st.TLD("dev")["u7"] = &store.Entry{DNS: "unknown"}
	st.TLD("io")["k-1"] = &store.Entry{DNS: "free", RDAP: "available", Since: day("2026-10-20"), Checked: day("2026-10-20")}
	st.TLD("io")["ab"] = &store.Entry{DNS: "free", RDAP: "dropping", Expires: day("2026-11-01"), Checked: day("2026-10-25"), Phase: "pending_delete"}
	statePath := filepath.Join(dir, "state.json.gz")
	if err := st.Save(statePath); err != nil {
		t.Fatal(err)
	}
	// .io has a 4 character bitmap with one swept shard; .dev has none.
	bmDir := filepath.Join(dir, "bitmaps")
	bm := store.NewBitmap(names.Space(4))
	bm.Set(names.Index("abcd"), true)
	if err := store.WriteAtomic(filepath.Join(bmDir, "io.bin"), bm); err != nil {
		t.Fatal(err)
	}
	cov := &store.Coverage{Shards: 7, Scans: map[string]*store.ShardScan{}}
	cov.Record(0, 7, &store.ShardScan{At: day("2026-10-25"), Names: shardSize(0, 7), Unknown: 3})
	if err := cov.Save(store.CoveragePath(filepath.Join(bmDir, "io.bin"))); err != nil {
		t.Fatal(err)
	}
	api := filepath.Join(dir, "api")
	err := cmdExport([]string{
		"-tlds", "dev,io",
		"-state", statePath,
		"-bitmaps", bmDir,
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
			if l := s.TLDs["dev"]["2"]; l.Available != 1 || l.Unknown != 1 || s.TLDs["dev"]["1"].Dropping != 1 || s.TLDs["dev"]["3"].Unverified != 2 || s.TLDs["dev"]["3"].Stale != 1 {
				t.Errorf("stats dev = %+v", s.TLDs["dev"])
			}
			if !s.TLDs["dev"]["4"].Missing || s.TLDs["dev"]["4"].Free != 0 {
				t.Errorf("dev without bitmap: 4 = %+v, want missing", s.TLDs["dev"]["4"])
			}
			io4 := s.TLDs["io"]["4"]
			if io4.Free != 1 || io4.Unknown != 3 || io4.Unscanned != io4.Valid-shardSize(0, 7) || io4.Missing {
				t.Errorf("io 4 = %+v", io4)
			}
			if c := s.Coverage["dev"]; c == nil || c.Status != covPartial || c.Lengths["3"].RDAPStale != 1 || c.Lengths["3"].RDAPPending != 1 || c.Lengths["2"].DNSUnknown != 1 {
				t.Errorf("dev coverage = %+v", c)
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
				t.Errorf("dropping = %s (want ab.io before x.dev: pending delete first)", b)
			}
			if d.Domains[0].Phase != "pending_delete" || d.Domains[0].RDAPExpiration.IsZero() || !d.Domains[0].Expires.Equal(d.Domains[0].RDAPExpiration) {
				t.Errorf("dropping fields = %+v", d.Domains[0])
			}
		}},
		{"dev.json", func(t *testing.T, b []byte) {
			var l APIList
			mustJSON(t, b, &l)
			want := TLDList{Available: []string{"q7", "abc"}, Dropping: l.Dropping, Unverified: []string{"old", "z9z"}, Stale: []string{"old"}, Unknown: []string{"u7"}, Checked: l.Checked}
			if l.Checked["q7"] != time.Date(2026, 10, 25, 0, 0, 0, 0, time.UTC).Unix() || len(l.Checked) != 2 {
				t.Errorf("checked = %v", l.Checked)
			}
			if l.TLD != "dev" || !reflect.DeepEqual(l.TLDList, want) || len(l.Dropping) != 1 {
				t.Errorf("dev.json = %s", b)
			}
			if l.MaxEvidenceAgeDays != 14 || l.OldestCheck == nil || l.OldestCheck.Format("2006-01-02") != "2026-10-24" {
				t.Errorf("dev.json evidence = %v %v", l.MaxEvidenceAgeDays, l.OldestCheck)
			}
		}},
		{"dev/2.json", func(t *testing.T, b []byte) {
			var l APIList
			mustJSON(t, b, &l)
			if l.Length != 2 || !reflect.DeepEqual(l.Available, []string{"q7"}) || len(l.Dropping) != 0 || len(l.Unverified) != 0 || !reflect.DeepEqual(l.Unknown, []string{"u7"}) {
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
		{"index.json", func(t *testing.T, b []byte) {
			var idx APIIndex
			mustJSON(t, b, &idx)
			if idx.MaxEvidenceAgeDays != 14 || !strings.Contains(idx.Statuses["available"], "Not a promise") || idx.Accuracy == "" {
				t.Errorf("index statuses = %+v", idx)
			}
		}},
		{"coverage.json", func(t *testing.T, b []byte) {
			var c CoverageReport
			mustJSON(t, b, &c)
			io := c.TLDs["io"]
			if io == nil || len(io.Lengths["4"].Shards) != 7 || io.Lengths["4"].Shards[0].ScannedAt == nil || io.Lengths["4"].Shards[1].ScannedAt != nil {
				t.Errorf("coverage io = %s", b)
			}
			if c.MaxEvidenceAgeDays != 14 || !strings.Contains(c.Scope, "ASCII") || c.Missing == nil {
				t.Errorf("coverage = %s", b)
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
			"q7.dev,dev,2,mixed,available,2026-10-01,2026-10-25T00:00:00Z,rdap_verified,1",
			"abc.dev,dev,3,letters,available,2026-09-30,2026-10-24T00:00:00Z,rdap_verified,2",
			"k-1.io,io,3,mixed,available,2026-10-20,2026-10-20T00:00:00Z,rdap_verified,6",
		}},
		{"dev.csv", []string{
			header,
			"x.dev,dev,1,letters,dropping,,2026-10-25T00:00:00Z,rdap_verified,1",
			"q7.dev,dev,2,mixed,available,2026-10-01,2026-10-25T00:00:00Z,rdap_verified,1",
			"u7.dev,dev,2,mixed,unknown,,,pending,",
			"abc.dev,dev,3,letters,available,2026-09-30,2026-10-24T00:00:00Z,rdap_verified,2",
			"old.dev,dev,3,letters,unverified,2026-09-01,2026-10-01T00:00:00Z,stale,25",
			"z9z.dev,dev,3,mixed,unverified,,,pending,",
		}},
		{"io.csv", []string{
			header,
			"ab.io,io,2,letters,dropping,,2026-10-25T00:00:00Z,rdap_verified,1",
			"k-1.io,io,3,mixed,available,2026-10-20,2026-10-20T00:00:00Z,rdap_verified,6",
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
	st := &store.State{Updated: time.Now().UTC(), TLDs: map[string]map[string]*store.Entry{"dev": {}}}
	if err := st.Save(filepath.Join(dir, "state.json")); err != nil {
		t.Fatal(err)
	}
	err := cmdExport([]string{
		"-tlds", "dev", "-state", filepath.Join(dir, "state.json"),
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

// A TLD without any scan data (e.g. a new TLD whose first job failed) must
// be left out, not published as "0 free".
func TestExportLeavesOutUnscannedTLDs(t *testing.T) {
	dir := t.TempDir()
	st := &store.State{Updated: time.Now().UTC(), TLDs: map[string]map[string]*store.Entry{}}
	st.SetScan("dev", st.Updated)
	statePath := filepath.Join(dir, "state.json")
	if err := st.Save(statePath); err != nil {
		t.Fatal(err)
	}
	args := func(tlds string) []string {
		return []string{"-tlds", tlds, "-state", statePath, "-bitmaps", dir, "-out", filepath.Join(dir, "data"),
			"-four-out", "", "-readme", "", "-api-out", filepath.Join(dir, "api")}
	}
	if err := cmdExport(args("dev,io")); err != nil {
		t.Fatal(err)
	}
	var s Stats
	b, _ := os.ReadFile(filepath.Join(dir, "data", "stats.json"))
	mustJSON(t, b, &s)
	if !reflect.DeepEqual(s.Order, []string{"dev"}) || !reflect.DeepEqual(s.Missing, []string{"io"}) || s.TLDs["io"] != nil {
		t.Errorf("stats = %s", b)
	}
	if _, err := os.Stat(filepath.Join(dir, "api", "v1", "io.json")); !os.IsNotExist(err) {
		t.Error("io.json written for a TLD without data")
	}
	if err := cmdExport(args("io")); err == nil {
		t.Error("export of only unscanned TLDs must fail")
	}
}

func mustJSON(t *testing.T, b []byte, v any) {
	t.Helper()
	if err := json.Unmarshal(b, v); err != nil {
		t.Fatalf("%v: %s", err, b)
	}
}
