package main

import (
	"context"
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
)

func verdictOf(dns probe.Result, r rdap.Status) string {
	switch {
	case r == rdap.Available:
		return "FREE"
	case r == rdap.Dropping:
		return "dropping soon"
	case r == rdap.Registered:
		return "taken"
	case dns == probe.Free:
		return "probably free (RDAP failed)"
	case dns == probe.Taken:
		return "taken"
	}
	return "unknown (DNS and RDAP failed)"
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
	recheck := fs.Duration("recheck", 10*24*time.Hour, "re-verify names whose last RDAP check is older than this")
	fs.Parse(args)

	st, err := store.Load(*statePath)
	if err != nil {
		return err
	}
	now := time.Now().UTC().Truncate(time.Second)
	var mu sync.Mutex
	var wg sync.WaitGroup
	errs := make(chan error, 16)

	for _, tld := range parseTLDs(*tlds) {
		wg.Add(1)
		go func(tld string) {
			defer wg.Done()
			p, err := probe.New(ctx, tld, *qps)
			if err != nil {
				errs <- err
				return
			}
			var labels []string
			for n := 1; n <= *maxLen; n++ {
				names.Each(n, 0, 1, func(_ int, s string) { labels = append(labels, s) })
			}
			fmt.Fprintf(os.Stderr, "scan .%s: %d names via %d servers\n", tld, len(labels), len(p.Servers()))
			var free, taken, unknown int
			sweep(ctx, p, labels, *workers, func(l string, r probe.Result) {
				mu.Lock()
				defer mu.Unlock()
				m := st.TLD(tld)
				switch r {
				case probe.Taken:
					taken++
					delete(m, l)
				case probe.Free:
					free++
					e, ok := m[l]
					if !ok {
						e = &store.Entry{Since: now.Unix()}
						m[l] = e
					}
					if e.DNS != "free" {
						e.RDAP = "" // state changed: verify again
					}
					e.DNS = "free"
				default:
					unknown++
					if e, ok := m[l]; ok {
						e.DNS = "unknown"
					}
				}
			})
			fmt.Fprintf(os.Stderr, "scan .%s done: %d not delegated, %d delegated, %d unknown\n", tld, free, taken, unknown)
			if ctx.Err() != nil {
				errs <- ctx.Err()
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

	if *rdapTime > 0 {
		verify(ctx, st, parseTLDs(*tlds), *rdapRPS, *rdapTime, *recheck, now)
	}
	st.Updated = now
	return st.Save(*statePath)
}

// verify confirms not-delegated names via RDAP. Never-checked names come
// first, then dropping names, then the stalest checks; each registry host gets
// its own time budget, enforced by a deadline, and runs in parallel with the
// others.
func verify(ctx context.Context, st *store.State, tlds []string, rps float64, budget time.Duration, recheck time.Duration, now time.Time) {
	type job struct {
		tld, label string
		e          *store.Entry
	}
	byHost := map[string][]job{}
	for _, tld := range tlds {
		host := rdap.Endpoints[tld]
		if host == "" {
			continue
		}
		for label, e := range st.TLDs[tld] {
			if e.RDAP == "" || e.RDAP == string(rdap.Dropping) || e.RDAP == string(rdap.Error) || now.Sub(e.CheckedAt()) > recheck {
				byHost[host] = append(byHost[host], job{tld, label, e})
			}
		}
	}
	rank := func(e *store.Entry) int {
		switch e.RDAP {
		case "":
			return 0
		case string(rdap.Dropping):
			return 1
		case string(rdap.Error):
			return 2
		}
		return 3
	}
	rc := rdap.New(rps, userAgent)
	var wg sync.WaitGroup
	var mu sync.Mutex
	for host, jobs := range byHost {
		sort.Slice(jobs, func(i, j int) bool {
			a, b := jobs[i], jobs[j]
			if rank(a.e) != rank(b.e) {
				return rank(a.e) < rank(b.e)
			}
			if len(a.label) != len(b.label) {
				return len(a.label) < len(b.label)
			}
			if a.e.Checked != b.e.Checked {
				return a.e.Checked < b.e.Checked
			}
			return a.tld+a.label < b.tld+b.label
		})
		if max := int(budget.Seconds() * rc.Rate(host)); len(jobs) > max {
			jobs = jobs[:max]
		}
		fmt.Fprintf(os.Stderr, "rdap %s: %d lookups\n", host, len(jobs))
		wg.Add(1)
		go func(host string, jobs []job) {
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
							j.e.RDAP = string(info.Status)
							j.e.Checked = now.Unix()
							j.e.Expires = 0
							if !info.Expires.IsZero() {
								j.e.Expires = info.Expires.Unix()
							}
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
		}(host, jobs)
	}
	wg.Wait()
}

func cmdScan4(ctx context.Context, args []string) error {
	fs := newFlags("scan4")
	tlds := fs.String("tlds", "", "comma-separated TLDs (default: all supported)")
	dir := fs.String("dir", "bitmaps", "directory with <tld>.bin free-bitmaps")
	shards := fs.Int("shards", 7, "split the 4 character space into this many daily shards")
	shard := fs.Int("shard", -1, "shard to scan (-1 = day of year mod shards)")
	qps := fs.Int("qps", 400, "DNS queries per second per TLD")
	workers := fs.Int("workers", 96, "concurrent DNS queries per TLD")
	fs.Parse(args)

	if *shard < 0 {
		*shard = time.Now().UTC().YearDay() % *shards
	}
	space := names.Space(4)
	var wg sync.WaitGroup
	errs := make(chan error, 16)
	for _, tld := range parseTLDs(*tlds) {
		wg.Add(1)
		go func(tld string) {
			defer wg.Done()
			path := filepath.Join(*dir, tld+".bin")
			bm, err := store.LoadBitmap(path, space)
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
