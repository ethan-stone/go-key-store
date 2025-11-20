package wal

import (
	"encoding/binary"
	"fmt"
	"os"
	"sync"
)

type Wal struct {
	sync.RWMutex
	file *os.File
}

const (
	Put = 1
	Del = 2
)

type WalEntry struct {
	OpType byte // 1 for PUT. 2 for DEL.
}

func NewWal(fileName string) *Wal {
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)

	if err != nil {
		panic(err)
	}

	return &Wal{
		file: file,
	}
}

func (wal *Wal) Write(entry *WalEntry) error {
	wal.Lock()
	defer wal.Unlock()

	err := binary.Write(wal.file, binary.LittleEndian, entry.OpType)

	if err != nil {
		panic(err)
	}

	err = wal.file.Sync()

	if err != nil {
		panic(err)
	}

	fmt.Println("Wrote wal entry")

	return nil
}

type WalEntryRead struct {
	entry *WalEntry
	size  int64
}

func (wal *Wal) Read(offset int64) (*WalEntryRead, error) {
	wal.RLock()
	defer wal.RUnlock()

	buf := make([]byte, 1)

	_, err := wal.file.ReadAt(buf, offset)

	if err != nil {
		panic(err)
	}

	opType := buf[0]

	if opType != Put && opType != Del {
		return nil, fmt.Errorf("invalid op type: %d", opType)
	}

	return &WalEntryRead{
		entry: &WalEntry{OpType: opType},
		size:  1,
	}, nil
}
