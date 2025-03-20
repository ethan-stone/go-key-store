package http_server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ethan-stone/go-key-store/internal/store"
)

type KeyValueResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type PutRequestBody struct {
	Value string `json:"value"`
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	val, ok := store.Store.Get(key)

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

	store.Store.Put(key, body.Value)

	w.WriteHeader(http.StatusOK)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	store.Store.Del(key)

	w.WriteHeader(http.StatusOK)
}

func NewHttpServer(address string) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /item/{key}", getHandler)
	mux.HandleFunc("POST /item/{key}", putHandler)
	mux.HandleFunc("DELETE /item/{key}", deleteHandler)

	// this is the actual server
	httpServer := &http.Server{
		Handler:      mux, // This is the important line
		Addr:         address,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return httpServer
}
