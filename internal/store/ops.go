package store

import (
	"log"
	"sync"
)

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
