package wal

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
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
	OpType      byte    // 1 for PUT. 2 for DEL.
	KeyLength   int32   // 4 bytes
	ValueLength int32   // 4 bytes
	KeyBytes    []byte  // variable
	ValueBytes  *[]byte // variable
	CheckSum    uint32  // 4 bytes
}

type WalEntryWrite struct {
	OpType      byte    // 1 for PUT. 2 for DEL.
	KeyLength   int32   // 4 bytes
	ValueLength int32   // 4 bytes
	KeyBytes    []byte  // variable
	ValueBytes  *[]byte // variable
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

func (writer *WalWriter) Write(entry *WalEntryWrite) error {

	buf := new(bytes.Buffer)

	err := binary.Write(buf, binary.LittleEndian, entry.OpType)

	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, binary.LittleEndian, entry.KeyLength)

	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, binary.LittleEndian, entry.ValueLength)

	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, binary.LittleEndian, entry.KeyBytes)

	if err != nil {
		panic(err)
	}

	if entry.OpType == Put {
		if entry.ValueBytes == nil {
			return fmt.Errorf("ValueBytes must not be nil for Put operation")
		}

		err = binary.Write(buf, binary.LittleEndian, *entry.ValueBytes)

		if err != nil {
			panic(err)
		}
	}

	checksum := crc32.ChecksumIEEE(buf.Bytes())

	err = binary.Write(buf, binary.LittleEndian, checksum)

	if err != nil {
		panic(err)
	}

	_, err = writer.file.Write(buf.Bytes())

	if err != nil {
		panic(err)
	}

	err = writer.file.Sync()

	if err != nil {
		panic(err)
	}

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
	headerSize := int64(9)
	checksumSize := int64(4)
	headerBuffer := make([]byte, headerSize)

	_, err := reader.file.ReadAt(headerBuffer, offset)

	if err != nil {
		if err == io.EOF {
			return nil, io.EOF
		}

		panic(err)
	}

	opType := headerBuffer[0]
	keyLength := binary.LittleEndian.Uint32(headerBuffer[1:5])
	valueLength := binary.LittleEndian.Uint32(headerBuffer[5:9])

	if opType != Put && opType != Del {
		return nil, fmt.Errorf("invalid op type: %d", opType)
	}

	dataBuffer := make([]byte, keyLength+valueLength)

	_, err = reader.file.ReadAt(dataBuffer, offset+headerSize)

	if err != nil {
		if err == io.EOF {
			return nil, io.EOF
		}

		panic(err)
	}

	keyBytes := dataBuffer[0:keyLength]

	var valueBytes []byte = nil

	if valueLength > 0 {
		valueBytes = dataBuffer[keyLength : keyLength+valueLength]
	}

	checksumBuf := make([]byte, checksumSize)

	_, err = reader.file.ReadAt(checksumBuf, offset+headerSize+int64(keyLength)+int64(valueLength))

	if err != nil {
		if err == io.EOF {
			return nil, io.EOF
		}
		panic(err)
	}

	entryBuf := make([]byte, headerSize+int64(keyLength)+int64(valueLength))
	_, err = reader.file.ReadAt(entryBuf, offset)

	if err != nil {
		if err == io.EOF {
			return nil, io.EOF
		}
		panic(err)
	}

	storedChecksum := binary.LittleEndian.Uint32(checksumBuf)
	computedChecksum := crc32.ChecksumIEEE(entryBuf)

	if storedChecksum != computedChecksum {
		return nil, fmt.Errorf("checksum mismatch")
	}

	return &WalEntryRead{
		entry: &WalEntry{OpType: opType, KeyLength: int32(keyLength), ValueLength: int32(valueLength), KeyBytes: keyBytes, ValueBytes: &valueBytes, CheckSum: storedChecksum},
		size:  headerSize + int64(keyLength) + int64(valueLength) + checksumSize,
	}, nil
}
