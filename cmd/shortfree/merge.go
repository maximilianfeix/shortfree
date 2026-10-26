package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/maximilianfeix/shortfree/internal/names"
	"github.com/maximilianfeix/shortfree/internal/store"
)

// cmdMerge combines the outputs of parallel bootstrap jobs. Each input
// directory may hold state/state.json.gz and bitmaps/<tld>.bin. State entries
// are merged per label (the more recent RDAP check wins); bitmaps are OR-ed,
// which is correct because every job sweeps a disjoint shard from an empty
// bitmap.
func cmdMerge(args []string) error {
	fs := newFlags("merge")
	statePath := fs.String("state", "state/state.json.gz", "merged state file to write")
	bitmaps := fs.String("bitmaps", "bitmaps", "directory for merged bitmaps")
	fs.Parse(args)
	if fs.NArg() == 0 {
		return fmt.Errorf("usage: shortfree merge [flags] <dir>...")
	}

	out := &store.State{TLDs: map[string]map[string]*store.Entry{}}
	space := names.Space(4)
	merged := map[string]store.Bitmap{}
	for _, dir := range fs.Args() {
		s, err := store.Load(filepath.Join(dir, "state", "state.json.gz"))
		if err != nil {
			return fmt.Errorf("%s: %w", dir, err)
		}
		if s.Updated.After(out.Updated) {
			out.Updated = s.Updated
		}
		for tld, m := range s.TLDs {
			dst := out.TLD(tld)
			for label, e := range m {
				if old, ok := dst[label]; !ok || e.Checked > old.Checked {
					dst[label] = e
				}
			}
		}
		bins, _ := filepath.Glob(filepath.Join(dir, "bitmaps", "*.bin"))
		for _, path := range bins {
			tld := strings.TrimSuffix(filepath.Base(path), ".bin")
			b, err := store.LoadBitmap(path, space)
			if err != nil {
				return err
			}
			if merged[tld] == nil {
				merged[tld] = store.NewBitmap(space)
			}
			for i := range b {
				merged[tld][i] |= b[i]
			}
		}
	}
	for tld, b := range merged {
		if err := store.WriteAtomic(filepath.Join(*bitmaps, tld+".bin"), b); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "merge: .%s 4 character bitmap has %d free names\n", tld, b.Count())
	}
	for tld, m := range out.TLDs {
		fmt.Fprintf(os.Stderr, "merge: .%s %d state entries\n", tld, len(m))
	}
	return out.Save(*statePath)
}
