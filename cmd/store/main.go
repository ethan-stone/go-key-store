package main

import (
	"log"
	"net"
	"os"

	"github.com/ethan-stone/go-key-store/internal/configuration"
	"github.com/ethan-stone/go-key-store/internal/http_server"
	"github.com/ethan-stone/go-key-store/internal/rpc"
	"github.com/ethan-stone/go-key-store/internal/store"
)

func main() {
	log.Default().SetFlags(log.Ldate | log.Ltime | log.Lmsgprefix)
	log.SetPrefix(configuration.GenerateNodeID() + " ")

	// mux is a router
	args := os.Args

	if len(args) != 3 {
		log.Fatalf("Must provide following command line arguments. go run . <http-port> <grpc-port>")
	}

	clusterConfig, err := configuration.LoadClusterConfigFromFile("cluster-config.json", "localhost:"+args[2])

	if err != nil {
		log.Fatalf("Failed to load cluster config file %v", err)
	}

	localStore := store.InitializeLocalKeyValueStore(clusterConfig)

	store.InitializeRemoteStores(clusterConfig)

	httpServer := http_server.NewHttpServer(":"+args[1], clusterConfig)

	go func() {

		log.Printf("HTTP server running on port %s", args[1])

		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatalf("failed to start http server %v", err)
		}
	}()

	list, err := net.Listen("tcp", ":"+args[2])

	if err != nil {
		log.Fatalf("failed to start grpc server %v", err)
	}

	log.Printf("GRPC server runnnig on port %s", args[2])

	grpcServer := rpc.NewRpcServer(localStore)

	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf("failed to start grpc server %v", err)
	}

}
