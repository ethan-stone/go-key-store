package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/ethan-stone/go-key-store/internal/configuration"
	"github.com/ethan-stone/go-key-store/internal/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

type server struct {
	rpc.UnimplementedStoreServiceServer
}

func (s *server) Ping(_ context.Context, req *rpc.PingRequest) (*rpc.PingResponse, error) {
	log.Println("Ping request received.")
	return &rpc.PingResponse{Ok: true}, nil
}

type ClusterConfig struct {
	Addresses []string `json:"addresses"`
}

func main() {
	log.Default().SetFlags(log.Ldate | log.Ltime | log.Lmsgprefix)
	log.SetPrefix(configuration.GenerateNodeID() + " ")

	// mux is a router
	args := os.Args

	if len(args) != 3 {
		log.Fatalf("Must provide following command line arguments. go run . <http-port> <grpc-port>")
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /item/{key}", getHandler)
	mux.HandleFunc("POST /item/{key}", putHandler)
	mux.HandleFunc("DELETE /item/{key}", deleteHandler)

	// this is the actual server
	httpServer := &http.Server{
		Handler:      mux, // This is the important line
		Addr:         ":" + args[1],
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {

		log.Printf("HTTP server running on port %s", args[1])

		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatalf("failed to start http server %v", err)
		}
	}()

	configFile, err := os.Open("cluster-config.json")

	if err != nil {
		log.Fatalf("Failed to open cluster config file %v", err)
	}

	byteResult, _ := io.ReadAll(configFile)

	var clusterConfig ClusterConfig
	json.Unmarshal(byteResult, &clusterConfig)

	configFile.Close()

	otherNodeAddresses := []string{}

	for i := range clusterConfig.Addresses {
		if clusterConfig.Addresses[i] != "localhost:"+args[2] {
			log.Printf("Adding address %s to other node addresses", clusterConfig.Addresses[i])
			otherNodeAddresses = append(otherNodeAddresses, clusterConfig.Addresses[i])
		}
	}

	for i := range otherNodeAddresses {
		address := otherNodeAddresses[i]
		conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))

		if err != nil {
			log.Fatalf("Failed to make grpc client %v", err)
		}

		defer conn.Close()

		client := rpc.NewStoreServiceClient(conn)

		go func() {
			for range time.NewTicker(time.Second * 5).C {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)

				r, err := client.Ping(ctx, &rpc.PingRequest{})

				if err != nil {
					log.Fatalf("Could not ping server %v", err)
				}

				log.Printf("Ping successful ok = %t", r.GetOk())

				cancel()
			}
		}()
	}

	list, err := net.Listen("tcp", ":"+args[2])

	if err != nil {
		log.Fatalf("failed to start grpc server %v", err)
	}

	log.Printf("GRPC server runnnig on port %s", args[2])

	grpcServer := grpc.NewServer()

	rpc.RegisterStoreServiceServer(grpcServer, &server{})

	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf("failed to start grpc server %v", err)
	}

}
