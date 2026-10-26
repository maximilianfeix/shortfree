package main

import (
	"fmt"
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
func cmdMerge(args []string) error {
	fs := newFlags("merge")
	statePath := fs.String("state", "state/state.json.gz", "merged state file to write")
	bitmaps := fs.String("bitmaps", "bitmaps", "directory for merged bitmaps")
	base := fs.String("base", "", "directory with the previous state/ and bitmaps/, kept for TLDs no input owns")
	or := fs.Bool("or", false, "bootstrap mode: merge state per label and OR bitmaps instead of taking each TLD from its owner")
	fs.Parse(args)
	if fs.NArg() == 0 {
		return fmt.Errorf("usage: shortfree merge [flags] <dir>...")
	}
	if *or && *base != "" {
		return fmt.Errorf("merge: -base and -or exclude each other")
	}

	var st *store.State
	var bms map[string]store.Bitmap
	var err error
	if *or {
		st, bms, err = mergeOr(fs.Args())
	} else {
		st, bms, err = mergeOwned(*base, fs.Args())
	}
	if err != nil {
		return err
	}
	for _, tld := range sortedKeys(bms) {
		if err := store.WriteAtomic(filepath.Join(*bitmaps, tld+".bin"), bms[tld]); err != nil {
			return err
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
}

func loadPart(dir string) (*part, error) {
	s, err := store.Load(filepath.Join(dir, "state", "state.json.gz"))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", dir, err)
	}
	p := &part{dir: dir, state: s, bitmaps: map[string]store.Bitmap{}}
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
		p.bitmaps[strings.TrimSuffix(filepath.Base(path), ".bin")] = store.Bitmap(b)
	}
	return p, nil
}

// mergeOwned starts from base (may be "") and replaces every TLD an input
// owns with that input's copy.
func mergeOwned(base string, dirs []string) (*store.State, map[string]store.Bitmap, error) {
	out := &store.State{TLDs: map[string]map[string]*store.Entry{}}
	bms := map[string]store.Bitmap{}
	if base != "" {
		b, err := loadPart(base)
		if err != nil {
			return nil, nil, err
		}
		out, bms = b.state, b.bitmaps
	}
	stateOwner, bitmapOwner := map[string]string{}, map[string]string{}
	for _, dir := range dirs {
		p, err := loadPart(dir)
		if err != nil {
			return nil, nil, err
		}
		if p.state.Updated.After(out.Updated) {
			out.Updated = p.state.Updated
		}
		for tld, m := range p.state.TLDs {
			if prev, ok := stateOwner[tld]; ok {
				return nil, nil, fmt.Errorf("merge: .%s state in both %s and %s", tld, prev, dir)
			}
			stateOwner[tld] = dir
			out.TLDs[tld] = m
		}
		for tld, b := range p.bitmaps {
			if prev, ok := bitmapOwner[tld]; ok {
				return nil, nil, fmt.Errorf("merge: .%s bitmap in both %s and %s", tld, prev, dir)
			}
			bitmapOwner[tld] = dir
			bms[tld] = b
		}
	}
	return out, bms, nil
}

// mergeOr is the bootstrap merge: per-label union of state, OR of bitmaps.
func mergeOr(dirs []string) (*store.State, map[string]store.Bitmap, error) {
	out := &store.State{TLDs: map[string]map[string]*store.Entry{}}
	bms := map[string]store.Bitmap{}
	for _, dir := range dirs {
		p, err := loadPart(dir)
		if err != nil {
			return nil, nil, err
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
		for tld, b := range p.bitmaps {
			if bms[tld] == nil {
				bms[tld] = store.NewBitmap(names.Space(4))
			}
			for i := range b {
				bms[tld][i] |= b[i]
			}
		}
	}
	return out, bms, nil
}

func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
