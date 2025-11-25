package store

import (
	"os"
	"testing"

	"github.com/ethan-stone/go-key-store/internal/wal"
)

func TestPut(t *testing.T) {
	store := &LocalKeyValueStore{
		data: make(map[string]string),
		wal: wal.NewWalWriter(&wal.WalWriterConfig{
			FileName: "wal.bin",
			SyncMode: wal.SyncModeAlways,
		}),
	}

	err := store.Put("a", "b")

	if err != nil {
		t.Fatalf("Did not expect an error when putting into store %v", err)
	}

	t.Cleanup(func() {
		os.Remove("wal.bin")
	})
}

func TestGetShouldReturnNotOkWhenKeyNotFound(t *testing.T) {
	store := &LocalKeyValueStore{
		data: make(map[string]string),
		wal: wal.NewWalWriter(&wal.WalWriterConfig{
			FileName: "wal.bin",
			SyncMode: wal.SyncModeAlways,
		}),
	}

	r, err := store.Get("a")

	if err != nil {
		t.Fatalf("Did not expect an error when getting store %v", err)
	}

	if r.Ok {
		t.Errorf("Did not expect to find key %v", "a")
	}

	t.Cleanup(func() {
		os.Remove("wal.bin")
	})
}

func TestShouldReturnOkWhenKeyFound(t *testing.T) {
	store := &LocalKeyValueStore{
		data: make(map[string]string),
		wal: wal.NewWalWriter(&wal.WalWriterConfig{
			FileName: "wal.bin",
			SyncMode: wal.SyncModeAlways,
		}),
	}

	err := store.Put("a", "b")

	if err != nil {
		t.Fatalf("Did not expect an error when putting into store %v", err)
	}

	r, err := store.Get("a")

	if err != nil {
		t.Fatalf("Did not expect an error when getting from store %v", err)
	}

	if !r.Ok {
		t.Errorf("Did not expect to not find key %s", "a")
	}

	t.Cleanup(func() {
		os.Remove("wal.bin")
	})
}

func TestShouldDelete(t *testing.T) {
	store := &LocalKeyValueStore{
		data: make(map[string]string),
		wal: wal.NewWalWriter(&wal.WalWriterConfig{
			FileName: "wal.bin",
			SyncMode: wal.SyncModeAlways,
		}),
	}

	err := store.Put("a", "b")

	if err != nil {
		t.Fatalf("Did not expect an error when putting into store %v", err)
	}

	r, err := store.Get("a")

	if err != nil {
		t.Fatalf("Did not expect an error when getting from store %v", err)
	}

	if !r.Ok {
		t.Errorf("Did not expect to not find key %s", "a")
	}

	err = store.Delete("a")

	if err != nil {
		t.Fatalf("Did not expect an error when deleting from store %v", err)
	}

	r, err = store.Get("a")

	if err != nil {
		t.Fatalf("Did not expect an error when getting from store %v", err)
	}

	if r.Ok {
		t.Errorf("Did not expect to find key %s", "a")
	}

	t.Cleanup(func() {
		os.Remove("wal.bin")
	})
}
