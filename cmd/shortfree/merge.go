package main

import (
	"fmt"
	"math/bits"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/maximilianfeix/shortfree/internal/names"
	"github.com/maximilianfeix/shortfree/internal/store"
)

// cmdMerge combines the outputs of parallel scan jobs. Each input directory
// may hold state/state.json.gz and bitmaps/<tld>.bin.
//
// By default (daily per-TLD scan) every input owns the TLDs in its state file
// and the bitmaps it carries, and those replace the base data wholesale: a
// name that became taken disappears because the owning job no longer lists
// it. -base supplies the previous data, which survives for TLDs whose job
// produced no output (failed or skipped). Two inputs claiming the same TLD is
// an error.
//
// With -or (bootstrap only) state entries are merged per label, the more
// recent RDAP check winning, and bitmaps are OR-ed. That is only correct
// because every bootstrap job sweeps a disjoint shard from an empty bitmap.
// -require-complete then refuses to write anything unless every TLD has a
// complete 1-3 character sweep and every 4 character shard was swept.
//
// Safeguard (daily mode): an input whose count of not-delegated names (or of
// free 4 character names in a shard) differs from the base by more than
// -max-jump (and at least -min-jump names) is rejected for that TLD; the
// base data is kept and marked "held". A broken resolver or registry can
// otherwise flip thousands of names in one run.
func cmdMerge(args []string) error {
	fs := newFlags("merge")
	statePath := fs.String("state", "state/state.json.gz", "merged state file to write")
	bitmaps := fs.String("bitmaps", "bitmaps", "directory for merged bitmaps and their .cov.json coverage")
	base := fs.String("base", "", "directory with the previous state/ and bitmaps/, kept for TLDs no input owns")
	or := fs.Bool("or", false, "bootstrap mode: merge state per label and OR bitmaps instead of taking each TLD from its owner")
	requireComplete := fs.Bool("require-complete", false, "fail unless every TLD in -tlds has a 1-3 character sweep and a fully swept 4 character bitmap")
	tlds := fs.String("tlds", "", "TLDs -require-complete expects (default: all supported)")
	maxJump := fs.Float64("max-jump", 0.5, "reject a TLD's new output if its count of free names changes by more than this fraction (0 = off)")
	minJump := fs.Int("min-jump", 200, "...and by at least this many names")
	fs.Parse(args)
	if fs.NArg() == 0 {
		return fmt.Errorf("usage: shortfree merge [flags] <dir>...")
	}
	if *or && *base != "" {
		return fmt.Errorf("merge: -base and -or exclude each other")
	}

	var st *store.State
	var bms map[string]store.Bitmap
	var covs map[string]*store.Coverage
	var err error
	if *or {
		st, bms, covs, err = mergeOr(fs.Args())
	} else {
		st, bms, covs, err = mergeOwned(*base, fs.Args(), jumpLimit{*maxJump, *minJump})
	}
	if err != nil {
		return err
	}
	if *requireComplete {
		if err := checkComplete(st, bms, covs, parseTLDs(*tlds)); err != nil {
			return err
		}
	}
	for _, tld := range sortedKeys(bms) {
		path := filepath.Join(*bitmaps, tld+".bin")
		if err := store.WriteAtomic(path, bms[tld]); err != nil {
			return err
		}
		if c := covs[tld]; c != nil {
			if err := c.Save(store.CoveragePath(path)); err != nil {
				return err
			}
		}
		fmt.Fprintf(os.Stderr, "merge: .%s 4 character bitmap has %d free names\n", tld, bms[tld].Count())
	}
	for _, tld := range sortedKeys(st.TLDs) {
		fmt.Fprintf(os.Stderr, "merge: .%s %d state entries\n", tld, len(st.TLDs[tld]))
	}
	return st.Save(*statePath)
}

// part is one input directory.
type part struct {
	dir     string
	state   *store.State
	bitmaps map[string]store.Bitmap
	covs    map[string]*store.Coverage // only for bitmaps that have a .cov.json
}

func loadPart(dir string) (*part, error) {
	s, err := store.Load(filepath.Join(dir, "state", "state.json.gz"))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", dir, err)
	}
	p := &part{dir: dir, state: s, bitmaps: map[string]store.Bitmap{}, covs: map[string]*store.Coverage{}}
	bins, err := filepath.Glob(filepath.Join(dir, "bitmaps", "*.bin"))
	if err != nil {
		return nil, err
	}
	for _, path := range bins {
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		// A truncated bitmap must not silently become an empty one.
		if want := (names.Space(4) + 7) / 8; len(b) != want {
			return nil, fmt.Errorf("%s: %d bytes, want %d", path, len(b), want)
		}
		tld := strings.TrimSuffix(filepath.Base(path), ".bin")
		p.bitmaps[tld] = store.Bitmap(b)
		// A bitmap without coverage file predates coverage records and
		// counts as fully swept at the state's update time.
		if _, err := os.Stat(store.CoveragePath(path)); err == nil || !s.Updated.IsZero() {
			c, err := store.BitmapCoverage(path, 7, s.Updated)
			if err != nil {
				return nil, err
			}
			p.covs[tld] = c
		}
	}
	return p, nil
}

// jumpLimit decides when a TLD's new output is implausible.
type jumpLimit struct {
	Ratio float64 // fraction of the old count; <= 0 disables the check
	Min   int     // minimum absolute change
}

// implausible returns why a change from old to cur names is implausible,
// or "".
func (l jumpLimit) implausible(what string, old, cur int) string {
	if l.Ratio <= 0 {
		return ""
	}
	d := cur - old
	if d < 0 {
		d = -d
	}
	if d >= l.Min && float64(d) > l.Ratio*float64(old) {
		return fmt.Sprintf("%s changed from %d to %d in one run (limit %.0f%%, %d names); previous data kept", what, old, cur, l.Ratio*100, l.Min)
	}
	return ""
}

// notDelegated counts the entries DNS reported as not delegated.
func notDelegated(m map[string]*store.Entry) int {
	n := 0
	for _, e := range m {
		if e.DNS == "free" {
			n++
		}
	}
	return n
}

// shardBits counts set bits per shard (first character index mod shards).
func shardBits(bm store.Bitmap, shards int) []int {
	out := make([]int, shards)
	per := names.Space(3)
	for bi, x := range bm {
		for ; x != 0; x &= x - 1 {
			i := bi*8 + bits.TrailingZeros8(x)
			out[(i/per)%shards]++
		}
	}
	return out
}

// bitmapJump compares the shards both coverages say were swept.
func bitmapJump(l jumpLimit, old, cur store.Bitmap, oldCov, curCov *store.Coverage) string {
	if oldCov == nil || curCov == nil || oldCov.Shards != curCov.Shards {
		return "" // nothing comparable (first sweeps, legacy bitmap)
	}
	ob, cb := shardBits(old, oldCov.Shards), shardBits(cur, curCov.Shards)
	for i := 0; i < oldCov.Shards; i++ {
		if oldCov.Scanned(i) == nil || curCov.Scanned(i) == nil {
			continue
		}
		if why := l.implausible(fmt.Sprintf("free 4 character names in shard %d", i), ob[i], cb[i]); why != "" {
			return why
		}
	}
	return ""
}

func warn(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintln(os.Stderr, "merge: "+msg)
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		fmt.Println("::warning title=shortfree merge::" + msg)
	}
}

// checkComplete fails unless every TLD has a recorded 1-3 character sweep
// and, where 4 character names are registrable, a bitmap with every shard
// swept. The bootstrap uses it so an incomplete run never replaces data.
func checkComplete(st *store.State, bms map[string]store.Bitmap, covs map[string]*store.Coverage, want []string) error {
	var missing []string
	for _, tld := range want {
		if st.Scans[tld] == nil {
			missing = append(missing, "."+tld+" 1-3 characters")
		}
		if minLen(tld) > 4 {
			continue
		}
		if bms[tld] == nil || !covs[tld].Complete() {
			missing = append(missing, "."+tld+" 4 characters")
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("merge: incomplete, refusing to write: missing %s", strings.Join(missing, ", "))
	}
	return nil
}

// mergeOwned starts from base (may be "") and replaces every TLD an input
// owns with that input's copy, unless lim finds the copy implausible: then
// the base data stays, marked as held.
func mergeOwned(base string, dirs []string, lim jumpLimit) (*store.State, map[string]store.Bitmap, map[string]*store.Coverage, error) {
	out := &store.State{TLDs: map[string]map[string]*store.Entry{}}
	bms := map[string]store.Bitmap{}
	covs := map[string]*store.Coverage{}
	prev := &part{state: &store.State{TLDs: map[string]map[string]*store.Entry{}}, bitmaps: map[string]store.Bitmap{}, covs: map[string]*store.Coverage{}}
	if base != "" {
		b, err := loadPart(base)
		if err != nil {
			return nil, nil, nil, err
		}
		prev = b
		// Copies: the base maps must keep the previous data for comparison.
		out = &store.State{Updated: b.state.Updated, TLDs: map[string]map[string]*store.Entry{}, Scans: map[string]*store.Scan{}}
		for tld, m := range b.state.TLDs {
			out.TLDs[tld] = m
			if sc := b.state.ScanOf(tld); sc != nil {
				out.Scans[tld] = sc
			}
		}
		for tld, bm := range b.bitmaps {
			bms[tld] = bm
		}
		for tld, c := range b.covs {
			covs[tld] = c
		}
	}
	stateOwner, bitmapOwner := map[string]string{}, map[string]string{}
	for _, dir := range dirs {
		p, err := loadPart(dir)
		if err != nil {
			return nil, nil, nil, err
		}
		if p.state.Updated.After(out.Updated) {
			out.Updated = p.state.Updated
		}
		for tld, m := range p.state.TLDs {
			if prev, ok := stateOwner[tld]; ok {
				return nil, nil, nil, fmt.Errorf("merge: .%s state in both %s and %s", tld, prev, dir)
			}
			stateOwner[tld] = dir
			if old, ok := prev.state.TLDs[tld]; ok {
				if why := lim.implausible("not-delegated 1-3 character names", notDelegated(old), notDelegated(m)); why != "" {
					warn(".%s held: %s", tld, why)
					sc := *prev.state.ScanOf(tld)
					sc.Held = why
					out.PutScan(tld, &sc)
					continue
				}
			}
			out.TLDs[tld] = m
			sc := p.state.ScanOf(tld)
			out.PutScan(tld, sc)
		}
		for tld, b := range p.bitmaps {
			if prev, ok := bitmapOwner[tld]; ok {
				return nil, nil, nil, fmt.Errorf("merge: .%s bitmap in both %s and %s", tld, prev, dir)
			}
			bitmapOwner[tld] = dir
			if old, ok := prev.bitmaps[tld]; ok {
				if why := bitmapJump(lim, old, b, prev.covs[tld], p.covs[tld]); why != "" {
					warn(".%s 4 characters held: %s", tld, why)
					c := *prev.covs[tld]
					c.Held = why
					covs[tld] = &c
					continue
				}
			}
			bms[tld] = b
			if c := p.covs[tld]; c != nil {
				covs[tld] = c
			} else {
				delete(covs, tld)
			}
		}
	}
	return out, bms, covs, nil
}

// mergeOr is the bootstrap merge: per-label union of state, OR of bitmaps,
// union of shard coverage.
func mergeOr(dirs []string) (*store.State, map[string]store.Bitmap, map[string]*store.Coverage, error) {
	out := &store.State{TLDs: map[string]map[string]*store.Entry{}}
	bms := map[string]store.Bitmap{}
	covs := map[string]*store.Coverage{}
	for _, dir := range dirs {
		p, err := loadPart(dir)
		if err != nil {
			return nil, nil, nil, err
		}
		if p.state.Updated.After(out.Updated) {
			out.Updated = p.state.Updated
		}
		for tld, m := range p.state.TLDs {
			dst := out.TLD(tld)
			for label, e := range m {
				if old, ok := dst[label]; !ok || e.Checked > old.Checked {
					dst[label] = e
				}
			}
		}
		for tld, sc := range p.state.Scans {
			if old := out.Scans[tld]; old == nil || sc.At > old.At {
				out.PutScan(tld, sc)
			}
		}
		for tld, b := range p.bitmaps {
			if bms[tld] == nil {
				bms[tld] = store.NewBitmap(names.Space(4))
			}
			for i := range b {
				bms[tld][i] |= b[i]
			}
		}
		for tld, c := range p.covs {
			dst := covs[tld]
			if dst == nil {
				dst = &store.Coverage{Shards: c.Shards, Scans: map[string]*store.ShardScan{}}
				covs[tld] = dst
			}
			if dst.Shards != c.Shards {
				return nil, nil, nil, fmt.Errorf("merge: .%s coverage with %d and %d shards", tld, dst.Shards, c.Shards)
			}
			for k, s := range c.Scans {
				if old := dst.Scans[k]; old == nil || s.At > old.At {
					dst.Scans[k] = s
				}
			}
		}
	}
	return out, bms, covs, nil
}

func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
