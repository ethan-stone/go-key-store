package wal

import (
	"os"
	"testing"
)

func TestWrite(t *testing.T) {
	wal := NewWal("wal.bin")

	walEntry := &WalEntry{
		OpType: Put,
	}

	err := wal.Write(walEntry)

	if err != nil {
		t.Fatalf("Did not expect an error when writing")
	}

	// clean up wal file
	os.Remove("wal.bin")
}

func TestRead(t *testing.T) {
	wal := NewWal("wal.bin")

	walEntry := &WalEntry{
		OpType: Put,
	}

	err := wal.Write(walEntry)

	if err != nil {
		t.Fatalf("Did not expect an error when writing")
	}

	walEntry2 := &WalEntry{
		OpType: Del,
	}

	err = wal.Write(walEntry2)

	if err != nil {
		t.Fatalf("Did not expect an error when writing")
	}

	readEntry, err := wal.Read(0)

	if err != nil {
		t.Fatalf("Did not expect an error when reading")
	}

	if readEntry.entry.OpType != walEntry.OpType {
		t.Errorf("Expected op type to be %d, got %d", walEntry.OpType, readEntry.entry.OpType)
	}

	if readEntry.size != 1 {
		t.Errorf("Expected size to be 1, got %d", readEntry.size)
	}

	readEntry2, err := wal.Read(1)

	if err != nil {
		t.Fatalf("Did not expect an error when reading")
	}

	if readEntry2.entry.OpType != walEntry2.OpType {
		t.Errorf("Expected op type to be %d, got %d", walEntry2.OpType, readEntry2.entry.OpType)
	}

	if readEntry2.size != 1 {
		t.Errorf("Expected size to be 1, got %d", readEntry2.size)
	}

	// clean up wal file
	os.Remove("wal.bin")
}
