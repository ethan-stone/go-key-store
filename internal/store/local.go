package store

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/ethan-stone/go-key-store/internal/service"
	"github.com/ethan-stone/go-key-store/internal/wal"
)

type LocalKeyValueStore struct {
	sync.RWMutex
	data      map[string]string
	walWriter wal.WalWriter
}

func (store *LocalKeyValueStore) Get(key string) (*service.GetResult, error) {
	store.RLock()
	defer store.RUnlock()
	val, ok := store.data[key]

	if !ok {
		return &service.GetResult{
			Ok:  false,
			Val: "",
		}, nil
	}

	return &service.GetResult{
		Ok:  true,
		Val: val,
	}, nil
}

func (store *LocalKeyValueStore) Put(key string, val string) error {
	store.Lock()
	defer store.Unlock()

	keyBytes := []byte(key)
	valBytes := []byte(val)

	store.walWriter.Write(&wal.WalEntryWrite{
		OpType:      wal.Put,
		KeyLength:   int32(len(key)),
		ValueLength: int32(len(val)),
		KeyBytes:    &keyBytes,
		ValueBytes:  &valBytes,
	})

	store.data[key] = val

	return nil
}

func (store *LocalKeyValueStore) Delete(key string) error {
	store.Lock()
	defer store.Unlock()

	keyBytes := []byte(key)

	store.walWriter.Write(&wal.WalEntryWrite{
		OpType:      wal.Del,
		KeyLength:   int32(len(key)),
		ValueLength: 0,
		KeyBytes:    &keyBytes,
		ValueBytes:  nil,
	})

	delete(store.data, key)

	return nil
}

func (store *LocalKeyValueStore) SetWalWriter(walWriter wal.WalWriter) {
	Store.walWriter = walWriter
}

// Example snapshot
// sequenceNumber is the sequence number of the last WAL entry that is included in the snapshot.
// entryCount is the number of entries in the snapshot.
// keyLength is the length of the key in bytes.
// valueLength is the length of the value in bytes.
// keyBytes is the bytes of the key.
// valueBytes is the bytes of the value.
// [sequenceNumber (8 bytes)][entryCount (8 bytes)][keyLength (4 bytes)][valueLength (4 bytes)][keyBytes (variable)][valueBytes (variable)]...

func (store *LocalKeyValueStore) TakeSnapshot(path string) error {
	store.RLock()
	defer store.RUnlock()

	tempFile, err := os.CreateTemp("", "snapshot-*.bin")

	if err != nil {
		return err
	}

	err = store.walWriter.Write(&wal.WalEntryWrite{
		OpType:      wal.Snapshot,
		KeyLength:   0,
		ValueLength: 0,
		KeyBytes:    nil,
		ValueBytes:  nil,
	})

	if err != nil {
		return err
	}

	sequenceNumber := store.walWriter.GetSequenceNumber()

	if err := binary.Write(tempFile, binary.LittleEndian, sequenceNumber); err != nil {
		return fmt.Errorf("writing sequence number: %w", err)
	}

	count := uint64(len(store.data))

	if err := binary.Write(tempFile, binary.LittleEndian, count); err != nil {
		return fmt.Errorf("writing entry count: %w", err)
	}

	for k, v := range store.data {
		keyBytes, valBytes := []byte(k), []byte(v)
		binary.Write(tempFile, binary.LittleEndian, uint32(len(keyBytes)))
		binary.Write(tempFile, binary.LittleEndian, uint32(len(valBytes)))
		tempFile.Write(keyBytes)
		tempFile.Write(valBytes)
	}

	tempFile.Sync()
	tempFile.Close()

	if err := os.Rename(tempFile.Name(), path); err != nil {
		return fmt.Errorf("moving snapshot into place: %w", err)
	}

	fmt.Println("Snapshot written ->", path)
	return nil
}

func CleanupOldWals(dir string) {
	files, _ := filepath.Glob(filepath.Join(dir, "wal_*.bin"))
	if len(files) == 1 {
		return
	}

	sort.Strings(files)

	for _, f := range files[:len(files)-1] {
		os.Remove(f)
		fmt.Println("Deleted old WAL:", f)
	}
}

func (store *LocalKeyValueStore) CheckpointAndRotate() error {
	fmt.Println("Starting checkpoint...")

	// 1. Write the snapshot
	if err := store.TakeSnapshot("snapshot.bin"); err != nil {
		return fmt.Errorf("snapshot failed: %w", err)
	}

	// 2. Rotate WAL to a new file
	if err := store.walWriter.Rotate(); err != nil {
		return fmt.Errorf("rotate failed: %w", err)
	}

	// 3. Clean up older WALs
	CleanupOldWals(store.walWriter.GetDirectory())

	fmt.Println("Checkpoint complete -> new WAL started")
	return nil
}

func (store *LocalKeyValueStore) StartCheckpointManager(interval time.Duration, maxWalSize int64) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			if store.walWriter.GetFile() == nil {
				continue
			}

			info, err := store.walWriter.GetFile().Stat()

			if err != nil {
				continue
			}

			if info.Size() > maxWalSize {
				fmt.Printf("WAL size (%d MB) exceeds limit; checkpointing...\n",
					info.Size()/1024/1024)
				if err := store.CheckpointAndRotate(); err != nil {
					fmt.Println("checkpoint error:", err)
				}
			}
		}
	}()
}

func (store *LocalKeyValueStore) LoadFromSnapshot(path string) error {
	f, err := os.Open(path)

	if os.IsNotExist(err) {
		return nil
	}

	if err != nil {
		return err
	}

	defer f.Close()

	var sequenceNumber uint64
	if err := binary.Read(f, binary.LittleEndian, &sequenceNumber); err != nil {
		return fmt.Errorf("reading sequence number: %w", err)
	}

	var count uint64
	if err := binary.Read(f, binary.LittleEndian, &count); err != nil {
		return err
	}

	for i := uint64(0); i < count; i++ {
		var kLen, vLen uint32
		binary.Read(f, binary.LittleEndian, &kLen)
		binary.Read(f, binary.LittleEndian, &vLen)

		key := make([]byte, kLen)
		val := make([]byte, vLen)
		io.ReadFull(f, key)
		io.ReadFull(f, val)

		store.Put(string(key), string(val))
	}
	return nil
}

func (store *LocalKeyValueStore) Close() error {
	return store.walWriter.Close()
}

var Store *LocalKeyValueStore

type InitializeLocalKeyValueStoreConfig struct {
	WalWriter wal.WalWriter
}

func InitializeLocalKeyValueStore(config *InitializeLocalKeyValueStoreConfig) *LocalKeyValueStore {

	Store = &LocalKeyValueStore{
		data:      make(map[string]string),
		walWriter: config.WalWriter,
	}

	return Store
}
