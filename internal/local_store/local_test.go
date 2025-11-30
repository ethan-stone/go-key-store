package local_store

import (
	"os"
	"sync"
	"testing"

	"github.com/ethan-stone/go-key-store/internal/wal"
	"github.com/google/uuid"
)

func TestPut(t *testing.T) {
	randomWalDirectory := "wals_" + uuid.New().String()

	store := &LocalKeyValueStore{
		data:   make(map[string]string),
		dataMu: sync.RWMutex{},
		walWriter: wal.NewFileWalWriter(&wal.WalWriterConfig{
			Directory:      randomWalDirectory,
			SyncMode:       wal.SyncModeAlways,
			Index:          0,
			SequenceNumber: 0,
		}),
		walWriterMu:               sync.Mutex{},
		cond:                      sync.NewCond(&sync.Mutex{}),
		lastAppliedSequenceNumber: 0,
	}

	store.SubscribeToWalEntries()

	err := store.Put("a", "b")

	if err != nil {
		t.Fatalf("Did not expect an error when putting into store %v", err)
	}

	t.Cleanup(func() {
		os.RemoveAll(randomWalDirectory)
	})
}

func TestGetShouldReturnNotOkWhenKeyNotFound(t *testing.T) {
	randomWalDirectory := "wals_" + uuid.New().String()

	store := &LocalKeyValueStore{
		data:   make(map[string]string),
		dataMu: sync.RWMutex{},
		walWriter: wal.NewFileWalWriter(&wal.WalWriterConfig{
			Directory:      randomWalDirectory,
			SyncMode:       wal.SyncModeAlways,
			Index:          0,
			SequenceNumber: 0,
		}),
		walWriterMu:               sync.Mutex{},
		cond:                      sync.NewCond(&sync.Mutex{}),
		lastAppliedSequenceNumber: 0,
	}

	store.SubscribeToWalEntries()

	r, err := store.Get("a")

	if err != nil {
		t.Fatalf("Did not expect an error when getting store %v", err)
	}

	if r.Ok {
		t.Errorf("Did not expect to find key %v", "a")
	}

	t.Cleanup(func() {
		os.RemoveAll(randomWalDirectory)
	})
}

func TestShouldReturnOkWhenKeyFound(t *testing.T) {
	randomWalDirectory := "wals_" + uuid.New().String()

	store := &LocalKeyValueStore{
		data:   make(map[string]string),
		dataMu: sync.RWMutex{},
		walWriter: wal.NewFileWalWriter(&wal.WalWriterConfig{
			Directory:      randomWalDirectory,
			SyncMode:       wal.SyncModeAlways,
			Index:          0,
			SequenceNumber: 0,
		}),
		walWriterMu:               sync.Mutex{},
		cond:                      sync.NewCond(&sync.Mutex{}),
		lastAppliedSequenceNumber: 0,
	}

	store.SubscribeToWalEntries()

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
		os.RemoveAll(randomWalDirectory)
	})
}

func TestShouldDelete(t *testing.T) {
	randomWalDirectory := "wals_" + uuid.New().String()

	store := &LocalKeyValueStore{
		data:   make(map[string]string),
		dataMu: sync.RWMutex{},
		walWriter: wal.NewFileWalWriter(&wal.WalWriterConfig{
			Directory:      randomWalDirectory,
			SyncMode:       wal.SyncModeAlways,
			Index:          0,
			SequenceNumber: 0,
		}),
		walWriterMu:               sync.Mutex{},
		cond:                      sync.NewCond(&sync.Mutex{}),
		lastAppliedSequenceNumber: 0,
	}

	store.SubscribeToWalEntries()

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
		os.RemoveAll(randomWalDirectory)
	})
}

func TestShouldTakeSnapshot(t *testing.T) {
	randomWalDirectory := "wals_" + uuid.New().String()

	store := &LocalKeyValueStore{
		data:   make(map[string]string),
		dataMu: sync.RWMutex{},
		walWriter: wal.NewFileWalWriter(&wal.WalWriterConfig{
			Directory:      randomWalDirectory,
			SyncMode:       wal.SyncModeAlways,
			Index:          0,
			SequenceNumber: 0,
		}),
		walWriterMu:               sync.Mutex{},
		cond:                      sync.NewCond(&sync.Mutex{}),
		lastAppliedSequenceNumber: 0,
	}

	store.SubscribeToWalEntries()

	err := store.Put("a", "b")

	if err != nil {
		t.Fatalf("Did not expect an error when putting into store %v", err)
	}

	err = store.TakeSnapshot()

	if err != nil {
		t.Fatalf("Did not expect an error when taking snapshot %v", err)
	}

	t.Cleanup(func() {
		os.Remove("snapshot.bin")
		os.RemoveAll(randomWalDirectory)
	})
}

func TestShouldLoadFromSnapshot(t *testing.T) {
	randomWalDirectory := "wals_" + uuid.New().String()

	store := &LocalKeyValueStore{
		data:   make(map[string]string),
		dataMu: sync.RWMutex{},
		walWriter: wal.NewFileWalWriter(&wal.WalWriterConfig{
			Directory:      randomWalDirectory,
			SyncMode:       wal.SyncModeAlways,
			Index:          0,
			SequenceNumber: 0,
		}),
		walWriterMu:               sync.Mutex{},
		cond:                      sync.NewCond(&sync.Mutex{}),
		lastAppliedSequenceNumber: 0,
	}

	store.SubscribeToWalEntries()

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

	err := store.TakeSnapshot()

	if err != nil {
		t.Fatalf("Did not expect an error when taking snapshot %v", err)
	}

	newStore := &LocalKeyValueStore{
		data:                      make(map[string]string),
		dataMu:                    sync.RWMutex{},
		walWriter:                 wal.NewNoopWalWriter(),
		walWriterMu:               sync.Mutex{},
		cond:                      sync.NewCond(&sync.Mutex{}),
		lastAppliedSequenceNumber: 0,
	}

	newStore.SubscribeToWalEntries()

	err = newStore.LoadFromSnapshot("snapshot.bin")

	if err != nil {
		t.Fatalf("Did not expect an error when loading from snapshot %v", err)
	}

	if newStore.lastAppliedSequenceNumber != store.lastAppliedSequenceNumber {
		t.Errorf("Did not expect to have different last applied sequence numbers %d != %d", newStore.lastAppliedSequenceNumber, store.lastAppliedSequenceNumber)
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
		os.Remove("snapshot.bin")
		os.RemoveAll(randomWalDirectory)
	})
}
