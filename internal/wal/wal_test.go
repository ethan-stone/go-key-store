package wal

import (
	"os"
	"testing"
)

func TestWrite(t *testing.T) {
	wal := NewWalWriter("wal.bin")

	walEntry := &WalEntry{
		OpType:    Put,
		KeyLength: 3,
	}

	err := wal.Write(walEntry)

	if err != nil {
		t.Fatalf("Did not expect an error when writing")
	}

	// clean up wal file
	os.Remove("wal.bin")
}

func TestRead(t *testing.T) {
	wal := NewWalWriter("wal.bin")

	walEntries := []*WalEntry{
		{
			OpType:    Put,
			KeyLength: 2,
		},
		{
			OpType:    Del,
			KeyLength: 5,
		},
	}

	for _, walEntry := range walEntries {
		err := wal.Write(walEntry)
		if err != nil {
			t.Fatalf("Did not expect an error when writing")
		}
	}

	reader := NewWalReader("wal.bin")

	offset := int64(0)

	for _, walEntry := range walEntries {
		readEntry, err := reader.Read(offset)

		if err != nil {
			t.Fatalf("Did not expect an error when reading")
		}

		if readEntry.entry.OpType != walEntry.OpType {
			t.Errorf("Expected op type to be %d, got %d", walEntry.OpType, readEntry.entry.OpType)
		}
		if readEntry.entry.KeyLength != walEntry.KeyLength {
			t.Errorf("Expected key length to be %d, got %d", walEntry.KeyLength, readEntry.entry.KeyLength)
		}
		if readEntry.size != 5 {
			t.Errorf("Expected size to be 5, got %d", readEntry.size)
		}

		offset += readEntry.size
	}

	// clean up wal file
	os.Remove("wal.bin")
}
