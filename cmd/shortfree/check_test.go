package main

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestWriteCheckJSON(t *testing.T) {
	res := fakeChecker().checkAll(context.Background(), []string{"free.dev", "Taken.COM", "gone.io", "nope"})
	var buf bytes.Buffer
	if err := writeCheck(&buf, res, true); err != nil {
		t.Fatal(err)
	}
	var got []checkResult
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("not a JSON array: %v\n%s", err, buf.Bytes())
	}
	want := []struct {
		domain, verdict string
		free            bool
	}{
		{"free.dev", vNotRegistered, true},
		{"taken.com", "taken", false},
		{"gone.io", vDropping, false},
		{"nope", "invalid (want label.tld with a-z, 0-9 and inner hyphens)", false},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d results, want %d", len(got), len(want))
	}
	for i, w := range want {
		if got[i].Domain != w.domain || got[i].Verdict != w.verdict || got[i].Free != w.free {
			t.Errorf("result %d = %+v, want %s %q free=%v", i, got[i], w.domain, w.verdict, w.free)
		}
	}
	if got[2].Expires == nil {
		t.Error("dropping result lost its expiry date")
	}
}

func TestWriteCheckJSONEmpty(t *testing.T) {
	var buf bytes.Buffer
	if err := writeCheck(&buf, []checkResult{}, true); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(buf.String()) != "[]" {
		t.Errorf("empty result = %q, want []", buf.String())
	}
}

func TestWriteCheckTable(t *testing.T) {
	res := fakeChecker().checkAll(context.Background(), []string{"free.dev", "taken.com"})
	var buf bytes.Buffer
	if err := writeCheck(&buf, res, false); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 || !strings.HasPrefix(lines[0], "DOMAIN") ||
		!strings.Contains(lines[1], vNotRegistered) || !strings.Contains(lines[2], "taken") {
		t.Errorf("table:\n%s", buf.String())
	}
}
