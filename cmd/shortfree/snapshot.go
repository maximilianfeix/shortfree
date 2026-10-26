package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"github.com/maximilianfeix/shortfree/internal/store"
)

// A snapshot is one immutable, dated archive of a day's public data plus a
// manifest.json that describes it. The publish workflow uploads the archive
// under a new name every run and replaces manifest.json last, so a consumer
// that reads the manifest first always gets one complete, consistent run
// (never a mix of two runs while assets are being replaced).

// Manifest is manifest.json.
type Manifest struct {
	Schema  int       `json:"schema"`  // 1
	Name    string    `json:"name"`    // snapshot-YYYY-MM-DD[-suffix]
	Updated time.Time `json:"updated"` // scan time of the data
	Created time.Time `json:"created"`
	Archive FileSum   `json:"archive"`
	Files   []FileSum `json:"files"` // contents of the archive, under Name/
	Order   []string  `json:"order"`
	Missing []string  `json:"missing"`
	// MaxEvidenceAgeDays is the RDAP evidence age limit the data was
	// exported with.
	MaxEvidenceAgeDays float64              `json:"max_evidence_age_days"`
	TLDs               map[string]*TLDTotal `json:"tlds"`
	Source             *Source              `json:"source,omitempty"`
}

// FileSum is one file with its SHA-256.
type FileSum struct {
	Name   string `json:"name"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}

// TLDTotal summarises one TLD in the manifest.
type TLDTotal struct {
	Verified      int        `json:"rdap_verified"` // 1-3 chars, current RDAP 404
	Dropping      int        `json:"dropping"`
	Unverified    int        `json:"unverified"`
	Stale         int        `json:"stale"`
	DNSUnknown    int        `json:"dns_unknown"`
	Four          *int       `json:"four_dns_only,omitempty"` // nil: no 4 character data
	FourUnscanned int        `json:"four_unscanned"`
	Coverage      string     `json:"coverage"`
	ScannedAt     time.Time  `json:"scanned_at"`
	OldestCheck   *time.Time `json:"oldest_check,omitempty"`
}

var snapshotName = regexp.MustCompile(`^snapshot-\d{4}-\d{2}-\d{2}(-[0-9a-z]+)?$`)

func cmdSnapshot(args []string) error {
	fs := newFlags("snapshot")
	in := fs.String("in", "snapshot-src", "directory whose files go into the archive (flat)")
	name := fs.String("name", "", "snapshot name, snapshot-YYYY-MM-DD with optional -suffix (default: from the data's scan date)")
	statsPath := fs.String("stats", "data/stats.json", "stats.json of the exported data")
	coveragePath := fs.String("coverage", "data/coverage.json", "coverage.json of the exported data")
	out := fs.String("out", "snapshot", "output directory for <name>.tar.gz and manifest.json")
	fs.Parse(args)

	var stats Stats
	if err := readJSON(*statsPath, &stats); err != nil {
		return err
	}
	var cov CoverageReport
	if err := readJSON(*coveragePath, &cov); err != nil {
		return err
	}
	if *name == "" {
		*name = "snapshot-" + stats.Updated.UTC().Format("2006-01-02")
	}
	if !snapshotName.MatchString(*name) {
		return fmt.Errorf("snapshot: name %q, want snapshot-YYYY-MM-DD[-suffix]", *name)
	}
	m, archive, err := buildSnapshot(*in, *name, &stats, &cov)
	if err != nil {
		return err
	}
	if err := store.WriteAtomic(filepath.Join(*out, m.Archive.Name), archive); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(*out, "manifest.json"), m); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "snapshot: %s, %d files, %d bytes\n", m.Archive.Name, len(m.Files), m.Archive.Size)
	return nil
}

// buildSnapshot archives every regular file of dir (sorted, under name/,
// with the scan time as modification time, so equal input gives an equal
// archive) and returns the manifest and the archive bytes.
func buildSnapshot(dir, name string, stats *Stats, cov *CoverageReport) (*Manifest, []byte, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, err
	}
	m := &Manifest{
		Schema: 1, Name: name, Updated: stats.Updated, Created: time.Now().UTC().Truncate(time.Second),
		Order: stats.Order, Missing: stats.Missing, MaxEvidenceAgeDays: cov.MaxEvidenceAgeDays,
		TLDs: map[string]*TLDTotal{}, Source: cov.Source, Files: []FileSum{},
	}
	if m.Missing == nil {
		m.Missing = []string{}
	}
	var buf bytes.Buffer
	zw, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	tw := tar.NewWriter(zw)
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, e := range entries {
		if !e.Type().IsRegular() {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, nil, err
		}
		hdr := &tar.Header{Name: name + "/" + e.Name(), Mode: 0o644, Size: int64(len(b)), ModTime: stats.Updated.UTC(), Typeflag: tar.TypeReg, Format: tar.FormatPAX}
		if err := tw.WriteHeader(hdr); err != nil {
			return nil, nil, err
		}
		if _, err := tw.Write(b); err != nil {
			return nil, nil, err
		}
		m.Files = append(m.Files, sum(e.Name(), b))
	}
	if len(m.Files) == 0 {
		return nil, nil, fmt.Errorf("snapshot: %s is empty", dir)
	}
	if err := tw.Close(); err != nil {
		return nil, nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, nil, err
	}
	m.Archive = sum(name+".tar.gz", buf.Bytes())

	for _, tld := range stats.Order {
		ls := stats.TLDs[tld]
		if ls == nil {
			continue
		}
		t := &TLDTotal{}
		for _, n := range []string{"1", "2", "3"} {
			if l := ls[n]; l != nil {
				t.Verified += l.Available
				t.Dropping += l.Dropping
				t.Unverified += l.Unverified
				t.Stale += l.Stale
				t.DNSUnknown += l.Unknown
			}
		}
		if l := ls["4"]; l != nil && !l.Missing {
			four := l.Free
			t.Four, t.FourUnscanned = &four, l.Unscanned
			t.DNSUnknown += l.Unknown
		}
		if c := cov.TLDs[tld]; c != nil {
			t.Coverage, t.ScannedAt = c.Status, c.ScannedAt
			for _, n := range []string{"1", "2", "3"} {
				if lc := c.Lengths[n]; lc != nil && lc.OldestCheck != nil && (t.OldestCheck == nil || lc.OldestCheck.Before(*t.OldestCheck)) {
					t.OldestCheck = lc.OldestCheck
				}
			}
		}
		m.TLDs[tld] = t
	}
	return m, buf.Bytes(), nil
}

func sum(name string, b []byte) FileSum {
	h := sha256.Sum256(b)
	return FileSum{Name: name, Size: int64(len(b)), SHA256: hex.EncodeToString(h[:])}
}
