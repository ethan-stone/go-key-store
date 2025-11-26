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

	val := []byte("111")

	walEntry := &WalEntryWrite{
		OpType:      Put,
		KeyLength:   3,
		ValueLength: 3,
		KeyBytes:    []byte("abc"),
		ValueBytes:  &val,
	}

	err := wal.Write(walEntry)

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

	val := []byte("111")

	walEntries := []*WalEntryWrite{
		{
			OpType:      Put,
			KeyLength:   2,
			ValueLength: 3,
			KeyBytes:    []byte("ab"),
			ValueBytes:  &val,
		},
		{
			OpType:      Del,
			KeyLength:   5,
			ValueLength: 0,
			KeyBytes:    []byte("abcde"),
			ValueBytes:  nil,
		},
	}

	expectedWalEntries := []*WalEntryRead{
		{
			Entry: &WalEntry{OpType: Put, KeyLength: 2, ValueLength: 3, KeyBytes: []byte("ab"), ValueBytes: &val},
			Size:  26,
		},
		{
			Entry: &WalEntry{OpType: Del, KeyLength: 5, ValueLength: 0, KeyBytes: []byte("abcde"), ValueBytes: nil},
			Size:  26,
		},
	}

	for _, walEntry := range walEntries {
		err := wal.Write(walEntry)
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
		if !bytes.Equal(expectedWalEntry.Entry.KeyBytes, readEntry.Entry.KeyBytes) {
			t.Errorf("Expected key bytes to be %s, got %s", string(expectedWalEntry.Entry.KeyBytes), string(readEntry.Entry.KeyBytes))
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

	val := []byte("111")

	walEntry := &WalEntryWrite{
		OpType:      Put,
		KeyLength:   2,
		ValueLength: 3,
		KeyBytes:    []byte("ab"),
		ValueBytes:  &val,
	}
	err := wal.Write(walEntry)

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
