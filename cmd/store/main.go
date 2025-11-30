package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ethan-stone/go-key-store/internal/configuration"
	"github.com/ethan-stone/go-key-store/internal/gossip"
	"github.com/ethan-stone/go-key-store/internal/http_server"
	"github.com/ethan-stone/go-key-store/internal/rpc"
	"github.com/ethan-stone/go-key-store/internal/store"
	"github.com/ethan-stone/go-key-store/internal/wal"
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

	grpcClientManager := rpc.NewGrpcClientManager(rpc.NewRpcClient)

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

	noopWalWriter := wal.NewNoopWalWriter()

	localStore := store.InitializeLocalKeyValueStore(&store.InitializeLocalKeyValueStoreConfig{
		WalWriter: noopWalWriter,
	})

	err := localStore.LoadFromSnapshot("snapshot.bin")

	if err != nil {
		log.Fatalf("failed to load from snapshot %v", err)
	}

	walFiles, _ := filepath.Glob("wals/wal_*.bin")
	sort.Strings(walFiles)

	for _, walPath := range walFiles {
		fmt.Println("replaying", walPath)
		walReader := wal.NewFileWalReader(walPath)
		offset := int64(0)

		for {
			entry, err := walReader.Read(offset)

			if err != nil {
				if err == io.EOF {
					break
				}

				panic(err)
			}

			localStore.ApplyWalEntry(entry.Entry)
			fmt.Println("Applied wal entry ->", entry.Entry.SequenceNumber)

			offset += entry.Size

		}

		walReader.Close()

	}

	nextIndex := 0

	if len(walFiles) > 0 {
		latestWalFile := walFiles[len(walFiles)-1]
		latestWalFileIndex := strings.Split(strings.Split(latestWalFile, "wal_")[1], ".bin")[0]
		nextIndex, err = strconv.Atoi(latestWalFileIndex)

		if err != nil {
			log.Fatalf("failed to convert latest wal file index to int %v", err)
		}
	}

	nextIndex++

	fmt.Println("Store loaded with last applied sequence number ->", localStore.GetLastAppliedSequenceNumber())

	// For simplicity, we start a new WAL file.
	// For max efficiency, we could try to reuse the most recent WAL file if valid/healthy, but not right now.
	localStore.SetWalWriter(wal.NewFileWalWriter(
		&wal.WalWriterConfig{
			Directory:      "wals",
			SyncMode:       wal.SyncModeAlways,
			Index:          nextIndex,
			SequenceNumber: localStore.GetLastAppliedSequenceNumber(),
		},
	))

	localStore.SubscribeToWalEntries()
	localStore.StartSnapshotManager(time.Second*1, 64) // 64MB

	httpServer := http_server.NewHttpServer(
		&http_server.HttpServerConfig{
			Address:          ":" + httpPort,
			ConfigManager:    configurationManager,
			RpcClientManager: grpcClientManager,
		},
	)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("HTTP server running on port %s", httpPort)

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start http server %v", err)
		}
	}()

	list, err := net.Listen("tcp", ":"+grpcPort)

	if err != nil {
		log.Fatalf("failed to start grpc server %v", err)
	}

	log.Printf("GRPC server runnnig on port %s", grpcPort)

	grpcServer := rpc.NewRpcServer(localStore, configurationManager, grpcClientManager)

	go func() {
		if err := grpcServer.Serve(list); err != nil {
			log.Fatalf("failed to start grpc server %v", err)
		}
	}()

	<-stop

	// Create a context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := httpServer.Shutdown(ctx); err != nil {
		fmt.Printf("Forced shutdown: %v\n", err)
	}

	grpcServer.GracefulStop()

	localStore.Close()

	fmt.Println("Server stopped cleanly.")
}
