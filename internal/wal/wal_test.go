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

	walEntry := &WalEntry{
		OpType:    Put,
		KeyLength: 2,
	}

	err := wal.Write(walEntry)

	if err != nil {
		t.Fatalf("Did not expect an error when writing")
	}

	walEntry2 := &WalEntry{
		OpType:    Del,
		KeyLength: 5,
	}

	err = wal.Write(walEntry2)

	if err != nil {
		t.Fatalf("Did not expect an error when writing")
	}

	reader := NewWalReader("wal.bin")

	readEntry, err := reader.Read(0)

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
		t.Errorf("Expected size to be 1, got %d", readEntry.size)
	}

	readEntry2, err := reader.Read(readEntry.size)

	if err != nil {
		t.Fatalf("Did not expect an error when reading")
	}

	if readEntry2.entry.OpType != walEntry2.OpType {
		t.Errorf("Expected op type to be %d, got %d", walEntry2.OpType, readEntry2.entry.OpType)
	}

	if readEntry2.entry.KeyLength != walEntry2.KeyLength {
		t.Errorf("Expected key length to be %d, got %d", walEntry2.KeyLength, readEntry2.entry.KeyLength)
	}

	if readEntry2.size != 5 {
		t.Errorf("Expected size to be 1, got %d", readEntry2.size)
	}

	// clean up wal file
	os.Remove("wal.bin")
}
