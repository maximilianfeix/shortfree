package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/maximilianfeix/shortfree/internal/names"
	"github.com/maximilianfeix/shortfree/internal/probe"
	"github.com/maximilianfeix/shortfree/internal/rdap"
	"github.com/maximilianfeix/shortfree/internal/store"
)

// Tests sweep 4 character shard 3 of 37 (labels starting with "d", ~48k
// names) against a fake check function: no network.
const testShard, testShards = 3, 37

func shardLabels() []string {
	var labels []string
	names.Each(4, testShard, testShards, func(_ int, s string) { labels = append(labels, s) })
	return labels
}

// recorder is a fake DNS check that answers every name as taken after delay
// and remembers which names it was asked about.
type recorder struct {
	delay time.Duration
	mu    sync.Mutex
	asked map[string]bool
}

func (r *recorder) check(ctx context.Context, label string) probe.Result {
	time.Sleep(r.delay)
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.asked == nil {
		r.asked = map[string]bool{}
	}
	r.asked[label] = true
	return probe.Taken
}

// A sweep cut off by its time budget must keep the shard's earlier coverage
// record and every bit it did not query: no unqueried name may turn taken.
func TestScan4DeadlineKeepsUnsweptNames(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.bin")
	labels := shardLabels()
	bm := store.NewBitmap(names.Space(4))
	for _, l := range labels {
		bm.Set(names.Index(l), true) // every name free at the last sweep
	}
	if err := store.WriteAtomic(path, bm); err != nil {
		t.Fatal(err)
	}
	old := &store.Coverage{Shards: testShards, Scans: map[string]*store.ShardScan{}}
	old.Record(testShard, testShards, &store.ShardScan{At: 1000, Names: len(labels)})
	if err := old.Save(store.CoveragePath(path)); err != nil {
		t.Fatal(err)
	}

	// Enumerating the labels takes a while under -race; at 4 workers x 10 ms
	// the 3 s budget then covers at most ~1200 of ~49k names.
	rec := &recorder{delay: 10 * time.Millisecond}
	if err := scan4TLD(context.Background(), "test", path, rec.check, testShard, testShards, 4, time.Now().Add(3*time.Second), time.Time{}); err != nil {
		t.Fatal(err)
	}
	if len(rec.asked) == 0 || len(rec.asked) >= len(labels) {
		t.Fatalf("queried %d of %d names, want a cut-off sweep", len(rec.asked), len(labels))
	}
	got, err := store.ReadBitmap(path, names.Space(4))
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range labels {
		if free := got.Get(names.Index(l)); free == rec.asked[l] {
			t.Fatalf("%s: free=%v, queried=%v; want answered names taken, the rest unchanged", l, free, rec.asked[l])
		}
	}
	cov, err := store.LoadCoverage(store.CoveragePath(path), testShards)
	if err != nil {
		t.Fatal(err)
	}
	if s := cov.Scanned(testShard); s == nil || s.At != 1000 {
		t.Errorf("shard record = %+v, want the previous one (at 1000)", s)
	}
}

// The first sweep of a new TLD that gets cut off must leave a coverage file
// without the shard: a bitmap without coverage file would count as fully
// swept (legacy) and its zero bits as taken.
func TestScan4DeadlineFirstSweepNotLegacy(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.bin")
	legacyAt := time.Unix(5000, 0)
	rec := &recorder{delay: time.Millisecond}
	if err := scan4TLD(context.Background(), "test", path, rec.check, testShard, testShards, 4, time.Now().Add(50*time.Millisecond), legacyAt); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("bitmap not written: %v", err)
	}
	cov, err := store.BitmapCoverage(path, testShards, legacyAt)
	if err != nil {
		t.Fatal(err)
	}
	if cov.LegacyAt != 0 || len(cov.Scans) != 0 {
		t.Errorf("coverage = %+v, want nothing swept", cov)
	}
}

// Without a budget the whole shard is swept and recorded, unknowns keep
// their previous bit.
func TestScan4CompleteSweepRecordsShard(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.bin")
	labels := shardLabels()
	bm := store.NewBitmap(names.Space(4))
	bm.Set(names.Index(labels[2]), true) // free before, no answer now
	if err := store.WriteAtomic(path, bm); err != nil {
		t.Fatal(err)
	}
	answer := func(_ context.Context, l string) probe.Result {
		switch l {
		case labels[0]:
			return probe.Free
		case labels[2]:
			return probe.Unknown
		}
		return probe.Taken
	}
	before := time.Now().Unix()
	if err := scan4TLD(context.Background(), "test", path, answer, testShard, testShards, 8, time.Time{}, time.Time{}); err != nil {
		t.Fatal(err)
	}
	got, _ := store.ReadBitmap(path, names.Space(4))
	if got.Count() != 2 || !got.Get(names.Index(labels[0])) || !got.Get(names.Index(labels[2])) {
		t.Errorf("bitmap has %d free names, want %s (free) and %s (unknown, kept)", got.Count(), labels[0], labels[2])
	}
	cov, err := store.LoadCoverage(store.CoveragePath(path), testShards)
	if err != nil {
		t.Fatal(err)
	}
	s := cov.Scanned(testShard)
	if s == nil || s.Names != len(labels) || s.Unknown != 1 || s.At < before {
		t.Errorf("shard record = %+v, want %d names, 1 unknown, at >= %d", s, len(labels), before)
	}
}

// A cancelled sweep (SIGINT from the workflow's hard stop) saves the answers
// it got, records nothing and reports the cancellation.
func TestScan4CancelSavesAnswersOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.bin")
	ctx, cancel := context.WithCancel(context.Background())
	var n sync.Mutex
	count := 0
	check := func(_ context.Context, l string) probe.Result {
		n.Lock()
		defer n.Unlock()
		if count++; count == 100 {
			cancel()
		}
		return probe.Free
	}
	if err := scan4TLD(ctx, "test", path, check, testShard, testShards, 1, time.Time{}, time.Time{}); err == nil {
		t.Fatal("want the cancellation as error")
	}
	got, _ := store.ReadBitmap(path, names.Space(4))
	if c := got.Count(); c < 100 || c >= len(shardLabels()) {
		t.Errorf("bitmap has %d free names, want the ~100 answered before cancel", c)
	}
	cov, err := store.LoadCoverage(store.CoveragePath(path), testShards)
	if err != nil {
		t.Fatal(err)
	}
	if len(cov.Scans) != 0 {
		t.Errorf("coverage = %+v, want nothing recorded", cov.Scans)
	}
}

// The daily publish path for a job whose 4 character sweep was cut off: the
// part's coverage keeps the shard unscanned through merge and export, and
// the answers it did get are published.
func TestMergePartialScan4Part(t *testing.T) {
	root := t.TempDir()
	day1 := time.Now().Add(-24 * time.Hour)
	base, part := filepath.Join(root, "prev"), filepath.Join(root, "tld-dev")
	writePart(t, base, day1, map[string]map[string]*store.Entry{"dev": {"q7": free(1)}}, map[string][]string{"dev": {"aaaa"}})
	baseCov := &store.Coverage{Shards: 7, Scans: map[string]*store.ShardScan{}}
	baseCov.Record(0, 7, &store.ShardScan{At: day1.Unix(), Names: shardSize(0, 7)})
	if err := baseCov.Save(filepath.Join(base, "bitmaps", "dev.cov.json")); err != nil {
		t.Fatal(err)
	}
	// The job: restore base, sweep 1-3 characters, then get cut off in shard 1.
	writePart(t, part, time.Now(), map[string]map[string]*store.Entry{"dev": {"q7": free(2)}}, nil)
	for _, f := range []string{"dev.bin", "dev.cov.json"} {
		b, err := os.ReadFile(filepath.Join(base, "bitmaps", f))
		if err != nil {
			t.Fatal(err)
		}
		if err := store.WriteAtomic(filepath.Join(part, "bitmaps", f), b); err != nil {
			t.Fatal(err)
		}
	}
	answered := map[string]bool{}
	var mu sync.Mutex
	check := func(_ context.Context, l string) probe.Result {
		time.Sleep(10 * time.Millisecond)
		mu.Lock()
		defer mu.Unlock()
		answered[l] = true
		return probe.Free
	}
	if err := scan4TLD(context.Background(), "dev", filepath.Join(part, "bitmaps", "dev.bin"), check, 1, 7, 4, time.Now().Add(3*time.Second), day1); err != nil {
		t.Fatal(err)
	}

	_, bms, covs, err := mergeOwned(base, []string{part}, jumpLimit{0.5, 200})
	if err != nil {
		t.Fatal(err)
	}
	c := covs["dev"]
	if c == nil || c.Scanned(0) == nil || c.Scanned(1) != nil || c.Complete() || c.LegacyAt != 0 {
		t.Fatalf("merged coverage = %+v, want shard 0 swept, shard 1 not", c)
	}
	if unscanned, _ := shardCounts(c); unscanned != validFour()-shardSize(0, 7) {
		t.Errorf("unscanned = %d, want every name outside shard 0", unscanned)
	}
	if got := bms["dev"].Count(); got != len(answered)+1 {
		t.Errorf("merged bitmap has %d free names, want aaaa + %d answered", got, len(answered))
	}
}

func validFour() int {
	n := 0
	names.Each(4, 0, 1, func(int, string) { n++ })
	return n
}

func TestSweepDeadlinePassed(t *testing.T) {
	rec := &recorder{}
	sent := sweep(context.Background(), "test", rec.check, []string{"a", "b"}, 2, time.Now().Add(-time.Second), func(string, probe.Result) {})
	if sent != 0 || len(rec.asked) != 0 {
		t.Errorf("sent %d, asked %v; want nothing after the deadline", sent, rec.asked)
	}
}

// fakeRDAP serves TLD "test": canaryOK decides whether nic.test exists,
// every other name is not found.
func fakeRDAP(t *testing.T, canaryOK bool) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if canaryOK && strings.HasSuffix(r.URL.Path, "/nic.test") {
			w.Write([]byte(`{"status":["active"]}`))
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(srv.Close)
	rdap.Endpoints["test"] = srv.URL + "/"
	t.Cleanup(func() { delete(rdap.Endpoints, "test") })
}

func TestVerifyTrustsHealthyRegistry(t *testing.T) {
	fakeRDAP(t, true)
	now := time.Now().UTC()
	st := &store.State{TLDs: map[string]map[string]*store.Entry{"test": {"q7": {DNS: "free"}}}}
	verify(context.Background(), st, []string{"test"}, 1000, map[string]time.Duration{"test": time.Minute}, 24*time.Hour, now)
	if e := st.TLDs["test"]["q7"]; e.RDAP != string(rdap.Available) || e.Checked != now.Unix() {
		t.Errorf("q7 = %+v, want available", e)
	}
}

// A registry that answers 404 even for its own nic.<tld> must not turn
// queued names into "available".
func TestVerifySkipsRegistryFailingCanary(t *testing.T) {
	fakeRDAP(t, false)
	st := &store.State{TLDs: map[string]map[string]*store.Entry{"test": {"q7": {DNS: "free"}}}}
	verify(context.Background(), st, []string{"test"}, 1000, map[string]time.Duration{"test": time.Minute}, 24*time.Hour, time.Now())
	if e := st.TLDs["test"]["q7"]; e.RDAP != "" || e.Checked != 0 {
		t.Errorf("q7 = %+v, want still unverified", e)
	}
}

func TestCheckBelowMinLen(t *testing.T) {
	c := &checker{
		dns: func(_ context.Context, label, tld string) (probe.Result, error) {
			if label == "bit" {
				return probe.Taken, nil
			}
			return probe.Free, nil
		},
		rdap: func(_ context.Context, label, tld string) (rdap.Info, bool, error) {
			if label == "bit" {
				return rdap.Info{Status: rdap.Registered}, true, nil
			}
			return rdap.Info{Status: rdap.Available}, true, nil
		},
	}
	for domain, want := range map[string]string{
		"ab.so":   "not registrable (below the registry's minimum length of 3)",
		"abc.ly":  "not registrable (below the registry's minimum length of 4)",
		"bit.ly":  "taken", // grandfathered registrations stay taken
		"abc.so":  vNotRegistered,
		"q7.dev":  vNotRegistered,
		"abcd.ly": vNotRegistered,
	} {
		r := c.check(context.Background(), domain)
		if r.Verdict != want || r.Free != (want == vNotRegistered) {
			t.Errorf("%s: verdict %q free %v, want %q", domain, r.Verdict, r.Free, want)
		}
	}
}

// Long labels used to be rejected at random: names.Index overflowed.
func TestSplitDomainLongLabels(t *testing.T) {
	for _, d := range []string{"sf-ci-dac349e53fb6270b.tools", "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz.com", "a.io"} {
		if _, _, ok := splitDomain(d); !ok {
			t.Errorf("splitDomain(%q) rejected", d)
		}
	}
	for _, d := range []string{"a_b.io", "-a.io", strings.Repeat("a", 64) + ".io"} {
		if _, _, ok := splitDomain(d); ok {
			t.Errorf("splitDomain(%q) accepted", d)
		}
	}
}

func TestVerifyZeroBudgetSkipsTLD(t *testing.T) {
	fakeRDAP(t, true)
	st := &store.State{TLDs: map[string]map[string]*store.Entry{"test": {"q7": {DNS: "free"}}}}
	verify(context.Background(), st, []string{"test"}, 1000, map[string]time.Duration{"test": 0}, 24*time.Hour, time.Now())
	if e := st.TLDs["test"]["q7"]; e.RDAP != "" {
		t.Errorf("q7 = %+v, want untouched", e)
	}
}
