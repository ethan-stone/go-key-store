package main

import (
	"log"
	"net"
	"os"
	"time"

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

	// mux is a router
	args := os.Args

	if len(args) != 2 {
		log.Fatalf("Must provide following command line arguments. go run . <bootstrap-config-file-path>")
	}

	nodeBootstrapConfig, err := configuration.LoadNodeBootstrapConfigFromFile(args[1])

	if err != nil {
		log.Fatalf("could not load bootstrap config file %v", err)
	}

	thisNodeConfig := &configuration.NodeConfig{
		ID:        nodeID,
		Address:   "localhost:" + nodeBootstrapConfig.GrpcPort,
		HashSlots: nodeBootstrapConfig.HashSlots,
	}

	var clusterConfig *configuration.ClusterConfig

	grpcClientManager := rpc.NewGrpcClientManager()

	// if the seed node addresses is 0, then that means this node is a seed node
	if len(nodeBootstrapConfig.SeedNodeAddresses) > 0 {
		// TODO randomize seed node selection
		seedNodeRpcClient, err := grpcClientManager.GetOrCreateRpcClient(&rpc.RpcClientConfig{
			Address: nodeBootstrapConfig.SeedNodeAddresses[0],
		})

		if err != nil {
			log.Fatalf("failed to initialize rpc client for seed node with address %s", nodeBootstrapConfig.SeedNodeAddresses[0])
		}

		gossipClient := gossip.NewGossipClient(&gossip.GossipClientConfig{
			ThisNode:  thisNodeConfig,
			RpcClient: seedNodeRpcClient,
		})

		otherNodes, err := gossipClient.Gossip()

		if err != nil {
			log.Fatal("failed to gossip with seed node")
		}

		clusterConfig = &configuration.ClusterConfig{
			ThisNode:   thisNodeConfig,
			OtherNodes: otherNodes,
		}

		configuration.SetClusterConfig(clusterConfig)

		go func() {
			for range time.NewTicker(time.Second * 5).C {
				otherNodes, err := gossipClient.Gossip()

				if err != nil {
					log.Fatal("failed to gossip with seed node")
				}

				clusterConfig = &configuration.ClusterConfig{
					ThisNode:   thisNodeConfig,
					OtherNodes: otherNodes,
				}

				configuration.SetClusterConfig(clusterConfig)

			}
		}()
	} else {
		clusterConfig = &configuration.ClusterConfig{
			ThisNode:   thisNodeConfig,
			OtherNodes: make([]*configuration.NodeConfig, 0),
		}

		configuration.SetClusterConfig(clusterConfig)
	}

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

	if err != nil {
		log.Fatalf("failed to load node bootstrap config file %v", err)
	}

	localStore := store.InitializeLocalKeyValueStore()

	httpServer := http_server.NewHttpServer(
		&http_server.HttpServerConfig{
			Address:          ":" + nodeBootstrapConfig.HttpPort,
			ClusterConfig:    clusterConfig,
			RpcClientManager: grpcClientManager,
		},
	)

	go func() {
		log.Printf("HTTP server running on port %s", nodeBootstrapConfig.HttpPort)

		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatalf("failed to start http server %v", err)
		}
	}()

	list, err := net.Listen("tcp", ":"+nodeBootstrapConfig.GrpcPort)

	if err != nil {
		log.Fatalf("failed to start grpc server %v", err)
	}

	log.Printf("GRPC server runnnig on port %s", nodeBootstrapConfig.GrpcPort)

	grpcServer := rpc.NewRpcServer(localStore, grpcClientManager)

	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf("failed to start grpc server %v", err)
	}

}
