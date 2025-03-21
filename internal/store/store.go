package store

import (
	"log"
	"sync"
)

type KeyValueStore struct {
	sync.RWMutex
	data map[string]string
}

type OpTypeEnum int

const (
	Get OpTypeEnum = iota
	Put
	Delete
)

type OpLogEntry struct {
	OpType OpTypeEnum
	Key    string
	Val    *string // Delete entries will not have the value
}

type OpLogEntries struct {
	sync.RWMutex
	entries []*OpLogEntry
}

func (opLog *OpLogEntries) AddEntry(entry *OpLogEntry) {
	opLog.Lock()
	defer opLog.Unlock()
	opLog.entries = append(opLog.entries, entry)

	var opTypeString string

	if entry.OpType == Get {
		opTypeString = "GET"
	} else if entry.OpType == Put {
		opTypeString = "PUT"
	} else {
		opTypeString = "DELETE"
	}

	if entry.OpType == Delete {
		log.Printf("%s %s", opTypeString, entry.Key)
	} else {
		log.Printf("%s %s %s", opTypeString, entry.Key, *entry.Val)
	}

}

var OpLog = &OpLogEntries{
	entries: make([]*OpLogEntry, 0),
}

func (store *KeyValueStore) Get(key string) (string, bool) {
	store.RLock()
	defer store.RUnlock()
	val, ok := store.data[key]

	// not finding a key is considered a no op
	if ok {
		defer OpLog.AddEntry(&OpLogEntry{
			OpType: Get,
			Key:    key,
			Val:    &val,
		})
	}

	return val, ok
}

func (store *KeyValueStore) Put(key string, val string) {
	store.Lock()
	defer store.Unlock()
	store.data[key] = val

	defer OpLog.AddEntry(&OpLogEntry{
		OpType: Put,
		Key:    key,
		Val:    &val,
	})
}

func (store *KeyValueStore) Del(key string) {
	store.Lock()
	defer store.Unlock()
	defer OpLog.AddEntry(&OpLogEntry{
		OpType: Delete,
		Key:    key,
		Val:    nil,
	})
	delete(store.data, key)
}

var Store = &KeyValueStore{
	data: make(map[string]string),
}
