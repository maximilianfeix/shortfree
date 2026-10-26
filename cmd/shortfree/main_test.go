package main

import (
	"os"
	"path/filepath"
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
		{probe.Free, rdap.Available, "FREE"},
		{probe.Free, rdap.Dropping, "dropping soon"},
		{probe.Free, rdap.Registered, "taken"},
		{probe.Free, rdap.Error, "probably free"},
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
	log, err := diffChanges(filepath.Join(dir, "available.json"), filepath.Join(dir, "changes.json"), cur, 30)
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
	prev := &Available{TLDs: map[string]*TLDList{"dev": {Available: []string{"aa", "bb"}, Unverified: []string{}}}}
	if err := writeJSON(filepath.Join(dir, "available.json"), prev); err != nil {
		t.Fatal(err)
	}
	cur := &Available{
		Updated: time.Date(2026, 10, 27, 3, 0, 0, 0, time.UTC),
		TLDs:    map[string]*TLDList{"dev": {Available: []string{"bb", "cc"}, Unverified: []string{}}},
	}
	log, err := diffChanges(filepath.Join(dir, "available.json"), filepath.Join(dir, "changes.json"), cur, 30)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for _, c := range log {
		got[c.Domain] = c.Event
	}
	if got["cc.dev"] != "freed" || got["aa.dev"] != "taken" || len(got) != 2 {
		t.Fatalf("changes = %+v", log)
	}
}
