package main

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/ethan-stone/go-key-store/internal/configuration"
	"github.com/ethan-stone/go-key-store/internal/http_server"
	"github.com/ethan-stone/go-key-store/internal/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type KeyValueResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type PutRequestBody struct {
	Value string `json:"value"`
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

	httpServer := http_server.NewHttpServer(":" + args[1])

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

		client, err := rpc.NewRpcClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))

		if err != nil {
			log.Fatalf("Failed to make grpc client %v", err)
		}

		go func() {
			for range time.NewTicker(time.Second * 5).C {

				r, err := client.Ping()

				if err != nil || !r {
					log.Fatalf("Could not ping server %v", err)
				}
			}
		}()
	}

	list, err := net.Listen("tcp", ":"+args[2])

	if err != nil {
		log.Fatalf("failed to start grpc server %v", err)
	}

	log.Printf("GRPC server runnnig on port %s", args[2])

	grpcServer := rpc.NewRpcServer()

	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf("failed to start grpc server %v", err)
	}

}
