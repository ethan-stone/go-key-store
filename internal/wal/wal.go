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
	Rotate() error
	GetDirectory() string
	GetFile() *os.File
	GetSequenceNumber() uint64
	SetSequenceNumber(sequenceNumber uint64)
}

type FileWalWriter struct {
	file           *os.File
	index          int
	sequenceNumber uint64
	syncMode       int
	directory      string
}

const (
	Put         = 1
	Del         = 2
	Snapshot    = 3
	WalRotation = 4
)

// Example
// [sequenceNumber (8 bytes)][opType (1 byte)][keyLength (4 bytes)][valueLength (4 bytes)][keyBytes (variable)][valueBytes (variable)][checksum (4 bytes)]

type WalEntry struct {
	SequenceNumber uint64  // 8 bytes
	OpType         byte    // 1 for PUT. 2 for DEL. 3 for SNAPSHOT. 4. for WAL_ROTATION.
	KeyLength      int32   // 4 bytes // Will be 0 for SNAPSHOT and WAL_ROTATION.
	ValueLength    int32   // 4 bytes // Will be 0 for SNAPSHOT and WAL_ROTATION.
	KeyBytes       *[]byte // variable // Will be nil for SNAPSHOT and WAL_ROTATION.
	ValueBytes     *[]byte // variable // Will be nil for DEL, SNAPSHOT and WAL_ROTATION.
	CheckSum       uint32  // 4 bytes
}

type WalEntryWrite struct {
	OpType      byte    // 1 for PUT. 2 for DEL.
	KeyLength   int32   // 4 bytes
	ValueLength int32   // 4 bytes
	KeyBytes    *[]byte // variable
	ValueBytes  *[]byte // variable
}

const (
	SyncModeNone = iota
	SyncModePeriodic
	SyncModeAlways
)

type WalWriterConfig struct {
	Directory      string
	SyncMode       int
	Index          int
	SequenceNumber uint64
}

func NewFileWalWriter(config *WalWriterConfig) *FileWalWriter {
	if err := os.MkdirAll(config.Directory, 0755); err != nil {
		panic(fmt.Errorf("creating WAL directory: %w", err))
	}

	filename := fmt.Sprintf("%s/wal_%04d.bin", config.Directory, config.Index)
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)

	if err != nil {
		panic(err)
	}

	writer := &FileWalWriter{
		file:           file,
		directory:      config.Directory,
		index:          config.Index,
		syncMode:       config.SyncMode,
		sequenceNumber: config.SequenceNumber,
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
	switch entry.OpType {
	case Put:
		buf, err := writer.serializePutEntry(entry)
		if err != nil {
			return err
		}
		_, err = writer.file.Write(buf)
		if err != nil {
			return err
		}
	case Del:
		buf, err := writer.serializeDelEntry(entry)
		if err != nil {
			return err
		}
		_, err = writer.file.Write(buf)
		if err != nil {
			return err
		}
	case Snapshot:
		buf, err := writer.serializeSnapshotEntry()
		if err != nil {
			return err
		}
		_, err = writer.file.Write(buf)
		if err != nil {
			return err
		}
	case WalRotation:
		buf, err := writer.serializeWalRotationEntry()
		if err != nil {
			return err
		}
		_, err = writer.file.Write(buf)
		if err != nil {
			return err
		}
	}

	if writer.syncMode == SyncModeAlways {
		err := writer.file.Sync()
		if err != nil {
			return err
		}
	}

	writer.sequenceNumber++

	return nil
}

func (writer *FileWalWriter) serializePutEntry(entry *WalEntryWrite) ([]byte, error) {

	if entry.OpType != Put {
		return nil, fmt.Errorf("serializePutEntry can only be used for Put operations")
	}

	if entry.KeyBytes == nil {
		return nil, fmt.Errorf("KeyBytes must not be nil for Put operation")
	}

	if entry.ValueBytes == nil {
		return nil, fmt.Errorf("ValueBytes must not be nil for Put operation")
	}

	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, writer.sequenceNumber)
	binary.Write(buf, binary.LittleEndian, entry.OpType)
	binary.Write(buf, binary.LittleEndian, entry.KeyLength)
	binary.Write(buf, binary.LittleEndian, entry.ValueLength)
	binary.Write(buf, binary.LittleEndian, entry.KeyBytes)
	binary.Write(buf, binary.LittleEndian, entry.ValueBytes)

	checksum := crc32.ChecksumIEEE(buf.Bytes())

	binary.Write(buf, binary.LittleEndian, checksum)

	return buf.Bytes(), nil
}

func (writer *FileWalWriter) serializeDelEntry(entry *WalEntryWrite) ([]byte, error) {
	if entry.OpType != Del {
		return nil, fmt.Errorf("serializeDelEntry can only be used for Del operations")
	}

	if entry.KeyBytes == nil {
		return nil, fmt.Errorf("KeyBytes must not be nil for Del operation")
	}

	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, writer.sequenceNumber)
	binary.Write(buf, binary.LittleEndian, entry.OpType)
	binary.Write(buf, binary.LittleEndian, entry.KeyLength)
	binary.Write(buf, binary.LittleEndian, int32(0))
	binary.Write(buf, binary.LittleEndian, entry.KeyBytes)

	checksum := crc32.ChecksumIEEE(buf.Bytes())

	binary.Write(buf, binary.LittleEndian, checksum)

	return buf.Bytes(), nil
}

func (writer *FileWalWriter) serializeSnapshotEntry() ([]byte, error) {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, writer.sequenceNumber)
	binary.Write(buf, binary.LittleEndian, byte(3))  // SNAPSHOT OpType
	binary.Write(buf, binary.LittleEndian, int32(0)) // KeyLength
	binary.Write(buf, binary.LittleEndian, int32(0)) // ValueLength

	checksum := crc32.ChecksumIEEE(buf.Bytes())

	binary.Write(buf, binary.LittleEndian, checksum)

	return buf.Bytes(), nil
}

func (writer *FileWalWriter) serializeWalRotationEntry() ([]byte, error) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, writer.sequenceNumber)
	binary.Write(buf, binary.LittleEndian, byte(4))  // WAL_ROTATION OpType
	binary.Write(buf, binary.LittleEndian, int32(0)) // KeyLength
	binary.Write(buf, binary.LittleEndian, int32(0)) // ValueLength

	checksum := crc32.ChecksumIEEE(buf.Bytes())

	binary.Write(buf, binary.LittleEndian, checksum)

	return buf.Bytes(), nil
}

func (writer *FileWalWriter) Close() error {
	err := writer.file.Sync()

	if err != nil {
		return err
	}

	return writer.file.Close()
}

func (writer *FileWalWriter) Rotate() error {
	// Close the current WAL
	err := writer.Write(&WalEntryWrite{
		OpType:      WalRotation,
		KeyLength:   0,
		ValueLength: 0,
		KeyBytes:    nil,
		ValueBytes:  nil,
	})

	if err != nil {
		return err
	}

	writer.file.Sync()
	writer.file.Close()

	// Increment index
	writer.index++

	// Create new file
	newFileName := fmt.Sprintf("%s/wal_%04d.bin", writer.directory, writer.index)
	f, err := os.OpenFile(newFileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return fmt.Errorf("failed to rotate WAL: %w", err)
	}
	writer.file = f
	fmt.Println("Started new WAL:", newFileName)
	return nil
}

func (writer *FileWalWriter) GetDirectory() string {
	return writer.directory
}
func (writer *FileWalWriter) GetFile() *os.File {
	return writer.file
}

func (writer *FileWalWriter) GetSequenceNumber() uint64 {
	return writer.sequenceNumber
}

func (writer *FileWalWriter) SetSequenceNumber(sequenceNumber uint64) {
	writer.sequenceNumber = sequenceNumber
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

func (write *NoopWalWriter) Rotate() error {
	return nil
}

func (writer *NoopWalWriter) GetDirectory() string {
	return ""
}

func (writer *NoopWalWriter) GetFile() *os.File {
	return nil
}

func (writer *NoopWalWriter) GetSequenceNumber() uint64 {
	return 0
}

func (writer *NoopWalWriter) SetSequenceNumber(sequenceNumber uint64) {
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
	// Read sequence number
	sequenceNumberBuf := make([]byte, 8)
	_, err := reader.file.ReadAt(sequenceNumberBuf, offset)
	if err != nil {
		if err == io.EOF {
			return nil, io.EOF
		}
		return nil, err
	}
	sequenceNumber := binary.LittleEndian.Uint64(sequenceNumberBuf)
	cursor := offset + 8

	// Read OpType first (1 byte)
	opBuf := make([]byte, 1)
	_, err = reader.file.ReadAt(opBuf, cursor)
	if err != nil {
		if err == io.EOF {
			return nil, io.EOF
		}
		return nil, err
	}
	opType := opBuf[0]
	cursor += 1

	// Dispatch to proper deserializer
	switch opType {
	case Put:
		return reader.deserializePutEntry(sequenceNumber, cursor, offset)
	case Del:
		return reader.deserializeDelEntry(sequenceNumber, cursor, offset)
	case Snapshot:
		return reader.deserializeSnapshotEntry(sequenceNumber, cursor, offset)
	case WalRotation:
		return reader.deserializeWalRotationEntry(sequenceNumber, cursor, offset)
	default:
		return nil, fmt.Errorf("invalid OpType: %d", opType)
	}
}

func (reader *FileWalReader) deserializePutEntry(seq uint64, cursor int64, baseOffset int64) (*WalEntryRead, error) {
	var (
		keyLenBuf   = make([]byte, 4)
		valueLenBuf = make([]byte, 4)
		checksumBuf = make([]byte, 4)
	)

	// Read key length and value length (int32 each)
	_, err := reader.file.ReadAt(keyLenBuf, cursor)
	if err != nil {
		return nil, err
	}
	keyLen := int32(binary.LittleEndian.Uint32(keyLenBuf))
	cursor += 4

	_, err = reader.file.ReadAt(valueLenBuf, cursor)
	if err != nil {
		return nil, err
	}
	valueLen := int32(binary.LittleEndian.Uint32(valueLenBuf))
	cursor += 4

	// Read key bytes and value bytes
	keyBytes := make([]byte, keyLen)
	_, err = reader.file.ReadAt(keyBytes, cursor)
	if err != nil {
		return nil, err
	}
	cursor += int64(keyLen)

	valueBytes := make([]byte, valueLen)
	_, err = reader.file.ReadAt(valueBytes, cursor)
	if err != nil {
		return nil, err
	}
	cursor += int64(valueLen)

	// Read checksum
	_, err = reader.file.ReadAt(checksumBuf, cursor)
	if err != nil {
		return nil, err
	}
	storedChecksum := binary.LittleEndian.Uint32(checksumBuf)

	// Verify checksum
	entrySize := 8 + 1 + 4 + 4 + int64(keyLen) + int64(valueLen)
	entryBuf := make([]byte, entrySize)
	_, err = reader.file.ReadAt(entryBuf, baseOffset)
	if err != nil {
		return nil, err
	}
	computedChecksum := crc32.ChecksumIEEE(entryBuf)
	if storedChecksum != computedChecksum {
		return nil, fmt.Errorf("checksum mismatch")
	}

	totalSize := entrySize + 4 // add 4 for checksum
	return &WalEntryRead{
		Entry: &WalEntry{
			SequenceNumber: seq,
			OpType:         Put,
			KeyLength:      keyLen,
			ValueLength:    valueLen,
			KeyBytes:       &keyBytes,
			ValueBytes:     &valueBytes,
			CheckSum:       storedChecksum,
		},
		Size: totalSize,
	}, nil
}

func (reader *FileWalReader) deserializeDelEntry(seq uint64, cursor int64, baseOffset int64) (*WalEntryRead, error) {
	var (
		keyLenBuf   = make([]byte, 4)
		zeroValBuf  = make([]byte, 4)
		checksumBuf = make([]byte, 4)
	)

	// Read key length
	_, err := reader.file.ReadAt(keyLenBuf, cursor)
	if err != nil {
		return nil, err
	}

	keyLen := int32(binary.LittleEndian.Uint32(keyLenBuf))
	cursor += 4

	// Read the ValueLength (should be 0)
	_, err = reader.file.ReadAt(zeroValBuf, cursor)
	if err != nil {
		return nil, err
	}
	valLen := int32(binary.LittleEndian.Uint32(zeroValBuf))

	if valLen != 0 {
		return nil, fmt.Errorf("value length should be 0 for Del operation")
	}

	cursor += 4

	keyBytes := make([]byte, keyLen)
	_, err = reader.file.ReadAt(keyBytes, cursor)
	if err != nil {
		return nil, err
	}
	cursor += int64(keyLen)

	// Read checksum
	_, err = reader.file.ReadAt(checksumBuf, cursor)
	if err != nil {
		return nil, err
	}
	storedChecksum := binary.LittleEndian.Uint32(checksumBuf)

	// Verify checksum
	entrySize := 8 + 1 + 4 + 4 + int64(keyLen)
	entryBuf := make([]byte, entrySize)
	_, err = reader.file.ReadAt(entryBuf, baseOffset)
	if err != nil {
		return nil, err
	}
	computedChecksum := crc32.ChecksumIEEE(entryBuf)
	if storedChecksum != computedChecksum {
		return nil, fmt.Errorf("checksum mismatch")
	}

	totalSize := entrySize + 4
	return &WalEntryRead{
		Entry: &WalEntry{
			SequenceNumber: seq,
			OpType:         Del,
			KeyLength:      keyLen,
			ValueLength:    0,
			KeyBytes:       &keyBytes,
			CheckSum:       storedChecksum,
		},
		Size: totalSize,
	}, nil
}

func (reader *FileWalReader) deserializeSnapshotEntry(seq uint64, cursor int64, baseOffset int64) (*WalEntryRead, error) {
	checksumBuf := make([]byte, 4)

	// Skip the two int32s that are both zero
	cursor += 8

	_, err := reader.file.ReadAt(checksumBuf, cursor)
	if err != nil {
		return nil, err
	}

	storedChecksum := binary.LittleEndian.Uint32(checksumBuf)

	entrySize := 8 + 1 + 4 + 4 // seq + optype + 2 lengths
	entryBuf := make([]byte, entrySize)
	_, err = reader.file.ReadAt(entryBuf, baseOffset)
	if err != nil {
		return nil, err
	}

	computedChecksum := crc32.ChecksumIEEE(entryBuf)
	if storedChecksum != computedChecksum {
		return nil, fmt.Errorf("checksum mismatch")
	}

	return &WalEntryRead{
		Entry: &WalEntry{
			SequenceNumber: seq,
			OpType:         Snapshot,
			KeyLength:      0,
			ValueLength:    0,
			CheckSum:       storedChecksum,
		},
		Size: int64(entrySize + 4),
	}, nil
}

func (reader *FileWalReader) deserializeWalRotationEntry(seq uint64, cursor int64, baseOffset int64) (*WalEntryRead, error) {
	checksumBuf := make([]byte, 4)

	// Skip the two int32s that are both zero
	cursor += 8

	_, err := reader.file.ReadAt(checksumBuf, cursor)
	if err != nil {
		return nil, err
	}

	storedChecksum := binary.LittleEndian.Uint32(checksumBuf)

	entrySize := 8 + 1 + 4 + 4 // seq + optype + 2 lengths
	entryBuf := make([]byte, entrySize)
	_, err = reader.file.ReadAt(entryBuf, baseOffset)
	if err != nil {
		return nil, err
	}

	computedChecksum := crc32.ChecksumIEEE(entryBuf)
	if storedChecksum != computedChecksum {
		return nil, fmt.Errorf("checksum mismatch")
	}

	return &WalEntryRead{
		Entry: &WalEntry{
			SequenceNumber: seq,
			OpType:         WalRotation,
			KeyLength:      0,
			ValueLength:    0,
			CheckSum:       storedChecksum,
		},
		Size: int64(entrySize + 4),
	}, nil
}

func (reader *FileWalReader) Close() error {
	return reader.file.Close()
}
