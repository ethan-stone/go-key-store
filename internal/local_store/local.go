package local_store

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/ethan-stone/go-key-store/internal/service"
	"github.com/ethan-stone/go-key-store/internal/wal"
)

type LocalKeyValueStore struct {
	data                      map[string]string
	dataMu                    sync.RWMutex
	walWriter                 wal.WalWriter
	walWriterMu               sync.Mutex
	lastAppliedSequenceNumber uint64
	cond                      *sync.Cond
}

func (store *LocalKeyValueStore) Get(key string) (*service.GetResult, error) {
	store.dataMu.RLock()
	defer store.dataMu.RUnlock()
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
	store.walWriterMu.Lock()
	defer store.walWriterMu.Unlock()

	keyBytes := []byte(key)
	valBytes := []byte(val)

	sequenceNumber, err := store.walWriter.Write(&wal.WalEntryWrite{
		OpType:      wal.Put,
		KeyLength:   int32(len(key)),
		ValueLength: int32(len(val)),
		KeyBytes:    &keyBytes,
		ValueBytes:  &valBytes,
	})

	if err != nil {
		return err
	}

	store.WaitForApply(sequenceNumber)

	return nil
}

func (store *LocalKeyValueStore) put(key string, val string) error {
	store.dataMu.Lock()
	defer store.dataMu.Unlock()

	store.data[key] = val
	return nil
}

func (store *LocalKeyValueStore) Delete(key string) error {
	store.walWriterMu.Lock()
	defer store.walWriterMu.Unlock()

	keyBytes := []byte(key)

	sequenceNumber, err := store.walWriter.Write(&wal.WalEntryWrite{
		OpType:      wal.Del,
		KeyLength:   int32(len(key)),
		ValueLength: 0,
		KeyBytes:    &keyBytes,
		ValueBytes:  nil,
	})

	if err != nil {
		return err
	}

	store.WaitForApply(sequenceNumber)

	return nil
}

func (store *LocalKeyValueStore) delete(key string) error {
	store.dataMu.Lock()
	defer store.dataMu.Unlock()

	delete(store.data, key)
	return nil
}

func (store *LocalKeyValueStore) ApplyWalEntry(entry *wal.WalEntry) error {

	switch entry.OpType {
	case wal.Put:
		store.put(string(*entry.KeyBytes), string(*entry.ValueBytes))
	case wal.Del:
		store.delete(string(*entry.KeyBytes))
	case wal.Snapshot:
		store.takeSnapshot(entry.SequenceNumber, "snapshot.bin")
	default:
		return fmt.Errorf("invalid OpType: %d", entry.OpType)
	}

	store.lastAppliedSequenceNumber = entry.SequenceNumber

	store.cond.L.Lock()
	store.cond.Broadcast()
	store.cond.L.Unlock()

	return nil
}

func (store *LocalKeyValueStore) SetWalWriter(walWriter wal.WalWriter) {
	store.walWriter = walWriter
}

// Example snapshot
// sequenceNumber is the sequence number of the last WAL entry that is included in the snapshot.
// entryCount is the number of entries in the snapshot.
// keyLength is the length of the key in bytes.
// valueLength is the length of the value in bytes.
// keyBytes is the bytes of the key.
// valueBytes is the bytes of the value.
// [sequenceNumber (8 bytes)][entryCount (8 bytes)][keyLength (4 bytes)][valueLength (4 bytes)][keyBytes (variable)][valueBytes (variable)]...

func (store *LocalKeyValueStore) TakeSnapshot() error {
	store.walWriterMu.Lock()
	defer store.walWriterMu.Unlock()

	sequenceNumber, err := store.walWriter.Write(&wal.WalEntryWrite{
		OpType:      wal.Snapshot,
		KeyLength:   0,
		ValueLength: 0,
		KeyBytes:    nil,
		ValueBytes:  nil,
	})

	if err != nil {
		return err
	}

	store.WaitForApply(sequenceNumber)

	fmt.Println("Snapshot complete. Latest sequence number ->", sequenceNumber)

	return nil
}

func (store *LocalKeyValueStore) takeSnapshot(lastSequenceNumber uint64, path string) error {
	store.dataMu.RLock()
	defer store.dataMu.RUnlock()

	tempFile, err := os.CreateTemp("", "snapshot-*.bin")

	if err != nil {
		return err
	}

	if err := binary.Write(tempFile, binary.LittleEndian, lastSequenceNumber); err != nil {
		return fmt.Errorf("writing last sequence number: %w", err)
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

	if err := store.walWriter.Rotate(); err != nil {
		return fmt.Errorf("rotate failed: %w", err)
	}

	CleanupOldWals(store.walWriter.GetDirectory())

	log.Printf("Snapshot written to %s. Last applied sequence number -> %d", path, lastSequenceNumber)

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

func (store *LocalKeyValueStore) Snapshot() error {
	log.Println("Starting snapshot...")

	if err := store.TakeSnapshot(); err != nil {
		return fmt.Errorf("snapshot failed: %w", err)
	}

	log.Println("Snapshot complete and WAL rotated.")
	return nil
}

func (store *LocalKeyValueStore) StartSnapshotManager(interval time.Duration, maxWalSize int64) {
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
				log.Printf("WAL size (%d MB) exceeds limit; snapshotting and rotating...",
					info.Size()/1024/1024)
				if err := store.Snapshot(); err != nil {
					fmt.Println("checkpoint error:", err)
				}
			}
		}
	}()
}

func (store *LocalKeyValueStore) SubscribeToWalEntries() {
	go func() {
		for entry := range store.walWriter.Subscribe() {
			err := store.ApplyWalEntry(entry)
			if err != nil {
				fmt.Println("error applying wal entry:", err)
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

		store.put(string(key), string(val))
	}
	store.lastAppliedSequenceNumber = sequenceNumber
	return nil
}

func (store *LocalKeyValueStore) GetLastAppliedSequenceNumber() uint64 {
	return store.lastAppliedSequenceNumber
}

func (store *LocalKeyValueStore) Close() error {
	return store.walWriter.Close()
}

func (store *LocalKeyValueStore) WaitForApply(target uint64) {
	store.cond.L.Lock()
	for store.lastAppliedSequenceNumber < target {
		store.cond.Wait()
	}
	store.cond.L.Unlock()
}

func (store *LocalKeyValueStore) GetWalWriter() wal.WalWriter {
	return store.walWriter
}

var Store *LocalKeyValueStore

type InitializeLocalKeyValueStoreConfig struct {
	WalWriter wal.WalWriter
}

func InitializeLocalKeyValueStore(config *InitializeLocalKeyValueStoreConfig) *LocalKeyValueStore {

	Store = &LocalKeyValueStore{
		data:                      make(map[string]string),
		dataMu:                    sync.RWMutex{},
		walWriter:                 config.WalWriter,
		walWriterMu:               sync.Mutex{},
		lastAppliedSequenceNumber: 0,
		cond:                      sync.NewCond(&sync.Mutex{}),
	}

	return Store
}
