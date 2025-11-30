package wal

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestWrite(t *testing.T) {
	wal := NewFileWalWriter(&WalWriterConfig{
		Directory:      "wals",
		SyncMode:       SyncModeAlways,
		Index:          0,
		SequenceNumber: 0,
	})

	t.Cleanup(func() {
		os.Remove("wals/wal_0000.bin")
	})

	key := []byte("abc")
	val := []byte("111")

	walEntry := &WalEntryWrite{
		OpType:      Put,
		KeyLength:   3,
		ValueLength: 3,
		KeyBytes:    &key,
		ValueBytes:  &val,
	}

	_, err := wal.Write(walEntry)

	if err != nil {
		t.Fatalf("Did not expect an error when writing")
	}

}

func TestRead(t *testing.T) {
	wal := NewFileWalWriter(&WalWriterConfig{
		Directory:      "wals",
		SyncMode:       SyncModeAlways,
		Index:          0,
		SequenceNumber: 0,
	})

	t.Cleanup(func() {
		os.Remove("wals/wal_0000.bin")
	})

	key1 := []byte("ab")
	key2 := []byte("abcde")
	val := []byte("111")

	walEntries := []*WalEntryWrite{
		{
			OpType:      Put,
			KeyLength:   2,
			ValueLength: 3,
			KeyBytes:    &key1,
			ValueBytes:  &val,
		},
		{
			OpType:      Del,
			KeyLength:   5,
			ValueLength: 0,
			KeyBytes:    &key2,
			ValueBytes:  nil,
		},
		{
			OpType:      Snapshot,
			KeyLength:   0,
			ValueLength: 0,
			KeyBytes:    nil,
			ValueBytes:  nil,
		},
		{
			OpType:      WalRotation,
			KeyLength:   0,
			ValueLength: 0,
			KeyBytes:    nil,
			ValueBytes:  nil,
		},
	}

	expectedWalEntries := []*WalEntryRead{
		{
			Entry: &WalEntry{OpType: Put, KeyLength: 2, ValueLength: 3, KeyBytes: &key1, ValueBytes: &val},
			Size:  8 + 1 + 4 + 4 + 2 + 3 + 4,
		},
		{
			Entry: &WalEntry{OpType: Del, KeyLength: 5, ValueLength: 0, KeyBytes: &key2, ValueBytes: nil},
			Size:  8 + 1 + 4 + 4 + 5 + 4,
		},
		{
			Entry: &WalEntry{OpType: Snapshot, KeyLength: 0, ValueLength: 0, KeyBytes: nil, ValueBytes: nil},
			Size:  8 + 1 + 4 + 4 + 4,
		},
		{
			Entry: &WalEntry{OpType: WalRotation, KeyLength: 0, ValueLength: 0, KeyBytes: nil, ValueBytes: nil},
			Size:  8 + 1 + 4 + 4 + 4,
		},
	}

	for _, walEntry := range walEntries {
		_, err := wal.Write(walEntry)
		if err != nil {
			t.Fatalf("Did not expect an error when writing")
		}
	}

	reader := NewFileWalReader("wals/wal_0000.bin")

	offset := int64(0)

	for _, expectedWalEntry := range expectedWalEntries {
		readEntry, err := reader.Read(offset)

		if err != nil {
			t.Fatalf("Did not expect an error when reading: %v", err)
		}

		if readEntry.Entry.OpType != expectedWalEntry.Entry.OpType {
			t.Errorf("Expected op type to be %d, got %d", expectedWalEntry.Entry.OpType, readEntry.Entry.OpType)
		}
		if readEntry.Entry.KeyLength != expectedWalEntry.Entry.KeyLength {
			t.Errorf("Expected key length to be %d, got %d", expectedWalEntry.Entry.KeyLength, readEntry.Entry.KeyLength)
		}
		if readEntry.Entry.ValueLength != expectedWalEntry.Entry.ValueLength {
			t.Errorf("Expected value length to be %d, got %d", expectedWalEntry.Entry.ValueLength, readEntry.Entry.ValueLength)
		}
		if expectedWalEntry.Entry.KeyBytes != nil && readEntry.Entry.KeyBytes != nil && !bytes.Equal(*expectedWalEntry.Entry.KeyBytes, *readEntry.Entry.KeyBytes) {
			t.Errorf("Expected key bytes to be %s, got %s", string(*expectedWalEntry.Entry.KeyBytes), string(*readEntry.Entry.KeyBytes))
		}
		if expectedWalEntry.Entry.ValueBytes != nil && readEntry.Entry.ValueBytes != nil && !bytes.Equal(*readEntry.Entry.ValueBytes, *expectedWalEntry.Entry.ValueBytes) {
			t.Errorf("Expected value bytes to be %s, got %s", string(*expectedWalEntry.Entry.ValueBytes), string(*readEntry.Entry.ValueBytes))
		}
		if readEntry.Size != expectedWalEntry.Size {
			t.Errorf("Expected size to be %d, got %d", expectedWalEntry.Size, readEntry.Size)
		}

		offset += readEntry.Size
	}

}

func TestShouldGetEOFWhenReadingPastEnd(t *testing.T) {
	wal := NewFileWalWriter(&WalWriterConfig{
		Directory:      "wals",
		SyncMode:       SyncModeAlways,
		Index:          0,
		SequenceNumber: 0,
	})

	t.Cleanup(func() {
		os.Remove("wals/wal_0000.bin")
	})

	key := []byte("ab")
	val := []byte("111")

	walEntry := &WalEntryWrite{
		OpType:      Put,
		KeyLength:   2,
		ValueLength: 3,
		KeyBytes:    &key,
		ValueBytes:  &val,
	}
	_, err := wal.Write(walEntry)

	if err != nil {
		t.Fatalf("Did not expect an error when writing: %v", err)
	}

	reader := NewFileWalReader("wals/wal_0000.bin")

	readEntry, err := reader.Read(0)

	if err != nil {
		t.Fatalf("Did not expect an error when reading: %v", err)
	}

	finalReadEntry, err := reader.Read(readEntry.Size)

	if err != io.EOF {
		t.Errorf("Expected EOF, got %v", err)
	}

	if finalReadEntry != nil {
		t.Errorf("Expected nil, got %v", finalReadEntry)
	}

}
