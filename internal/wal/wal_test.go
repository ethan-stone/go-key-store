package wal

import (
	"io"
	"os"
	"testing"
)

func TestWrite(t *testing.T) {
	wal := NewWalWriter("test_write.bin")

	walEntry := &WalEntry{
		OpType:      Put,
		KeyLength:   3,
		ValueLength: 3,
	}

	err := wal.Write(walEntry)

	if err != nil {
		t.Fatalf("Did not expect an error when writing")
	}

	// clean up wal file
	os.Remove("test_write.bin")
}

func TestRead(t *testing.T) {
	wal := NewWalWriter("test_read.bin")

	walEntries := []*WalEntry{
		{
			OpType:      Put,
			KeyLength:   2,
			ValueLength: 3,
		},
		{
			OpType:      Del,
			KeyLength:   5,
			ValueLength: 0,
		},
	}

	expectedWalEntries := []*WalEntryRead{
		{
			entry: &WalEntry{OpType: Put, KeyLength: 2, ValueLength: 3},
			size:  9,
		},
		{
			entry: &WalEntry{OpType: Del, KeyLength: 5, ValueLength: 0},
			size:  9,
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
		if readEntry.size != expectedWalEntry.size {
			t.Errorf("Expected size to be %d, got %d", expectedWalEntry.size, readEntry.size)
		}

		offset += readEntry.size
	}

	// clean up wal file
	os.Remove("test_read.bin")
}

func TestShouldGetEOFWhenReadingPastEnd(t *testing.T) {
	wal := NewWalWriter("test_eof.bin")

	walEntry := &WalEntry{
		OpType:      Put,
		KeyLength:   2,
		ValueLength: 3,
	}
	err := wal.Write(walEntry)

	if err != nil {
		t.Fatalf("Did not expect an error when writing")
	}

	reader := NewWalReader("test_eof.bin")

	readEntry, err := reader.Read(0)

	if err != nil {
		t.Fatalf("Did not expect an error when reading")
	}

	finalReadEntry, err := reader.Read(readEntry.size)

	if err != io.EOF {
		t.Errorf("Expected EOF, got %v", err)
	}

	if finalReadEntry != nil {
		t.Errorf("Expected nil, got %v", finalReadEntry)
	}

	os.Remove("test_eof.bin")
}
