package store

import (
	"io"
	"sync"

	"github.com/ethan-stone/go-key-store/internal/service"
	"github.com/ethan-stone/go-key-store/internal/wal"
)

type LocalKeyValueStore struct {
	sync.RWMutex
	data map[string]string
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

	OpLog.AddEntry(&OpLogEntry{
		OpType: Get,
		Key:    key,
		Val:    &val,
	})

	return &service.GetResult{
		Ok:  true,
		Val: val,
	}, nil
}

func (store *LocalKeyValueStore) Put(key string, val string) error {
	store.Lock()
	defer store.Unlock()
	store.data[key] = val

	defer OpLog.AddEntry(&OpLogEntry{
		OpType: Put,
		Key:    key,
		Val:    &val,
	})

	return nil
}

func (store *LocalKeyValueStore) Delete(key string) error {

	store.Lock()
	defer store.Unlock()
	defer OpLog.AddEntry(&OpLogEntry{
		OpType: Delete,
		Key:    key,
		Val:    nil,
	})
	delete(store.data, key)

	return nil
}

var Store *LocalKeyValueStore

func InitializeLocalKeyValueStore(walReader *wal.WalReader) *LocalKeyValueStore {

	offset := int64(0)

	Store = &LocalKeyValueStore{
		data: make(map[string]string),
	}

	for {
		entry, err := walReader.Read(offset)

		if err != nil {
			if err == io.EOF {
				break
			}

			panic(err)
		}

		if entry.Entry.OpType == wal.Put {
			Store.Put(string(entry.Entry.KeyBytes), string(*entry.Entry.ValueBytes))
		} else if entry.Entry.OpType == wal.Del {
			Store.Delete(string(entry.Entry.KeyBytes))
		}

		offset += entry.Size
	}

	return Store
}
