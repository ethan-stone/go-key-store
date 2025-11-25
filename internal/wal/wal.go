package wal

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"time"
)

type WalWriter interface {
	Write(entry *WalEntryWrite) error
	Close() error
}

type FileWalWriter struct {
	file   *os.File
	config *WalWriterConfig
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

const (
	SyncModeNone = iota
	SyncModePeriodic
	SyncModeAlways
)

type WalWriterConfig struct {
	FileName string
	SyncMode int
}

func NewFileWalWriter(config *WalWriterConfig) *FileWalWriter {
	file, err := os.OpenFile(config.FileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)

	if err != nil {
		panic(err)
	}

	writer := &FileWalWriter{
		file:   file,
		config: config,
	}

	if config.SyncMode == SyncModePeriodic {
		go func() {
			for {
				time.Sleep(time.Second * 1)
				writer.file.Sync()
			}
		}()
	}

	return writer
}

func (writer *FileWalWriter) Write(entry *WalEntryWrite) error {

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

	if writer.config.SyncMode == SyncModeAlways {
		err = writer.file.Sync()

		if err != nil {
			panic(err)
		}
	}

	if err != nil {
		panic(err)
	}

	return nil
}

func (writer *FileWalWriter) Close() error {
	err := writer.file.Sync()

	if err != nil {
		return err
	}

	return writer.file.Close()
}

type NoopWalWriter struct {
}

func NewNoopWalWriter() *NoopWalWriter {
	return &NoopWalWriter{}
}

func (writer *NoopWalWriter) Write(entry *WalEntryWrite) error {
	return nil
}

func (writer *NoopWalWriter) Close() error {
	return nil
}

type WalEntryRead struct {
	Entry *WalEntry
	Size  int64
}

type FileWalReader struct {
	file *os.File
}

func NewFileWalReader(fileName string) *FileWalReader {
	file, err := os.OpenFile(fileName, os.O_RDONLY, 0666)

	if err != nil {
		panic(err)
	}

	return &FileWalReader{
		file: file,
	}
}

func (reader *FileWalReader) Read(offset int64) (*WalEntryRead, error) {
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

	fmt.Printf("opType: %v\n", opType)
	fmt.Printf("keyLength: %v\n", keyLength)
	fmt.Printf("valueLength: %v\n", valueLength)

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
		fmt.Printf("%v\n", storedChecksum)
		fmt.Printf("%v\n", computedChecksum)
		return nil, fmt.Errorf("checksum mismatch")
	}

	return &WalEntryRead{
		Entry: &WalEntry{OpType: opType, KeyLength: int32(keyLength), ValueLength: int32(valueLength), KeyBytes: keyBytes, ValueBytes: &valueBytes, CheckSum: storedChecksum},
		Size:  headerSize + int64(keyLength) + int64(valueLength) + checksumSize,
	}, nil
}
