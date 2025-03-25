package http_server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ethan-stone/go-key-store/internal/configuration"
	"github.com/ethan-stone/go-key-store/internal/rpc"
	"github.com/ethan-stone/go-key-store/internal/store"
)

type KeyValueResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type PutRequestBody struct {
	Value string `json:"value"`
}

func getHandler(clusterConfig *configuration.ClusterConfig, rpcClientManager rpc.RpcClientManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")

		store, err := store.GetStore(key, clusterConfig, rpcClientManager)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		result, err := store.Get(key)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !result.Ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(KeyValueResponse{Key: key, Value: result.Val})
	}
}

func putHandler(clusterConfig *configuration.ClusterConfig, rpcClientManager rpc.RpcClientManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")

		store, err := store.GetStore(key, clusterConfig, rpcClientManager)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var body PutRequestBody

		err = json.NewDecoder(r.Body).Decode(&body)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		store.Put(key, body.Value)

		w.WriteHeader(http.StatusOK)

	}
}

func deleteHandler(clusterConfig *configuration.ClusterConfig, rpcClientManager rpc.RpcClientManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")

		store, err := store.GetStore(key, clusterConfig, rpcClientManager)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		store.Delete(key)

		w.WriteHeader(http.StatusOK)
	}

}

type HttpServerConfig struct {
	Address          string
	ClusterConfig    *configuration.ClusterConfig
	RpcClientManager rpc.RpcClientManager
}

func NewHttpServer(config *HttpServerConfig) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /item/{key}", getHandler(config.ClusterConfig, config.RpcClientManager))
	mux.HandleFunc("POST /item/{key}", putHandler(config.ClusterConfig, config.RpcClientManager))
	mux.HandleFunc("DELETE /item/{key}", deleteHandler(config.ClusterConfig, config.RpcClientManager))

	// this is the actual server
	httpServer := &http.Server{
		Handler:      mux, // This is the important line
		Addr:         config.Address,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return httpServer
}
