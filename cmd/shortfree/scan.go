package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/maximilianfeix/shortfree/internal/names"
	"github.com/maximilianfeix/shortfree/internal/probe"
	"github.com/maximilianfeix/shortfree/internal/rdap"
	"github.com/maximilianfeix/shortfree/internal/store"
	"github.com/maximilianfeix/shortfree/internal/tlds"
)

// Verdicts of a live check. An RDAP 404 proves only that the registry has no
// registration for the name, not that a registrar will sell it: reserved,
// premium and restricted names also answer 404.
const (
	vNotRegistered = "not registered (RDAP 404)"
	vDropping      = "dropping (redemption or pending delete)"
	vTaken         = "taken"
)

func verdictOf(dns probe.Result, r rdap.Status) string {
	switch {
	case r == rdap.Available:
		return vNotRegistered
	case r == rdap.Dropping:
		return vDropping
	case r == rdap.Registered:
		return vTaken
	case dns == probe.Free:
		return "not in DNS (RDAP failed)"
	case dns == probe.Taken:
		return vTaken
	}
	return "unknown (DNS and RDAP failed)"
}

// registrationOf is the registration evidence behind a verdict.
func registrationOf(dns probe.Result, r rdap.Status) string {
	switch {
	case r == rdap.Available:
		return "not_registered"
	case r == rdap.Dropping:
		return "dropping"
	case r == rdap.Registered, dns == probe.Taken:
		return "registered"
	case dns == probe.Free:
		return "not_delegated"
	}
	return "unknown"
}

// sweep runs check for every label on prober p with the given number of
// workers, calling done for each result. It prints progress to stderr.
func sweep(ctx context.Context, p *probe.Prober, labels []string, workers int, done func(label string, r probe.Result)) {
	jobs := make(chan string, workers)
	var wg sync.WaitGroup
	var n atomic.Int64
	start := time.Now()
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for l := range jobs {
				done(l, p.Check(ctx, l))
				if c := n.Add(1); c%50000 == 0 {
					fmt.Fprintf(os.Stderr, "  .%s %d/%d (%.0f q/s)\n", p.TLD, c, len(labels), float64(c)/time.Since(start).Seconds())
				}
			}
		}()
	}
	for _, l := range labels {
		if ctx.Err() != nil {
			break
		}
		jobs <- l
	}
	close(jobs)
	wg.Wait()
}

func cmdScan(ctx context.Context, args []string) error {
	fs := newFlags("scan")
	tlds := fs.String("tlds", "", "comma-separated TLDs (default: all supported)")
	maxLen := fs.Int("max-len", 3, "longest label length to sweep")
	statePath := fs.String("state", "state/state.json.gz", "state file (.gz = compressed)")
	qps := fs.Int("qps", 250, "DNS queries per second per TLD")
	workers := fs.Int("workers", 64, "concurrent DNS queries per TLD")
	rdapRPS := fs.Float64("rdap-rps", 0, "override RDAP requests per second for every registry host (0 = per-registry defaults)")
	rdapTime := fs.Duration("rdap-time", 25*time.Minute, "RDAP time budget per registry host and run (0 = skip RDAP)")
	rdapTimeNew := fs.Duration("rdap-time-new", 0, "RDAP time budget for TLDs without previous state, i.e. their first run (0 = same as -rdap-time)")
	recheck := fs.Duration("recheck", 10*24*time.Hour, "re-verify names whose last RDAP check is older than this")
	keepOnly := fs.Bool("keep-only", false, "save only the scanned TLDs (output of one job in a per-TLD matrix)")
	fs.Parse(args)

	st, err := store.Load(*statePath)
	if err != nil {
		return err
	}
	now := time.Now().UTC().Truncate(time.Second)
	scanned := parseTLDs(*tlds)
	budgets := rdapBudgets(st, scanned, *rdapTime, *rdapTimeNew)
	errs := make(chan error, len(scanned))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, tld := range scanned {
		wg.Add(1)
		go func(tld string) {
			defer wg.Done()
			p, err := probe.New(ctx, tld, *qps)
			if err != nil {
				errs <- err
				return
			}
			labels := shortLabels(tld, *maxLen)
			mu.Lock()
			// Create the TLD even if nothing turns out free: an empty entry
			// in the saved state tells merge that every name was taken.
			st.TLD(tld)
			pruneBelowMinLen(st, tld)
			mu.Unlock()
			fmt.Fprintf(os.Stderr, "scan .%s: %d names via %d servers\n", tld, len(labels), len(p.Servers()))
			var free, taken, unknown int
			sweep(ctx, p, labels, *workers, func(l string, r probe.Result) {
				mu.Lock()
				defer mu.Unlock()
				sweepResult(st.TLD(tld), l, r, now)
				switch r {
				case probe.Taken:
					taken++
				case probe.Free:
					free++
				default:
					unknown++
				}
			})
			fmt.Fprintf(os.Stderr, "scan .%s done: %d not delegated, %d delegated, %d unknown\n", tld, free, taken, unknown)
			if ctx.Err() != nil {
				errs <- ctx.Err()
				return
			}
			mu.Lock()
			st.SetScan(tld, now)
			mu.Unlock()
		}(tld)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			return err
		}
	}

	verify(ctx, st, scanned, *rdapRPS, budgets, *recheck, now)
	if *keepOnly {
		keepTLDs(st, scanned)
	}
	st.Updated = now
	return st.Save(*statePath)
}

// sweepResult applies one DNS answer for label to a TLD's state entries.
// Delegated names are dropped (registered). Names without a DNS answer are
// kept with DNS "unknown", so the published data can say "not checked"
// instead of silently treating them as taken. Earlier RDAP evidence is kept:
// it expires by age (see statusOf), not by a DNS hiccup.
func sweepResult(m map[string]*store.Entry, label string, r probe.Result, now time.Time) {
	e, ok := m[label]
	switch r {
	case probe.Taken:
		delete(m, label)
	case probe.Free:
		if !ok {
			e = &store.Entry{}
			m[label] = e
		}
		if e.Since == 0 {
			e.Since = now.Unix()
		}
		e.DNS = "free"
	default:
		if !ok {
			e = &store.Entry{}
			m[label] = e
		}
		e.DNS = "unknown"
	}
}

// shortLabels returns every label of length 1..maxLen that tld's registry
// accepts.
func shortLabels(tld string, maxLen int) []string {
	var labels []string
	for n := tlds.Get(tld).MinLen; n <= maxLen; n++ {
		names.Each(n, 0, 1, func(_ int, s string) { labels = append(labels, s) })
	}
	return labels
}

// minLen is the shortest label tld's registry accepts.
func minLen(tld string) int { return tlds.Get(tld).MinLen }

// pruneBelowMinLen drops state entries the registry would never register,
// e.g. from before a minimum length was known.
func pruneBelowMinLen(st *store.State, tld string) {
	p := tlds.Get(tld)
	for label := range st.TLDs[tld] {
		if !p.Allows(label) {
			delete(st.TLDs[tld], label)
		}
	}
}

// keepTLDs removes every TLD except keep from st, so a matrix job's output
// carries only the TLDs it owns.
func keepTLDs(st *store.State, keep []string) {
	for tld := range st.TLDs {
		if !contains(keep, tld) {
			delete(st.TLDs, tld)
		}
	}
	for tld := range st.Scans {
		if !contains(keep, tld) {
			delete(st.Scans, tld)
		}
	}
}

// rdapBudgets returns the RDAP time budget per TLD: fresh for TLDs without
// any state (their first run, when every name still needs a check), normal
// for the rest. fresh <= 0 means normal.
func rdapBudgets(st *store.State, scanned []string, normal, fresh time.Duration) map[string]time.Duration {
	out := map[string]time.Duration{}
	for _, tld := range scanned {
		out[tld] = normal
		if fresh > 0 && normal > 0 && len(st.TLDs[tld]) == 0 {
			out[tld] = fresh
		}
	}
	return out
}

// canary looks up a name known to be registered on tld. Unless the registry
// confirms it, its 404s cannot be trusted (an endpoint that moved or answers
// 404 for everything would mark every queued name as available).
//
// It also looks up a long random label that cannot be registered by anyone
// (negative control): an endpoint that answers "registered" for everything
// would otherwise remove every queued name from the lists.
func canary(ctx context.Context, rc *rdap.Client, tld string) error {
	label := tlds.Get(tld).Canary
	cctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	info, err := rc.Lookup(cctx, label, tld)
	if err != nil {
		return fmt.Errorf("canary %s.%s: %w", label, tld, err)
	}
	if info.Status != rdap.Registered && info.Status != rdap.Dropping {
		return fmt.Errorf("canary %s.%s is %s, want registered: RDAP answers untrusted", label, tld, info.Status)
	}
	neg := negativeCanary()
	info, err = rc.Lookup(cctx, neg, tld)
	if err != nil {
		return fmt.Errorf("negative canary %s.%s: %w", neg, tld, err)
	}
	if info.Status != rdap.Available {
		return fmt.Errorf("negative canary %s.%s is %s, want not found: RDAP answers untrusted", neg, tld, info.Status)
	}
	return nil
}

// negativeCanary returns a random 28 character label nobody has registered.
func negativeCanary() string {
	var b [8]byte
	rand.Read(b[:])
	return fmt.Sprintf("sf-canary-%x", b)
}

// verify confirms not-delegated names via RDAP. Never-checked names come
// first, then dropping names, then the stalest checks; each registry host gets
// its own time budget (the largest of its TLDs' budgets), enforced by a
// deadline, and runs in parallel with the others. TLDs with a zero budget, and
// TLDs whose registry fails the canary lookup, are skipped.
func verify(ctx context.Context, st *store.State, scanned []string, rps float64, budgets map[string]time.Duration, recheck time.Duration, now time.Time) {
	type job struct {
		tld, label string
		e          *store.Entry
	}
	rc := rdap.New(rps, userAgent)
	// Other jobs scan the other TLDs at the same time: use only our share
	// of each registry's total rate.
	rc.ShareWith(scanned, tlds.Default)
	byHost := map[string][]job{}
	hostBudget := map[string]time.Duration{}
	for _, tld := range scanned {
		host := rdap.Host(tld)
		if host == "" || budgets[tld] <= 0 {
			continue
		}
		var jobs []job
		for label, e := range st.TLDs[tld] {
			if e.RDAP == "" || e.RDAP == string(rdap.Dropping) || e.RDAP == string(rdap.Error) || now.Sub(e.CheckedAt()) > recheck {
				jobs = append(jobs, job{tld, label, e})
			}
		}
		if len(jobs) == 0 {
			continue
		}
		if err := canary(ctx, rc, tld); err != nil {
			fmt.Fprintf(os.Stderr, "rdap .%s: skipped, %v\n", tld, err)
			continue
		}
		hostBudget[host] = max(hostBudget[host], budgets[tld])
		byHost[host] = append(byHost[host], jobs...)
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	for host, jobs := range byHost {
		budget := hostBudget[host]
		jobs = prioritize(jobs, int(budget.Seconds()*rc.Rate(host)), func(j job) (*store.Entry, string, string) { return j.e, j.label, j.tld + "." + j.label })
		fmt.Fprintf(os.Stderr, "rdap %s: %d lookups within %s at %.2f/s\n", host, len(jobs), budget, rc.Rate(host))
		wg.Add(1)
		go func(host string, jobs []job, budget time.Duration) {
			defer wg.Done()
			// Slow or failing registries make lookups retry, so the job count
			// alone does not bound the run time.
			hctx, cancel := context.WithTimeout(ctx, budget)
			defer cancel()
			counts := map[rdap.Status]int{}
			// Two requests in flight per host keep the limiter busy while
			// one request waits on the network.
			ch := make(chan job)
			var inner sync.WaitGroup
			for w := 0; w < 2; w++ {
				inner.Add(1)
				go func() {
					defer inner.Done()
					for j := range ch {
						info, err := rc.Lookup(hctx, j.label, j.tld)
						if hctx.Err() != nil {
							continue // cut off by the deadline: keep the old result
						}
						mu.Lock()
						counts[info.Status]++
						if err == nil || j.e.RDAP == "" {
							applyRDAP(j.e, info, now)
						}
						mu.Unlock()
					}
				}()
			}
		dispatch:
			for _, j := range jobs {
				select {
				case ch <- j:
				case <-hctx.Done():
					break dispatch
				}
			}
			close(ch)
			inner.Wait()
			fmt.Fprintf(os.Stderr, "rdap %s done: %v\n", host, counts)
		}(host, jobs, budget)
	}
	wg.Wait()
}

// applyRDAP stores a lookup result. A registered or dropping name that now
// answers 404 records the time of that observed transition (Freed): only
// such names are reported as "freed" in the change log.
func applyRDAP(e *store.Entry, info rdap.Info, now time.Time) {
	if info.Status == rdap.Available && (e.RDAP == string(rdap.Registered) || e.RDAP == string(rdap.Dropping)) {
		e.Freed = now.Unix()
	}
	e.RDAP = string(info.Status)
	e.Checked = now.Unix()
	e.Expires = 0
	if !info.Expires.IsZero() {
		e.Expires = info.Expires.Unix()
	}
	e.Phase = info.Phase
}

// verifyRank orders RDAP work: new not-delegated names first, then dropping
// names (their state changes fast), then failed lookups, then names without
// a DNS answer, then rechecks of old evidence.
func verifyRank(e *store.Entry) int {
	switch {
	case e.RDAP == "" && e.DNS == "free":
		return 0
	case e.RDAP == string(rdap.Dropping):
		return 1
	case e.RDAP == string(rdap.Error):
		return 2
	case e.RDAP == "":
		return 3
	}
	return 4
}

// overdueShare is the part of a host's RDAP budget reserved for rechecks of
// old evidence. Without it, a steady stream of new names would starve
// rechecks and published evidence would only ever get older.
const overdueShare = 0.25

// prioritize sorts jobs by verifyRank (shortest labels, then oldest checks
// first within a rank) and cuts them to max, keeping up to overdueShare of
// max for rechecks (rank 4) when there are more jobs than max.
func prioritize[J any](jobs []J, max int, key func(J) (e *store.Entry, label, id string)) []J {
	sort.Slice(jobs, func(i, j int) bool {
		a, la, ia := key(jobs[i])
		b, lb, ib := key(jobs[j])
		if verifyRank(a) != verifyRank(b) {
			return verifyRank(a) < verifyRank(b)
		}
		if len(la) != len(lb) {
			return len(la) < len(lb)
		}
		if a.Checked != b.Checked {
			return a.Checked < b.Checked
		}
		return ia < ib
	})
	if max < 0 {
		max = 0
	}
	if len(jobs) <= max {
		return jobs
	}
	var first, overdue []J
	for _, j := range jobs {
		if e, _, _ := key(j); verifyRank(e) == 4 {
			overdue = append(overdue, j)
		} else {
			first = append(first, j)
		}
	}
	keep := min(len(overdue), int(float64(max)*overdueShare))
	if n := max - len(first); n > keep {
		keep = min(len(overdue), n) // spare budget beyond the reserve
	}
	out := append(first[:min(len(first), max-keep)], overdue[:keep]...)
	return out
}

func cmdScan4(ctx context.Context, args []string) error {
	fs := newFlags("scan4")
	tlds := fs.String("tlds", "", "comma-separated TLDs (default: all supported)")
	dir := fs.String("dir", "bitmaps", "directory with <tld>.bin free-bitmaps and their <tld>.cov.json shard coverage")
	shards := fs.Int("shards", 7, "split the 4 character space into this many daily shards")
	shard := fs.Int("shard", -1, "shard to scan (-1 = day of year mod shards)")
	qps := fs.Int("qps", 400, "DNS queries per second per TLD")
	workers := fs.Int("workers", 96, "concurrent DNS queries per TLD")
	fs.Parse(args)

	if *shards < 1 {
		return fmt.Errorf("scan4: -shards must be at least 1")
	}
	if *shard < 0 {
		*shard = time.Now().UTC().YearDay() % *shards
	}
	if *shard >= *shards {
		return fmt.Errorf("scan4: -shard %d out of range 0-%d", *shard, *shards-1)
	}
	space := names.Space(4)
	var wg sync.WaitGroup
	errs := make(chan error, 16)
	for _, tld := range parseTLDs(*tlds) {
		wg.Add(1)
		go func(tld string) {
			defer wg.Done()
			if minLen(tld) > 4 {
				return
			}
			path := filepath.Join(*dir, tld+".bin")
			bm, err := store.LoadBitmap(path, space)
			if err != nil {
				errs <- err
				return
			}
			cov, err := store.LoadCoverage(store.CoveragePath(path), *shards)
			if err != nil {
				errs <- err
				return
			}
			p, err := probe.New(ctx, tld, *qps)
			if err != nil {
				errs <- err
				return
			}
			var labels []string
			names.Each(4, *shard, *shards, func(_ int, s string) { labels = append(labels, s) })
			fmt.Fprintf(os.Stderr, "scan4 .%s shard %d/%d: %d names\n", tld, *shard, *shards, len(labels))
			var mu sync.Mutex
			var unknown int
			sweep(ctx, p, labels, *workers, func(l string, r probe.Result) {
				mu.Lock()
				defer mu.Unlock()
				switch r {
				case probe.Free:
					bm.Set(names.Index(l), true)
				case probe.Taken:
					bm.Set(names.Index(l), false)
				default:
					unknown++ // keep the previous bit
				}
			})
			if ctx.Err() != nil {
				errs <- ctx.Err()
				return
			}
			fmt.Fprintf(os.Stderr, "scan4 .%s done: %d free in bitmap, %d unknown this run\n", tld, bm.Count(), unknown)
			if err := store.WriteAtomic(path, bm); err != nil {
				errs <- err
				return
			}
			// Record the shard as swept only after the bitmap is saved.
			cov.Record(*shard, *shards, &store.ShardScan{At: time.Now().UTC().Unix(), Names: len(labels), Unknown: unknown})
			if err := cov.Save(store.CoveragePath(path)); err != nil {
				errs <- err
			}
		}(tld)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}
