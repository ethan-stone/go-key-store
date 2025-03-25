package store

import (
	"testing"
)

func TestPut(t *testing.T) {
	store := &LocalKeyValueStore{
		data: make(map[string]string),
	}

	err := store.Put("a", "b")

	if err != nil {
		t.Fatalf("Did not expect an error when putting into store %v", err)
	}
}

func TestGetShouldReturnNotOkWhenKeyNotFound(t *testing.T) {
	store := &LocalKeyValueStore{
		data: make(map[string]string),
	}

	r, err := store.Get("a")

	if err != nil {
		t.Fatalf("Did not expect an error when getting store %v", err)
	}

	if r.Ok {
		t.Errorf("Did not expect to find key %v", "a")
	}
}

func TestShouldReturnOkWhenKeyFound(t *testing.T) {
	store := &LocalKeyValueStore{
		data: make(map[string]string),
	}

	err := store.Put("a", "b")

	if err != nil {
		t.Fatalf("Did not expect an error when putting into store %v", err)
	}

	r, err := store.Get("a")

	if err != nil {
		t.Fatalf("Did not expect an error when getting from store %v", err)
	}

	if !r.Ok {
		t.Errorf("Did not expect to not find key %s", "a")
	}
}

func TestShouldDelete(t *testing.T) {
	store := &LocalKeyValueStore{
		data: make(map[string]string),
	}

	err := store.Put("a", "b")

	if err != nil {
		t.Fatalf("Did not expect an error when putting into store %v", err)
	}

	r, err := store.Get("a")

	if err != nil {
		t.Fatalf("Did not expect an error when getting from store %v", err)
	}

	if !r.Ok {
		t.Errorf("Did not expect to not find key %s", "a")
	}

	err = store.Delete("a")

	if err != nil {
		t.Fatalf("Did not expect an error when deleting from store %v", err)
	}

	r, err = store.Get("a")

	if err != nil {
		t.Fatalf("Did not expect an error when getting from store %v", err)
	}

	if r.Ok {
		t.Errorf("Did not expect to find key %s", "a")
	}
}
