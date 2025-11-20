package wal

import (
	"encoding/binary"
	"fmt"
	"os"
)

type WalWriter struct {
	file *os.File
}

const (
	Put = 1
	Del = 2
)

type WalEntry struct {
	OpType    byte  // 1 for PUT. 2 for DEL.
	KeyLength int32 // 4 bytes
}

func NewWalWriter(fileName string) *WalWriter {
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)

	if err != nil {
		panic(err)
	}

	return &WalWriter{
		file: file,
	}
}

func (writer *WalWriter) Write(entry *WalEntry) error {
	err := binary.Write(writer.file, binary.LittleEndian, entry.OpType)

	if err != nil {
		panic(err)
	}

	err = binary.Write(writer.file, binary.LittleEndian, entry.KeyLength)

	if err != nil {
		panic(err)
	}

	err = writer.file.Sync()

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

type WalReader struct {
	file *os.File
}

func NewWalReader(fileName string) *WalReader {
	file, err := os.OpenFile(fileName, os.O_RDONLY, 0666)

	if err != nil {
		panic(err)
	}

	return &WalReader{
		file: file,
	}
}

func (reader *WalReader) Read(offset int64) (*WalEntryRead, error) {
	buf := make([]byte, 5)

	_, err := reader.file.ReadAt(buf, offset)

	if err != nil {
		panic(err)
	}

	opType := buf[0]
	keyLength := binary.LittleEndian.Uint32(buf[1:5])

	if opType != Put && opType != Del {
		return nil, fmt.Errorf("invalid op type: %d", opType)
	}

	return &WalEntryRead{
		entry: &WalEntry{OpType: opType, KeyLength: int32(keyLength)},
		size:  5,
	}, nil
}
