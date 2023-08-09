package filecache

import (
	"bytes"
	"os"
	"testing"
)

func TestGetFound(t *testing.T) {
	fc, err := New()
	if err != nil {
		t.Error("cannot create filecache")
	}

	key := "key"
	value := []byte("value")

	fc.Start()
	defer fc.Stop()

	err = fc.Put(key, value)
	if err != nil {
		t.Error("cannot put key")
	}

	res, err := fc.Get(key)
	if err != nil {
		t.Error("cannot get key")
	}

	if !bytes.Equal(res, value) {
		t.Errorf("result is wrong, %+v != %+v", res, value)
	}
}

func TestGetNotFound(t *testing.T) {
	fc, err := New()
	if err != nil {
		t.Error("cannot create filecache")
	}

	fc.Start()
	defer fc.Stop()

	key := "key"

	_, err = fc.Get(key)
	if err != ErrNotFound {
		t.Error("key shall not be found")
	}
}

func TestGetExpiredKey(t *testing.T) {
	fc, err := New(WithTTL(0))
	if err != nil {
		t.Error("cannot create filecache")
	}

	key := "key"
	value := []byte("value")

	fc.Start()
	defer fc.Stop()

	err = fc.Put(key, value)
	if err != nil {
		t.Error("cannot put key")
	}

	_, err = fc.Get(key)
	if err == nil {
		t.Error("shall not get key")
	}
}

func TestCleanExpiredKey(t *testing.T) {
	fc, err := New(WithTTL(0))
	if err != nil {
		t.Error("cannot create filecache")
	}

	key := "key"
	value := []byte("value")

	fc.Start()
	defer fc.Stop()

	err = fc.Put(key, value)
	if err != nil {
		t.Error("cannot put key")
	}

	fc.cleanExpiredKey()

	fc.lock.RLock()
	defer fc.lock.RUnlock()

	if len(fc.cache) != 0 {
		t.Error("cleaner shall clean cache")
	}
}

func TestCleanFileCache(t *testing.T) {
	fc, err := New()
	if err != nil {
		t.Error("cannot create filecache")
	}

	fc.Start()

	fileInfo, err := os.Stat(fc.workdir)
	if err != nil {
		t.Error("cannot stat workdir")
	}

	if !fileInfo.IsDir() {
		t.Error("workdir is not directory")
	}

	fc.Stop()

	_, err = os.Stat(fc.workdir)
	if !os.IsNotExist(err) {
		t.Error("workdir shall be removed")
	}
}
