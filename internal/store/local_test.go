package store

import (
	"os"
	"testing"

	"github.com/ethan-stone/go-key-store/internal/wal"
	"github.com/google/uuid"
)

func TestPut(t *testing.T) {
	store := &LocalKeyValueStore{
		data: make(map[string]string),
		walWriter: wal.NewFileWalWriter(&wal.WalWriterConfig{
			Directory: "wals",
			SyncMode:  wal.SyncModeAlways,
		}),
	}

	err := store.Put("a", "b")

	if err != nil {
		t.Fatalf("Did not expect an error when putting into store %v", err)
	}

	t.Cleanup(func() {
		os.Remove("wals/wal_0000.bin")
	})
}

func TestGetShouldReturnNotOkWhenKeyNotFound(t *testing.T) {
	store := &LocalKeyValueStore{
		data: make(map[string]string),
		walWriter: wal.NewFileWalWriter(&wal.WalWriterConfig{
			Directory: "wals",
			SyncMode:  wal.SyncModeAlways,
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
		os.Remove("wals/wal_0000.bin")
	})
}

func TestShouldReturnOkWhenKeyFound(t *testing.T) {
	store := &LocalKeyValueStore{
		data: make(map[string]string),
		walWriter: wal.NewFileWalWriter(&wal.WalWriterConfig{
			Directory: "wals",
			SyncMode:  wal.SyncModeAlways,
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
		os.Remove("wals/wal_0000.bin")
	})
}

func TestShouldDelete(t *testing.T) {
	store := &LocalKeyValueStore{
		data: make(map[string]string),
		walWriter: wal.NewFileWalWriter(&wal.WalWriterConfig{
			Directory: "wals",
			SyncMode:  wal.SyncModeAlways,
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
		os.Remove("wals/wal_0000.bin")
	})
}

func TestShouldTakeSnapshot(t *testing.T) {
	randomSnapshotName := "snapshot_" + uuid.New().String() + ".bin"

	store := &LocalKeyValueStore{
		data:      make(map[string]string),
		walWriter: wal.NewNoopWalWriter(),
	}

	err := store.Put("a", "b")

	if err != nil {
		t.Fatalf("Did not expect an error when putting into store %v", err)
	}

	err = store.TakeSnapshot(randomSnapshotName)

	if err != nil {
		t.Fatalf("Did not expect an error when taking snapshot %v", err)
	}

	t.Cleanup(func() {
		os.Remove(randomSnapshotName)
	})
}

func TestShouldLoadFromSnapshot(t *testing.T) {
	randomSnapshotName := "snapshot_" + uuid.New().String() + ".bin"

	store := &LocalKeyValueStore{
		data:      make(map[string]string),
		walWriter: wal.NewNoopWalWriter(),
	}

	// write 100 random keys to the store
	keys := make([]string, 0, 100)
	for range 100 {
		key := uuid.New().String()
		keys = append(keys, key)
		val := uuid.New().String()
		if err := store.Put(key, val); err != nil {
			t.Fatalf("error on put: %v", err)
		}
	}

	err := store.TakeSnapshot(randomSnapshotName)

	if err != nil {
		t.Fatalf("Did not expect an error when taking snapshot %v", err)
	}

	newStore := &LocalKeyValueStore{
		data:      make(map[string]string),
		walWriter: wal.NewNoopWalWriter(),
	}

	err = newStore.LoadFromSnapshot(randomSnapshotName)

	if err != nil {
		t.Fatalf("Did not expect an error when loading from snapshot %v", err)
	}

	for i := range 100 {
		key := keys[i]

		originalStoreResult, err := store.Get(key)
		if err != nil {
			t.Fatalf("Did not expect an error when getting from store %v", err)
		}
		if !originalStoreResult.Ok {
			t.Errorf("Did not expect to not find key %s", key)
		}

		newStoreResult, err := newStore.Get(key)

		if err != nil {
			t.Fatalf("Did not expect an error when getting from store %v", err)
		}
		if !newStoreResult.Ok {
			t.Errorf("Did not expect to not find key %s", key)
		}
		if newStoreResult.Val != originalStoreResult.Val {
			t.Errorf("Did not expect to find key %s with value %s", key, originalStoreResult.Val)
		}
	}

	t.Cleanup(func() {
		os.Remove(randomSnapshotName)
	})
}
