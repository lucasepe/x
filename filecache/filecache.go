package filecache

import (
	"container/heap"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// File-backed cache suitable for CLI tools. It stores each item as two
// files in a directory: <hash>.data and <hash>.meta (JSON). Keys are
// arbitrary strings (e.g. request URLs). All operations are concurrent-safe.

var (
	ErrNotFound   = errors.New("item not found")
	ErrInvalidKey = errors.New("invalid key")
)

// Stats contains aggregate statistics for the cache.
type Stats struct {
	Count      int       `json:"count"`
	TotalBytes int64     `json:"total_bytes"`
	LastAccess time.Time `json:"last_access"` // most recent access among all items
}

// entryMeta is stored on disk alongside the data file.
type entryMeta struct {
	Key        string    `json:"key"`
	Size       int64     `json:"size"`
	LastAccess time.Time `json:"last_access"`
}

// FileCacheFS is the file-backed cache.
type FileCacheFS struct {
	dir      string
	mu       sync.RWMutex
	index    map[string]*entryMeta // map[hash]meta
	maxItems int
	maxBytes int64
	curBytes int64
	pq       priorityQueue
}

// New returns a FileCacheFS that stores cache files under dir. It will
// create the directory if missing.
func New(dir string) (*FileCacheFS, error) {
	if dir == "" {
		return nil, ErrInvalidKey
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	c := &FileCacheFS{
		dir:   dir,
		index: make(map[string]*entryMeta),
	}
	// load existing entries
	if err := c.loadIndex(); err != nil {
		return nil, err
	}
	// initialize heap (already populated by loadIndex)
	heap.Init(&c.pq)
	return c, nil
}

// ConfigureLimits sets maximum items and bytes. Zero means unlimited.
func (c *FileCacheFS) ConfigureLimits(maxItems int, maxBytes int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.maxItems = maxItems
	c.maxBytes = maxBytes
	// enforce in case limits are lower than current usage
	c.enforceLimits()
}

// Index returns a copy of the current index (hash -> entryMeta copy).
// This is handy for debugging / listing contents.
func (c *FileCacheFS) Index() map[string]entryMeta {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make(map[string]entryMeta, len(c.index))
	for h, m := range c.index {
		out[h] = *m
	}
	return out
}

func (c *FileCacheFS) Dir() string {
	return c.dir
}

// hashKey returns the hex sha256 of the key.
func hashKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

func (c *FileCacheFS) dataPath(hash string) string {
	return filepath.Join(c.dir, hash+".data")
}
func (c *FileCacheFS) metaPath(hash string) string {
	return filepath.Join(c.dir, hash+".meta")
}

// loadIndex scans the directory and loads existing .meta files.
func (c *FileCacheFS) loadIndex() error {
	entries, err := os.ReadDir(c.dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		name := e.Name()
		if filepath.Ext(name) != ".meta" {
			continue
		}
		full := filepath.Join(c.dir, name)
		b, err := os.ReadFile(full)
		if err != nil {
			// skip malformed entries
			continue
		}
		var m entryMeta
		if err := json.Unmarshal(b, &m); err != nil {
			continue
		}
		hash := name[:len(name)-len(".meta")]
		// keep pointer copy
		mm := m
		c.index[hash] = &mm
		c.curBytes += mm.Size
		heap.Push(&c.pq, &pqItem{hash: hash, lastAccess: mm.LastAccess.UnixNano()})
	}
	return nil
}

// Put stores bytes for key. It's atomic: write to temp then rename.
func (c *FileCacheFS) Put(key string, data []byte) error {
	if key == "" {
		return ErrInvalidKey
	}
	h := hashKey(key)
	dataPath := c.dataPath(h)
	metaPath := c.metaPath(h)

	// write data to temp
	tmpData := dataPath + ".tmp"
	if err := os.WriteFile(tmpData, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmpData, dataPath); err != nil {
		_ = os.Remove(tmpData)
		return err
	}

	m := &entryMeta{Key: key, Size: int64(len(data)), LastAccess: time.Now()}
	mb, _ := json.Marshal(m)
	tmpMeta := metaPath + ".tmp"
	if err := os.WriteFile(tmpMeta, mb, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmpMeta, metaPath); err != nil {
		_ = os.Remove(tmpMeta)
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	// adjust curBytes if replacing existing
	if old, ok := c.index[h]; ok {
		c.curBytes -= old.Size
	}
	c.index[h] = m
	c.curBytes += m.Size
	heap.Push(&c.pq, &pqItem{hash: h, lastAccess: m.LastAccess.UnixNano()})
	c.enforceLimits()
	return nil
}

// Get returns the cached bytes for key. It updates LastAccess.
func (c *FileCacheFS) Get(key string) ([]byte, error) {
	if key == "" {
		return nil, ErrInvalidKey
	}
	h := hashKey(key)

	c.mu.RLock()
	m, ok := c.index[h]
	c.mu.RUnlock()
	if !ok {
		return nil, ErrNotFound
	}

	dataPath := c.dataPath(h)
	b, err := os.ReadFile(dataPath)
	if err != nil {
		// file may be missing on disk — treat as not found and remove index
		c.mu.Lock()
		delete(c.index, h)
		c.mu.Unlock()
		return nil, ErrNotFound
	}

	// update last access
	now := time.Now()
	c.mu.Lock()
	m.LastAccess = now
	// write meta atomically (best-effort)
	if mb, err := json.Marshal(m); err == nil {
		_ = os.WriteFile(c.metaPath(h)+".tmp", mb, 0o644)
		_ = os.Rename(c.metaPath(h)+".tmp", c.metaPath(h))
	}
	heap.Push(&c.pq, &pqItem{hash: h, lastAccess: m.LastAccess.UnixNano()})
	c.mu.Unlock()

	return b, nil
}

// Del removes the cached item for key.
func (c *FileCacheFS) Del(key string) error {
	if key == "" {
		return ErrInvalidKey
	}
	h := hashKey(key)
	c.mu.Lock()
	defer c.mu.Unlock()
	m, ok := c.index[h]
	if !ok {
		return ErrNotFound
	}
	_ = os.Remove(c.dataPath(h))
	_ = os.Remove(c.metaPath(h))
	delete(c.index, h)
	c.curBytes -= m.Size
	return nil
}

// Clean removes all cached items from disk and clears the index.
func (c *FileCacheFS) Clean() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	entries, err := os.ReadDir(c.dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		name := e.Name()
		if filepath.Ext(name) == ".data" || filepath.Ext(name) == ".meta" {
			_ = os.Remove(filepath.Join(c.dir, name))
		}
	}
	c.index = make(map[string]*entryMeta)
	c.curBytes = 0
	c.pq = priorityQueue{}
	heap.Init(&c.pq)
	return nil
}

// Stats computes and returns basic statistics.
func (c *FileCacheFS) Stats() (Stats, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var s Stats
	var last time.Time
	for _, m := range c.index {
		s.Count++
		s.TotalBytes += m.Size
		if m.LastAccess.After(last) {
			last = m.LastAccess
		}
	}
	s.LastAccess = last
	return s, nil
}

// StreamPut lets you write from an io.Reader directly to the cache (useful for large responses).
func (c *FileCacheFS) StreamPut(key string, r io.Reader) error {
	if key == "" {
		return ErrInvalidKey
	}
	h := hashKey(key)
	dataPath := c.dataPath(h)
	metaPath := c.metaPath(h)
	tmp := dataPath + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	written, err := io.Copy(f, r)
	if err != nil {
		_ = os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, dataPath); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	m := &entryMeta{Key: key, Size: written, LastAccess: time.Now()}
	mb, _ := json.Marshal(m)
	tmpMeta := metaPath + ".tmp"
	if err := os.WriteFile(tmpMeta, mb, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmpMeta, metaPath); err != nil {
		_ = os.Remove(tmpMeta)
		return err
	}

	c.mu.Lock()
	if old, ok := c.index[h]; ok {
		c.curBytes -= old.Size
	}
	c.index[h] = m
	c.curBytes += written
	heap.Push(&c.pq, &pqItem{hash: h, lastAccess: m.LastAccess.UnixNano()})
	c.enforceLimits()
	c.mu.Unlock()
	return nil
}

// enforceLimits removes least recently used items until limits are respected.
// It tolerates stale entries in the heap (lazy deletion): when popping an
// item, it checks that the popped timestamp matches the current meta; if not
// it skips it.
func (c *FileCacheFS) enforceLimits() {
	for (c.maxItems > 0 && len(c.index) > c.maxItems) || (c.maxBytes > 0 && c.curBytes > c.maxBytes) {
		if c.pq.Len() == 0 {
			return
		}
		old := heap.Pop(&c.pq).(*pqItem)
		m, ok := c.index[old.hash]
		if !ok {
			// already removed
			continue
		}
		// if timestamps differ, this pq entry is stale
		if m.LastAccess.UnixNano() != old.lastAccess {
			continue
		}
		// remove actual entry
		_ = os.Remove(c.dataPath(old.hash))
		_ = os.Remove(c.metaPath(old.hash))
		delete(c.index, old.hash)
		c.curBytes -= m.Size
	}
}

// priorityQueue implements a min-heap by LastAccess
type pqItem struct {
	hash       string
	lastAccess int64 // unixnano
}

type priorityQueue []*pqItem

func (pq priorityQueue) Len() int { return len(pq) }
func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].lastAccess < pq[j].lastAccess
}
func (pq priorityQueue) Swap(i, j int)       { pq[i], pq[j] = pq[j], pq[i] }
func (pq *priorityQueue) Push(x interface{}) { *pq = append(*pq, x.(*pqItem)) }
func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}
