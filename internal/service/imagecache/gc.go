package imagecache

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// fileEntry tracks one cached binary + its sidecar for LRU eviction.
type fileEntry struct {
	binPath  string
	sidePath string
	atime    int64 // unix seconds; fallback to mtime when atime unavailable
	size     int64
}

// DirSize returns the total size (bytes) of the cache root.
func DirSize(dir string) (int64, error) {
	var total int64
	err := filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		total += info.Size()
		return nil
	})
	return total, err
}

// EvictLRU walks the cache and removes the least-recently-accessed entries
// (paired file+sidecar) until total size is under maxBytes. On platforms
// without a portable atime we fall back to mtime.
func (s *Store) EvictLRU(ctx context.Context, maxBytes int64) error {
	entries, total, err := s.collect()
	if err != nil {
		return err
	}
	if total <= maxBytes {
		return nil
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].atime < entries[j].atime })
	for _, e := range entries {
		if total <= maxBytes {
			break
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		_ = os.Remove(e.binPath)
		_ = os.Remove(e.sidePath)
		total -= e.size
	}
	return nil
}

func (s *Store) collect() ([]fileEntry, int64, error) {
	var out []fileEntry
	var total int64
	err := filepath.WalkDir(s.Dir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		name := d.Name()
		if strings.HasSuffix(name, ".license.json") {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		atime := info.ModTime().Unix() // portable fallback
		base := strings.TrimSuffix(p, filepath.Ext(p))
		side := base + ".license.json"
		// Count the sidecar too: the cap is on the whole directory (DirSize),
		// and sidecars can outweigh small images.
		size := info.Size()
		if si, err := os.Stat(side); err == nil {
			size += si.Size()
		}
		out = append(out, fileEntry{
			binPath:  p,
			sidePath: side,
			atime:    atime,
			size:     size,
		})
		total += size
		return nil
	})
	return out, total, err
}
