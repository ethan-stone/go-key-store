package store

import (
	"sync"

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

	valBytes := []byte(val)

	store.walWriter.Write(&wal.WalEntryWrite{
		OpType:      wal.Put,
		KeyLength:   int32(len(key)),
		ValueLength: int32(len(val)),
		KeyBytes:    []byte(key),
		ValueBytes:  &valBytes,
	})

	store.data[key] = val

	return nil
}

func (store *LocalKeyValueStore) Delete(key string) error {
	store.Lock()
	defer store.Unlock()

	store.walWriter.Write(&wal.WalEntryWrite{
		OpType:      wal.Del,
		KeyLength:   int32(len(key)),
		ValueLength: 0,
		KeyBytes:    []byte(key),
		ValueBytes:  nil,
	})

	delete(store.data, key)

	return nil
}

func (store *LocalKeyValueStore) SetWalWriter(walWriter wal.WalWriter) {
	Store.walWriter = walWriter
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
