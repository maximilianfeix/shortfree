// Package store persists scan state: a JSON record per non-delegated 1-3
// character name, and a free-bitmap per TLD for 4 character names.
package store

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Entry is what we know about one name that is not delegated in DNS. Keys are
// short and times are Unix seconds because the state holds ~200k entries.
type Entry struct {
	DNS     string `json:"d"`           // "free" or "unknown"
	RDAP    string `json:"r,omitempty"` // rdap.Status, empty until verified
	Checked int64  `json:"c,omitempty"` // last RDAP check
	Since   int64  `json:"s,omitempty"` // first seen non-delegated
	Expires int64  `json:"e,omitempty"` // expiration date from RDAP
}

// CheckedAt returns Checked as a time.
func (e *Entry) CheckedAt() time.Time { return time.Unix(e.Checked, 0).UTC() }

// ExpiresAt returns Expires as a time, zero if unknown.
func (e *Entry) ExpiresAt() time.Time {
	if e.Expires == 0 {
		return time.Time{}
	}
	return time.Unix(e.Expires, 0).UTC()
}

// State holds 1-3 character results, keyed by TLD then label.
type State struct {
	Updated time.Time                    `json:"updated"`
	TLDs    map[string]map[string]*Entry `json:"tlds"`
}

// Load reads path (gzip-compressed if it ends in .gz), returning an empty
// state if it does not exist.
func Load(path string) (*State, error) {
	s := &State{TLDs: map[string]map[string]*Entry{}}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if strings.HasSuffix(path, ".gz") {
		zr, err := gzip.NewReader(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
		if b, err = io.ReadAll(zr); err != nil {
			return nil, err
		}
	}
	if err := json.Unmarshal(b, s); err != nil {
		return nil, err
	}
	if s.TLDs == nil {
		s.TLDs = map[string]map[string]*Entry{}
	}
	return s, nil
}

// Save writes the state atomically, gzip-compressed if path ends in .gz.
func (s *State) Save(path string) error {
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	if strings.HasSuffix(path, ".gz") {
		var buf bytes.Buffer
		zw, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
		zw.Write(b)
		if err := zw.Close(); err != nil {
			return err
		}
		b = buf.Bytes()
	}
	return WriteAtomic(path, b)
}

// TLD returns the entry map for tld, creating it.
func (s *State) TLD(tld string) map[string]*Entry {
	m, ok := s.TLDs[tld]
	if !ok {
		m = map[string]*Entry{}
		s.TLDs[tld] = m
	}
	return m
}

// Labels returns the labels of tld sorted by length, then alphabetically.
func (s *State) Labels(tld string) []string {
	m := s.TLDs[tld]
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	SortLabels(out)
	return out
}

// SortLabels sorts by length, then alphabetically.
func SortLabels(l []string) {
	sort.Slice(l, func(i, j int) bool {
		if len(l[i]) != len(l[j]) {
			return len(l[i]) < len(l[j])
		}
		return l[i] < l[j]
	})
}

// Bitmap is a fixed-size bitset indexed by names.Index.
type Bitmap []byte

// NewBitmap returns a zeroed bitmap with room for n bits.
func NewBitmap(n int) Bitmap { return make(Bitmap, (n+7)/8) }

func (b Bitmap) Get(i int) bool { return b[i>>3]&(1<<(i&7)) != 0 }

func (b Bitmap) Set(i int, v bool) {
	if v {
		b[i>>3] |= 1 << (i & 7)
	} else {
		b[i>>3] &^= 1 << (i & 7)
	}
}

// Count returns the number of set bits.
func (b Bitmap) Count() int {
	n := 0
	for _, x := range b {
		for ; x != 0; x &= x - 1 {
			n++
		}
	}
	return n
}

// LoadBitmap reads a bitmap of n bits, or returns an empty one if the file is
// missing or has the wrong size.
func LoadBitmap(path string, n int) (Bitmap, error) {
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) || (err == nil && len(b) != (n+7)/8) {
		return NewBitmap(n), nil
	}
	return Bitmap(b), err
}

// WriteAtomic writes data to path via a temp file and rename.
func WriteAtomic(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
