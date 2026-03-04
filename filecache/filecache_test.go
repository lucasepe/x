package filecache

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func setupTempCache(t *testing.T) (*FileCacheFS, string) {
	t.Helper()
	dir := t.TempDir()
	cache, err := New(dir)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	return cache, dir
}

func TestPutGetDel(t *testing.T) {
	cache, _ := setupTempCache(t)

	key := "https://example.com/api/foo"
	data := []byte("bar")

	// Put
	if err := cache.Put(key, data); err != nil {
		t.Fatalf("Put error: %v", err)
	}

	// Get
	got, err := cache.Get(key)
	if err != nil {
		t.Fatalf("Get after Put error: %v", err)
	}
	if string(got) != string(data) {
		t.Errorf("Get returned %q, want %q", got, data)
	}

	// Del
	if err := cache.Del(key); err != nil {
		t.Fatalf("Del error: %v", err)
	}
	_, err = cache.Get(key)
	if err == nil {
		t.Errorf("Expected error after Del, got none")
	}
}

func TestClean(t *testing.T) {
	cache, _ := setupTempCache(t)

	keys := []string{"a", "b", "c"}
	for _, k := range keys {
		if err := cache.Put(k, []byte(k)); err != nil {
			t.Fatalf("Put %s failed: %v", k, err)
		}
	}

	// Clean
	if err := cache.Clean(); err != nil {
		t.Fatalf("Clean error: %v", err)
	}

	// Stats should reflect zero
	stats, err := cache.Stats()
	if err != nil {
		t.Fatalf("Stats error: %v", err)
	}
	if stats.Count != 0 {
		t.Errorf("Stats.Count = %d, want 0", stats.Count)
	}
	if stats.TotalBytes != 0 {
		t.Errorf("Stats.TotalBytes = %d, want 0", stats.TotalBytes)
	}
}

func TestConfigureLimitsEvictionLRU(t *testing.T) {
	cache, _ := setupTempCache(t)

	cache.ConfigureLimits(2, 1024) // small number of items

	// Put first two items
	if err := cache.Put("k1", []byte("data1")); err != nil {
		t.Fatalf("Put k1 failed: %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	if err := cache.Put("k2", []byte("data2")); err != nil {
		t.Fatalf("Put k2 failed: %v", err)
	}

	// Access k1 to make it more recent
	if _, err := cache.Get("k1"); err != nil {
		t.Fatalf("Get k1 error: %v", err)
	}

	// Put third item, causing eviction
	if err := cache.Put("k3", []byte("data3")); err != nil {
		t.Fatalf("Put k3 failed: %v", err)
	}

	// Now the cache should have exactly 2 items, and k2 should have been evicted.
	stats, _ := cache.Stats()
	if stats.Count != 2 {
		t.Errorf("After eviction, Count = %d, want 2", stats.Count)
	}

	// k2 should be gone
	if _, err := cache.Get("k2"); err == nil {
		t.Errorf("Expected k2 to be evicted, but Get succeeded")
	}

	// k1 and k3 should be present
	for _, key := range []string{"k1", "k3"} {
		if _, err := cache.Get(key); err != nil {
			t.Errorf("Expected %s present but Get failed: %v", key, err)
		}
	}
}

func TestStreamPut(t *testing.T) {
	cache, _ := setupTempCache(t)

	key := "stream"
	r := strings.NewReader("this is streamed_data")

	if err := cache.StreamPut(key, r); err != nil {
		t.Fatalf("StreamPut error: %v", err)
	}

	got, err := cache.Get(key)
	if err != nil {
		t.Fatalf("Get after StreamPut error: %v", err)
	}
	if string(got) != "this is streamed_data" {
		t.Errorf("Streamed data mismatch: got %q", got)
	}
}

func TestCorruptedMeta(t *testing.T) {
	cache, dir := setupTempCache(t)

	key := "x"
	data := []byte("y")
	if err := cache.Put(key, data); err != nil {
		t.Fatalf("Put error: %v", err)
	}

	// Corrupt metadata file
	h := hashKey(key)
	metaPath := filepath.Join(dir, h+".meta")
	if err := os.WriteFile(metaPath, []byte("not valid json"), 0o644); err != nil {
		t.Fatalf("Write corrupt meta error: %v", err)
	}

	// New cache instance should skip corrupted entry, not panic
	cache2, err := New(dir)
	if err != nil {
		t.Fatalf("New on existing corrupted meta failed: %v", err)
	}

	// That key should not be in index
	if _, err := cache2.Get(key); err == nil {
		t.Errorf("Expected Get to fail for corrupted entry %s", key)
	}
}

func TestInvalidKey(t *testing.T) {
	cache, _ := setupTempCache(t)

	if err := cache.Put("", []byte("val")); err == nil {
		t.Errorf("Put with empty key should error")
	}
	if _, err := cache.Get(""); err == nil {
		t.Errorf("Get with empty key should error")
	}
	if err := cache.Del(""); err == nil {
		t.Errorf("Del with empty key should error")
	}
}

func TestConcurrentAccess(t *testing.T) {
	cache, _ := setupTempCache(t)
	var wg sync.WaitGroup
	n := 50
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			k := fmt.Sprintf("k-%d", i%10)
			_ = cache.Put(k, []byte(fmt.Sprintf("v-%d", i)))
			_, _ = cache.Get(k)
		}(i)
	}
	wg.Wait()
}
