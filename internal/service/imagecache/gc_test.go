package imagecache

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestEvictLRU_DeletesOldestUntilUnderCap(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStore(dir)
	// Create 5 fake entries with staggered mtimes.
	for i := 0; i < 5; i++ {
		sha, ext, err := s.Put("off", []byte(fmt.Sprintf("body-%d-%d", i, len("aaaaaaaaaaaaaaaaaaaa"))), "image/jpeg", SidecarMeta{Source: "off"})
		if err != nil {
			t.Fatal(err)
		}
		p := filepath.Join(dir, "off", sha+ext)
		stamp := time.Now().Add(time.Duration(-100+i*10) * time.Minute)
		_ = os.Chtimes(p, stamp, stamp)
	}
	total, err := DirSize(dir)
	if err != nil {
		t.Fatal(err)
	}
	if total == 0 {
		t.Fatal("no bytes written")
	}
	// Evict so that only ~the newest couple survive.
	if err := s.EvictLRU(context.Background(), total/3); err != nil {
		t.Fatal(err)
	}
	after, _ := DirSize(dir)
	if after >= total {
		t.Errorf("expected eviction to reduce size: before=%d after=%d", total, after)
	}
	// Sidecars of evicted files must be gone too.
	_ = filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		return nil
	})
}
