package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/maximilianfeix/shortfree/internal/names"
	"github.com/maximilianfeix/shortfree/internal/probe"
	"github.com/maximilianfeix/shortfree/internal/rdap"
	"github.com/maximilianfeix/shortfree/internal/store"
)

// A registry that claims every name is registered (negative control fails)
// must not mark queued names as registered.
func TestVerifySkipsRegistryFailingNegativeCanary(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"objectClassName":"domain","status":["active"]}`))
	}))
	t.Cleanup(srv.Close)
	rdap.Endpoints["test"] = srv.URL + "/"
	t.Cleanup(func() { delete(rdap.Endpoints, "test") })
	st := &store.State{TLDs: map[string]map[string]*store.Entry{"test": {"q7": {DNS: "free"}}}}
	verify(context.Background(), st, []string{"test"}, 1000, map[string]time.Duration{"test": time.Minute}, 24*time.Hour, time.Now())
	if e := st.TLDs["test"]["q7"]; e.RDAP != "" {
		t.Errorf("q7 = %+v, want untouched", e)
	}
}

// An endpoint whose 404s are HTML pages (moved service) must not verify
// anything: the negative canary fails with a bad-response error.
func TestVerifySkipsHTMLNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/nic.test") {
			w.Write([]byte(`{"status":["active"]}`))
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("<h1>Not Found</h1>"))
	}))
	t.Cleanup(srv.Close)
	rdap.Endpoints["test"] = srv.URL + "/"
	t.Cleanup(func() { delete(rdap.Endpoints, "test") })
	st := &store.State{TLDs: map[string]map[string]*store.Entry{"test": {"q7": {DNS: "free"}}}}
	verify(context.Background(), st, []string{"test"}, 1000, map[string]time.Duration{"test": time.Minute}, 24*time.Hour, time.Now())
	if e := st.TLDs["test"]["q7"]; e.RDAP != "" {
		t.Errorf("q7 = %+v, want untouched", e)
	}
}

func TestApplyRDAPRecordsObservedFreeing(t *testing.T) {
	now := time.Date(2026, 10, 26, 3, 0, 0, 0, time.UTC)
	cases := []struct {
		before    string
		after     rdap.Info
		wantFreed bool
	}{
		{"registered", rdap.Info{Status: rdap.Available}, true},
		{"dropping", rdap.Info{Status: rdap.Available}, true},
		{"", rdap.Info{Status: rdap.Available}, false},          // first check: newly listed, not freed
		{"error", rdap.Info{Status: rdap.Available}, false},     // no registration was ever seen
		{"available", rdap.Info{Status: rdap.Available}, false}, // recheck
	}
	for _, c := range cases {
		e := &store.Entry{DNS: "free", RDAP: c.before}
		applyRDAP(e, c.after, now)
		if (e.Freed == now.Unix()) != c.wantFreed || e.Checked != now.Unix() || e.RDAP != string(c.after.Status) {
			t.Errorf("%q -> %s: %+v, want freed=%v", c.before, c.after.Status, e, c.wantFreed)
		}
	}
	e := &store.Entry{DNS: "free"}
	applyRDAP(e, rdap.Info{Status: rdap.Dropping, Phase: rdap.PhaseRedemption, Expires: now}, now)
	if e.Phase != rdap.PhaseRedemption || e.Expires != now.Unix() {
		t.Errorf("dropping entry = %+v", e)
	}
}

// End to end through verify: a registered name that answers 404 today is
// recorded as freed.
func TestVerifyRecordsFreed(t *testing.T) {
	fakeRDAP(t, true)
	now := time.Now().UTC().Truncate(time.Second)
	old := now.Add(-30 * 24 * time.Hour).Unix()
	st := &store.State{TLDs: map[string]map[string]*store.Entry{"test": {
		"q7": {DNS: "free", RDAP: "registered", Checked: old},
	}}}
	verify(context.Background(), st, []string{"test"}, 1000, map[string]time.Duration{"test": time.Minute}, 24*time.Hour, now)
	if e := st.TLDs["test"]["q7"]; e.RDAP != "available" || e.Freed != now.Unix() {
		t.Errorf("q7 = %+v, want available and freed now", e)
	}
}

func TestSweepResult(t *testing.T) {
	now := time.Unix(1000, 0)
	m := map[string]*store.Entry{
		"gone": {DNS: "free", RDAP: "available", Checked: 5},
		"kept": {DNS: "free", RDAP: "available", Checked: 5, Since: 7},
	}
	sweepResult(m, "gone", probe.Taken, now)
	sweepResult(m, "kept", probe.Unknown, now)
	sweepResult(m, "new", probe.Free, now)
	sweepResult(m, "flaky", probe.Unknown, now)
	if _, ok := m["gone"]; ok {
		t.Error("delegated name kept")
	}
	if e := m["kept"]; e.DNS != "unknown" || e.RDAP != "available" || e.Since != 7 {
		t.Errorf("kept = %+v: a DNS timeout must keep earlier RDAP evidence", e)
	}
	if e := m["new"]; e.DNS != "free" || e.Since != 1000 {
		t.Errorf("new = %+v", e)
	}
	// Without an answer the name is recorded as unknown, never dropped as taken.
	if e := m["flaky"]; e == nil || e.DNS != "unknown" || e.Since != 0 {
		t.Errorf("flaky = %+v", e)
	}
	sweepResult(m, "flaky", probe.Free, now)
	if e := m["flaky"]; e.DNS != "free" || e.Since != 1000 {
		t.Errorf("flaky after answer = %+v", e)
	}
}

func TestPrioritizeReservesOverdue(t *testing.T) {
	type j struct {
		label string
		e     *store.Entry
	}
	var jobs []j
	for i := 0; i < 100; i++ {
		jobs = append(jobs, j{fmt.Sprintf("n%02d", i), &store.Entry{DNS: "free"}}) // new names
	}
	for i := 0; i < 50; i++ {
		jobs = append(jobs, j{fmt.Sprintf("o%02d", i), &store.Entry{DNS: "free", RDAP: "available", Checked: int64(i)}})
	}
	key := func(x j) (*store.Entry, string, string) { return x.e, x.label, x.label }
	got := prioritize(append([]j{}, jobs...), 40, key)
	overdue := 0
	for _, x := range got {
		if x.e.RDAP == "available" {
			overdue++
		}
	}
	if len(got) != 40 || overdue != 10 {
		t.Fatalf("got %d jobs, %d overdue; want 40 with 10 (25%%) rechecks", len(got), overdue)
	}
	// The oldest evidence is rechecked first.
	if got[30].label != "o00" {
		t.Errorf("first recheck = %s, want o00", got[30].label)
	}
	// With room for everything, nothing is cut.
	if got := prioritize(append([]j{}, jobs[:20]...), 40, key); len(got) != 20 {
		t.Errorf("len = %d", len(got))
	}
	// Few new names: the spare budget goes to rechecks.
	if got := prioritize(append([]j{}, jobs[95:]...), 40, key); len(got) != 40 {
		t.Errorf("spare budget unused: %d", len(got))
	}
}

func TestCheckRejectsIDNAndKeepsRegistrabilitySeparate(t *testing.T) {
	c := fakeChecker()
	for _, d := range []string{"xn--bcher-kva.de", "bücher.de", "ab--cd.io", "q7.xn--p1ai"} {
		r := c.check(context.Background(), d)
		if !strings.HasPrefix(r.Verdict, "invalid") || r.DNS != "-" || r.Registration != "invalid" || r.Registrable != "no" {
			t.Errorf("%s: %+v, want rejected without lookups", d, r)
		}
	}
	if r := c.check(context.Background(), "xn--bcher-kva.de"); !strings.Contains(r.Verdict, "IDN") {
		t.Errorf("verdict %q does not explain the IDN scope", r.Verdict)
	}
	cases := map[string][2]string{
		"free.dev":   {"not_registered", "unconfirmed"},
		"google.com": {"registered", "no"},
		"gone.io":    {"dropping", "no"},
		"flaky.com":  {"not_delegated", "unconfirmed"},
		"nordap.gg":  {"not_delegated", "unconfirmed"},
	}
	for d, want := range cases {
		r := c.check(context.Background(), d)
		if r.Registration != want[0] || r.Registrable != want[1] {
			t.Errorf("%s: registration %q registrable %q, want %v", d, r.Registration, r.Registrable, want)
		}
	}
}

// writeCov writes dir/bitmaps/<tld>.cov.json with the given shards swept.
func writeCov(t *testing.T, dir, tld string, at int64, shards ...int) {
	t.Helper()
	c := &store.Coverage{Shards: 7, Scans: map[string]*store.ShardScan{}}
	for _, s := range shards {
		c.Record(s, 7, &store.ShardScan{At: at, Names: shardSize(s, 7)})
	}
	if err := c.Save(store.CoveragePath(filepath.Join(dir, "bitmaps", tld+".bin"))); err != nil {
		t.Fatal(err)
	}
}

func manyFree(prefix string, n int) map[string]*store.Entry {
	m := map[string]*store.Entry{}
	for i := 0; i < n; i++ {
		m[fmt.Sprintf("%s%d", prefix, i)] = &store.Entry{DNS: "free"}
	}
	return m
}

// A run in which a resolver fault suddenly makes thousands of names look
// not delegated must not replace yesterday's data.
func TestMergeHoldsImplausibleJump(t *testing.T) {
	root := t.TempDir()
	day1 := time.Date(2026, 10, 26, 3, 0, 0, 0, time.UTC)
	base := filepath.Join(root, "prev")
	writePart(t, base, day1, map[string]map[string]*store.Entry{"dev": manyFree("a", 300), "io": manyFree("b", 300)}, nil)
	writePart(t, filepath.Join(root, "dev"), day1.AddDate(0, 0, 1), map[string]map[string]*store.Entry{"dev": manyFree("a", 1300)}, nil)
	writePart(t, filepath.Join(root, "io"), day1.AddDate(0, 0, 1), map[string]map[string]*store.Entry{"io": manyFree("b", 310)}, nil)
	st, _, _, err := mergeOwned(base, []string{filepath.Join(root, "dev"), filepath.Join(root, "io")}, jumpLimit{0.5, 200})
	if err != nil {
		t.Fatal(err)
	}
	if len(st.TLDs["dev"]) != 300 || st.Scans["dev"] == nil || !strings.Contains(st.Scans["dev"].Held, "300 to 1300") {
		t.Errorf("dev: %d entries, scan %+v; want base kept and held", len(st.TLDs["dev"]), st.Scans["dev"])
	}
	if !st.Scans["dev"].ScannedAt().Equal(day1) {
		t.Errorf("held dev scanned at %v, want the base's %v", st.Scans["dev"].ScannedAt(), day1)
	}
	if len(st.TLDs["io"]) != 310 || st.Scans["io"].Held != "" {
		t.Errorf("io: %d entries, scan %+v; want the new output", len(st.TLDs["io"]), st.Scans["io"])
	}
}

func TestMergeHoldsImplausibleShard(t *testing.T) {
	root := t.TempDir()
	at := time.Date(2026, 10, 26, 3, 0, 0, 0, time.UTC).Unix()
	base := filepath.Join(root, "prev")
	var some, lots []string
	names.Each(4, 0, 7, func(i int, s string) {
		if len(lots) < 2000 {
			lots = append(lots, s)
		}
		if len(some) < 10 {
			some = append(some, s)
		}
	})
	writePart(t, base, time.Unix(at, 0), nil, map[string][]string{"dev": some, "io": some})
	writeCov(t, base, "dev", at, 0, 1)
	writeCov(t, base, "io", at, 1) // shard 0 never swept before: first sweep fills it
	for _, tld := range []string{"dev", "io"} {
		dir := filepath.Join(root, tld)
		writePart(t, dir, time.Unix(at+86400, 0), nil, map[string][]string{tld: lots})
		writeCov(t, dir, tld, at+86400, 0, 1)
	}
	_, bms, covs, err := mergeOwned(base, []string{filepath.Join(root, "dev"), filepath.Join(root, "io")}, jumpLimit{0.5, 200})
	if err != nil {
		t.Fatal(err)
	}
	if bms["dev"].Count() != 10 || !strings.Contains(covs["dev"].Held, "shard 0") {
		t.Errorf("dev: %d bits, cov %+v; want base kept and held", bms["dev"].Count(), covs["dev"])
	}
	if bms["io"].Count() != 2000 || covs["io"].Held != "" || covs["io"].Scanned(0) == nil {
		t.Errorf("io: %d bits, cov %+v; want the new bitmap", bms["io"].Count(), covs["io"])
	}
}

func TestMergeOrRequireComplete(t *testing.T) {
	root := t.TempDir()
	now := time.Now().UTC()
	short := &store.State{Updated: now, TLDs: map[string]map[string]*store.Entry{"dev": {"q7": free(now.Unix())}}}
	short.SetScan("dev", now)
	if err := short.Save(filepath.Join(root, "short-dev", "state", "state.json.gz")); err != nil {
		t.Fatal(err)
	}
	for s := 0; s < 7; s++ {
		dir := filepath.Join(root, fmt.Sprintf("four-%d", s))
		writePart(t, dir, now, nil, map[string][]string{"dev": nil})
		writeCov(t, dir, "dev", now.Unix(), s)
	}
	dirs := []string{filepath.Join(root, "short-dev")}
	for s := 0; s < 7; s++ {
		dirs = append(dirs, filepath.Join(root, fmt.Sprintf("four-%d", s)))
	}
	st, bms, covs, err := mergeOr(dirs)
	if err != nil {
		t.Fatal(err)
	}
	if !covs["dev"].Complete() || st.Scans["dev"] == nil {
		t.Fatalf("coverage %+v, scans %+v", covs["dev"], st.Scans)
	}
	if err := checkComplete(st, bms, covs, []string{"dev"}); err != nil {
		t.Errorf("complete bootstrap rejected: %v", err)
	}
	// One shard job failed: refuse.
	st, bms, covs, _ = mergeOr(dirs[:7])
	if err := checkComplete(st, bms, covs, []string{"dev"}); err == nil || !strings.Contains(err.Error(), ".dev 4 characters") {
		t.Errorf("err = %v, want missing 4 characters", err)
	}
	// A TLD's short job failed: refuse.
	if err := checkComplete(st, bms, covs, []string{"dev", "io"}); err == nil || !strings.Contains(err.Error(), ".io 1-3 characters") {
		t.Errorf("err = %v, want missing .io", err)
	}
	// The command writes bitmaps together with their coverage.
	out := filepath.Join(root, "out")
	err = cmdMerge(append([]string{"-or", "-require-complete", "-tlds", "dev", "-state", filepath.Join(out, "state.json.gz"), "-bitmaps", filepath.Join(out, "bitmaps")}, dirs...))
	if err != nil {
		t.Fatal(err)
	}
	if c, err := store.LoadCoverage(filepath.Join(out, "bitmaps", "dev.cov.json"), 7); err != nil || !c.Complete() {
		t.Errorf("written coverage %+v, %v", c, err)
	}
}

func TestSnapshot(t *testing.T) {
	v1 := exportFixture(t) // <tmp>/api/v1
	root := filepath.Dir(filepath.Dir(v1))
	src := filepath.Join(root, "src")
	for name, body := range map[string]string{"verified-dev.txt": "q7.dev\nabc.dev\n", "changes.json": "[]\n"} {
		if err := store.WriteAtomic(filepath.Join(src, name), []byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	out := filepath.Join(root, "snap")
	args := []string{"-in", src, "-stats", filepath.Join(root, "data", "stats.json"), "-coverage", filepath.Join(root, "data", "coverage.json"), "-out", out}
	if err := cmdSnapshot(args); err != nil {
		t.Fatal(err)
	}
	var m Manifest
	b, err := os.ReadFile(filepath.Join(out, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	mustJSON(t, b, &m)
	if m.Name != "snapshot-2026-10-26" || m.Archive.Name != "snapshot-2026-10-26.tar.gz" || len(m.Files) != 2 || m.Files[0].Name != "changes.json" {
		t.Fatalf("manifest = %s", b)
	}
	if d := m.TLDs["dev"]; d == nil || d.Verified != 2 || d.Stale != 1 || d.Four != nil || d.Coverage != covPartial || d.OldestCheck == nil {
		t.Errorf("dev total = %+v", d)
	}
	if io := m.TLDs["io"]; io == nil || io.Four == nil || *io.Four != 1 || io.FourUnscanned == 0 {
		t.Errorf("io total = %+v", io)
	}
	arc, err := os.ReadFile(filepath.Join(out, m.Archive.Name))
	if err != nil {
		t.Fatal(err)
	}
	if got := sum(m.Archive.Name, arc); got != m.Archive {
		t.Errorf("archive checksum %+v, manifest says %+v", got, m.Archive)
	}
	// The archive holds exactly the listed files with the listed checksums.
	zr, err := gzip.NewReader(bytes.NewReader(arc))
	if err != nil {
		t.Fatal(err)
	}
	tr := tar.NewReader(zr)
	var got []FileSum
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(tr)
		got = append(got, sum(strings.TrimPrefix(h.Name, m.Name+"/"), body))
	}
	if !reflect.DeepEqual(got, m.Files) {
		t.Errorf("archive %+v, manifest %+v", got, m.Files)
	}
	// Same input, same archive: reruns do not produce spurious changes.
	if err := cmdSnapshot(append(args, "-name", "snapshot-2026-10-26-2")); err != nil {
		t.Fatal(err)
	}
	if err := cmdSnapshot(append(args, "-name", "../evil")); err == nil {
		t.Error("bad snapshot name accepted")
	}
}

func TestFindRejectsNonASCIIFilters(t *testing.T) {
	s, _ := newTestServer(t)
	for _, in := range []findInput{{Contains: "ü"}, {StartsWith: "ä1"}, {Contains: "a.b"}} {
		if _, _, err := s.findFreeDomains(context.Background(), nil, in); err == nil || !strings.Contains(err.Error(), "a-z") {
			t.Errorf("%+v: err = %v", in, err)
		}
	}
}
