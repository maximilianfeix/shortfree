package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/maximilianfeix/shortfree/internal/probe"
	"github.com/maximilianfeix/shortfree/internal/rdap"
)

func TestVerdict(t *testing.T) {
	cases := []struct {
		dns  probe.Result
		r    rdap.Status
		want string
	}{
		{probe.Free, rdap.Available, vNotRegistered},
		{probe.Free, rdap.Dropping, vDropping},
		{probe.Free, rdap.Registered, "taken"},
		{probe.Free, rdap.Error, "not in DNS"},
		{probe.Taken, rdap.Error, "taken"},
		{probe.Unknown, rdap.Error, "unknown"},
	}
	for _, c := range cases {
		if got := verdictOf(c.dns, c.r); !strings.HasPrefix(got, c.want) {
			t.Errorf("verdictOf(%s, %s) = %q, want %q…", c.dns, c.r, got, c.want)
		}
	}
}

func TestFirstExportWritesEmptyChangeList(t *testing.T) {
	dir := t.TempDir()
	cur := &Available{Updated: time.Now(), TLDs: map[string]*TLDList{"dev": {Available: []string{"q7"}}}}
	log, err := diffChanges(filepath.Join(dir, "available.json"), filepath.Join(dir, "changes.json"), cur, 30, nil)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "changes.json")
	if err := writeJSON(path, log); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(path)
	if strings.TrimSpace(string(b)) != "[]" {
		t.Fatalf("changes.json = %s, want []", b)
	}
}

func TestDiffChanges(t *testing.T) {
	dir := t.TempDir()
	prevTime := time.Date(2026, 10, 26, 3, 0, 0, 0, time.UTC)
	prev := &Available{Updated: prevTime, TLDs: map[string]*TLDList{"dev": {
		Available:  []string{"aa", "bb", "st", "un"},
		Dropping:   []Dropping{{Name: "dr"}},
		Unverified: []string{"qq"},
	}}}
	if err := writeJSON(filepath.Join(dir, "available.json"), prev); err != nil {
		t.Fatal(err)
	}
	cur := &Available{
		Updated: prevTime.AddDate(0, 0, 1),
		TLDs: map[string]*TLDList{"dev": {
			// cc: RDAP registered -> 404 observed today; qq: queued name,
			// first RDAP check; dr: was dropping; ee: first seen today.
			Available:  []string{"bb", "cc", "qq", "dr", "ee"},
			Unverified: []string{"st"}, // evidence expired: not "taken"
			Unknown:    []string{"un"}, // no DNS answer: not "taken"
		}},
	}
	freedSince := func(d string, since time.Time) bool { return d == "cc.dev" && since.Equal(prevTime) }
	log, err := diffChanges(filepath.Join(dir, "available.json"), filepath.Join(dir, "changes.json"), cur, 30, freedSince)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for _, c := range log {
		got[c.Domain] = c.Event
	}
	want := map[string]string{"cc.dev": "freed", "dr.dev": "freed", "qq.dev": "newly_listed", "ee.dev": "newly_listed", "aa.dev": "taken"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("changes = %v, want %v", got, want)
	}
}
