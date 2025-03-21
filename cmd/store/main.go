package main

import (
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

	for i := range clusterConfig.OtherNodes {
		address := clusterConfig.OtherNodes[i].Address

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

	httpServer := http_server.NewHttpServer(":" + args[1])

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

	grpcServer := rpc.NewRpcServer()

	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf("failed to start grpc server %v", err)
	}

}
