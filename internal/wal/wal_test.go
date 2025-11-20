package wal

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestWrite(t *testing.T) {
	wal := NewWalWriter("test_write.bin")

	t.Cleanup(func() {
		os.Remove("test_write.bin")
	})

	val := []byte("111")

	walEntry := &WalEntry{
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
	wal := NewWalWriter("test_read.bin")

	t.Cleanup(func() {
		os.Remove("test_read.bin")
	})

	val := []byte("111")

	walEntries := []*WalEntry{
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
			entry: &WalEntry{OpType: Put, KeyLength: 2, ValueLength: 3, KeyBytes: []byte("ab"), ValueBytes: &val},
			size:  14,
		},
		{
			entry: &WalEntry{OpType: Del, KeyLength: 5, ValueLength: 0, KeyBytes: []byte("abcde"), ValueBytes: nil},
			size:  14,
		},
	}

	for _, walEntry := range walEntries {
		err := wal.Write(walEntry)
		if err != nil {
			t.Fatalf("Did not expect an error when writing")
		}
	}

	reader := NewWalReader("test_read.bin")

	offset := int64(0)

	for _, expectedWalEntry := range expectedWalEntries {
		readEntry, err := reader.Read(offset)

		if err != nil {
			t.Fatalf("Did not expect an error when reading: %v", err)
		}

		if readEntry.entry.OpType != expectedWalEntry.entry.OpType {
			t.Errorf("Expected op type to be %d, got %d", expectedWalEntry.entry.OpType, readEntry.entry.OpType)
		}
		if readEntry.entry.KeyLength != expectedWalEntry.entry.KeyLength {
			t.Errorf("Expected key length to be %d, got %d", expectedWalEntry.entry.KeyLength, readEntry.entry.KeyLength)
		}
		if readEntry.entry.ValueLength != expectedWalEntry.entry.ValueLength {
			t.Errorf("Expected value length to be %d, got %d", expectedWalEntry.entry.ValueLength, readEntry.entry.ValueLength)
		}
		if !bytes.Equal(expectedWalEntry.entry.KeyBytes, readEntry.entry.KeyBytes) {
			t.Errorf("Expected key bytes to be %s, got %s", string(expectedWalEntry.entry.KeyBytes), string(readEntry.entry.KeyBytes))
		}
		if expectedWalEntry.entry.ValueBytes != nil && readEntry.entry.ValueBytes != nil && !bytes.Equal(*readEntry.entry.ValueBytes, *expectedWalEntry.entry.ValueBytes) {
			t.Errorf("Expected value bytes to be %s, got %s", string(*expectedWalEntry.entry.ValueBytes), string(*readEntry.entry.ValueBytes))
		}
		if readEntry.size != expectedWalEntry.size {
			t.Errorf("Expected size to be %d, got %d", expectedWalEntry.size, readEntry.size)
		}

		offset += readEntry.size
	}

}

func TestShouldGetEOFWhenReadingPastEnd(t *testing.T) {
	wal := NewWalWriter("test_eof.bin")

	t.Cleanup(func() {
		os.Remove("test_eof.bin")
	})

	val := []byte("111")

	walEntry := &WalEntry{
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

	reader := NewWalReader("test_eof.bin")

	readEntry, err := reader.Read(0)

	if err != nil {
		t.Fatalf("Did not expect an error when reading: %v", err)
	}

	finalReadEntry, err := reader.Read(readEntry.size)

	if err != io.EOF {
		t.Errorf("Expected EOF, got %v", err)
	}

	if finalReadEntry != nil {
		t.Errorf("Expected nil, got %v", finalReadEntry)
	}

}
