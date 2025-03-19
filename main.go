package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

type KeyValueStore struct {
	sync.RWMutex
	data map[string]string
}

func (store *KeyValueStore) get(key string) (string, bool) {
	store.RLock()
	defer store.RUnlock()
	val, ok := store.data[key]
	return val, ok
}

func (store *KeyValueStore) put(key string, val string) {
	store.Lock()
	defer store.Unlock()
	store.data[key] = val
}

func (store *KeyValueStore) del(key string) {
	store.Lock()
	defer store.Unlock()
	delete(store.data, key)
}

type KeyValueResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type PutRequestBody struct {
	Value string `json:"value"`
}

var store = &KeyValueStore{
	data: make(map[string]string),
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	val, ok := store.get(key)

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(KeyValueResponse{Key: key, Value: val})
}

func putHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	var body PutRequestBody

	err := json.NewDecoder(r.Body).Decode(&body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	store.put(key, body.Value)

	w.WriteHeader(http.StatusOK)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	store.del(key)

	w.WriteHeader(http.StatusOK)
}

func main() {
	// mux is a router
	mux := http.NewServeMux()

	mux.HandleFunc("GET /item/{key}", getHandler)
	mux.HandleFunc("POST /item/{key}", putHandler)
	mux.HandleFunc("DELETE /item/{key}", deleteHandler)

	args := os.Args

	if len(args) != 2 {
		panic("Must provide following command line arguments. go run . <port>")
	}

	// this is the actual server
	s := &http.Server{
		Handler:      mux, // This is the important line
		Addr:         ":" + args[1],
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Println("Server running on port 8080")

	if err := s.ListenAndServe(); err != nil {
		panic(err)
	}
}
