package store

import "sync"

type KeyValueStore struct {
	sync.RWMutex
	data map[string]string
}

func (store *KeyValueStore) Get(key string) (string, bool) {
	store.RLock()
	defer store.RUnlock()
	val, ok := store.data[key]
	return val, ok
}

func (store *KeyValueStore) Put(key string, val string) {
	store.Lock()
	defer store.Unlock()
	store.data[key] = val
}

func (store *KeyValueStore) Del(key string) {
	store.Lock()
	defer store.Unlock()
	delete(store.data, key)
}

var Store = &KeyValueStore{
	data: make(map[string]string),
}
