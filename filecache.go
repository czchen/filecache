/*
The filecache package provides a simple file based cache with TTL (time-to-live) supported.
*/
package filecache

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type FileCache struct {
	opts        options
	workdir     string
	stopCleaner chan struct{}

	cache map[string]item
	lock  sync.RWMutex
}

// Cannot find cache for the key.
var ErrNotFound = fmt.Errorf("not found")

type item struct {
	expiredAt time.Time
	file      *os.File
}

// Create a new FileCache.
func New(opts ...Option) (*FileCache, error) {
	workdir, err := os.MkdirTemp("", "filecache-*")
	if err != nil {
		return nil, err
	}

	fc := &FileCache{
		workdir:     workdir,
		stopCleaner: make(chan struct{}),
		cache:       make(map[string]item),
	}

	fc.opts = getDefaultOptions()
	for _, opt := range opts {
		opt(&fc.opts)
	}

	return fc, nil
}

// Start this FileCache. This will start a cleaner goroutine to clean expired key periodically. User needs to call
// Stop afterward.
func (fc *FileCache) Start() {
	go fc.runCleaner()
}

// Stop this FileCache. Once this function is called, no further functions shall be called.
func (fc *FileCache) Stop() {
	fc.stopCleaner <- struct{}{}

	fc.lock.Lock()
	for key := range fc.cache {
		delete(fc.cache, key)
	}
	fc.lock.Unlock()

	os.RemoveAll(fc.workdir)
}

// Get the value for the given key. ErrNotFound will be returned if the key is not found.
func (fc *FileCache) Get(key string, value io.Writer) error {
	fc.lock.RLock()
	item, ok := fc.cache[key]
	fc.lock.RUnlock()

	if !ok {
		return ErrNotFound
	}

	fc.lock.Lock()
	defer fc.lock.Unlock()

	if time.Now().After(item.expiredAt) {
		delete(fc.cache, key)
		return ErrNotFound
	}

	item.expiredAt = time.Now().Add(fc.opts.timeToLive)

	item.file.Seek(0, io.SeekStart)
	_, err := io.Copy(value, item.file)
	if err != nil {
		return err
	}

	return nil
}

// Put the value for the given key.
func (fc *FileCache) Put(key string, value io.Reader) error {
	f, err := os.CreateTemp(fc.workdir, "cache-*")
	if err != nil {
		return err
	}

	i := item{
		expiredAt: time.Now().Add(fc.opts.timeToLive),
		file:      f,
	}

	_, err = io.Copy(i.file, value)
	if err != nil {
		return err
	}

	fc.lock.Lock()
	fc.cache[key] = i
	fc.lock.Unlock()

	return nil
}

func (fc *FileCache) runCleaner() {
	ticker := time.NewTicker(fc.opts.cleanerInterval)
	defer ticker.Stop()

	select {
	case <-ticker.C:
		fc.cleanExpiredKey()

	case <-fc.stopCleaner:
		break
	}
}

func (fc *FileCache) cleanExpiredKey() {
	now := time.Now()
	expiredKeys := make([]string, 0)

	fc.lock.RLock()
	for key, item := range fc.cache {
		if item.expiredAt.Before(now) {
			expiredKeys = append(expiredKeys, key)
		}
	}
	fc.lock.RUnlock()

	fc.lock.Lock()
	for _, key := range expiredKeys {
		if fc.cache[key].expiredAt.Before(now) {
			delete(fc.cache, key)
		}
	}
	fc.lock.Unlock()
}
