package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/ethan-stone/go-key-store/internal/configuration"
	"github.com/ethan-stone/go-key-store/internal/gossip"
	"github.com/ethan-stone/go-key-store/internal/http_server"
	"github.com/ethan-stone/go-key-store/internal/rpc"
	"github.com/ethan-stone/go-key-store/internal/store"
)

// 1. Read config file. This contains info about this node, and seed node to get info of other nodes.
// 2. Initialize RpcClient with seed node.
// 3. Gossip with seed node to get rest of the cluster config.
// 4. Update cluster config.
// 5. Initialize rest of rpc clients.
// 6. Start gRPC server for inter-node communications.
// 7. Start HTTP server for client requests.
func main() {
	log.Default().SetFlags(log.Ldate | log.Ltime | log.Lmsgprefix)

	nodeID := configuration.GenerateNodeID()

	log.SetPrefix(nodeID + " ")

	var (
		httpPort string
		grpcPort string
	)

	flag.StringVar(&httpPort, "http-port", "8080", "")
	flag.StringVar(&grpcPort, "grpc-port", "8081", "")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "Options:")
		flag.PrintDefaults() // Print default values for each flag
	}

	flag.Parse()

	thisNodeConfig := &configuration.NodeConfig{
		ID:        nodeID,
		Address:   "localhost:" + grpcPort,
		HashSlots: []int{0, 16838},
	}

	var clusterConfig *configuration.ClusterConfig

	grpcClientManager := rpc.NewGrpcClientManager()

	otherNodes := []*configuration.NodeConfig{}

	clusterConfig = &configuration.ClusterConfig{
		ThisNode:   thisNodeConfig,
		OtherNodes: otherNodes,
	}

	configurationManager := configuration.NewBaseConfigurationManager(clusterConfig)

	// initialize rpc clients
	for i := range clusterConfig.OtherNodes {
		// skip over current node
		if clusterConfig.OtherNodes[i].Address == clusterConfig.ThisNode.Address {
			continue
		}

		grpcClientManager.GetOrCreateRpcClient(&rpc.RpcClientConfig{
			Address: clusterConfig.OtherNodes[i].Address,
		})
	}

	gossiper := gossip.NewGossipClient(&gossip.GossipClientConfig{
		RpcClientManager: grpcClientManager,
		ConfigManager:    configurationManager,
	})

	gossiper.Gossip()

	localStore := store.InitializeLocalKeyValueStore()

	httpServer := http_server.NewHttpServer(
		&http_server.HttpServerConfig{
			Address:          ":" + httpPort,
			ConfigManager:    configurationManager,
			RpcClientManager: grpcClientManager,
		},
	)

	go func() {
		log.Printf("HTTP server running on port %s", httpPort)

		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatalf("failed to start http server %v", err)
		}
	}()

	list, err := net.Listen("tcp", ":"+grpcPort)

	if err != nil {
		log.Fatalf("failed to start grpc server %v", err)
	}

	log.Printf("GRPC server runnnig on port %s", grpcPort)

	grpcServer := rpc.NewRpcServer(localStore, grpcClientManager)

	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf("failed to start grpc server %v", err)
	}

}
