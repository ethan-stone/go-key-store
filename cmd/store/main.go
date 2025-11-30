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
	"github.com/ethan-stone/go-key-store/internal/http_server"
	"github.com/ethan-stone/go-key-store/internal/local_store"
	"github.com/ethan-stone/go-key-store/internal/rpc"
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

	var (
		nodeID     string
		configFile string
	)

	flag.StringVar(&nodeID, "node-id", "", "The ID of the node")
	flag.StringVar(&configFile, "config-file", "cluster.json", "The file to load the cluster config from")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "Options:")
		flag.PrintDefaults() // Print default values for each flag
	}

	flag.Parse()

	log.SetPrefix(nodeID + " ")

	configurationManager := configuration.NewBaseConfigurationManager()
	err := configurationManager.SetCurrentNodeID(nodeID)

	if err != nil {
		log.Fatalf("failed to set current node id %v", err)
	}

	err = configurationManager.LoadClusterConfig(configFile)

	if err != nil {
		log.Fatalf("failed to load cluster config %v", err)
	}

	currentNodeConfig, currentShardGroup, clusterConfig, err := configurationManager.GetCurrentNodeConfig()

	if err != nil {
		log.Fatalf("failed to get current node config: %v", err)
	}

	grpcClientManager := rpc.NewGrpcClientManager(rpc.NewRpcClient)

	// initialize rpc clients
	for i := range clusterConfig.ShardGroups {
		shardGroup := clusterConfig.ShardGroups[i]
		for j := range shardGroup.Nodes {
			node := shardGroup.Nodes[j]
			if node.ID != currentNodeConfig.ID {
				grpcClientManager.GetOrCreateRpcClient(&rpc.RpcClientConfig{
					Address: node.GrpcAddress,
				})
			}
		}
	}

	noopWalWriter := wal.NewNoopWalWriter()

	datadir := currentNodeConfig.DataDir

	localStore := local_store.InitializeLocalKeyValueStore(&local_store.InitializeLocalKeyValueStoreConfig{
		WalWriter: noopWalWriter,
		DataDir:   datadir,
	})

	err = localStore.LoadFromSnapshot(filepath.Join(datadir, "snapshot.bin"))

	if err != nil {
		log.Fatalf("failed to load from snapshot %v", err)
	}

	walFiles, _ := filepath.Glob(filepath.Join(datadir, "wals", "wal_*.bin"))
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

	walWriter := wal.NewFileWalWriter(
		&wal.WalWriterConfig{
			Directory:      filepath.Join(datadir, "wals"),
			SyncMode:       wal.SyncModeAlways,
			Index:          nextIndex - 1,
			SequenceNumber: localStore.GetLastAppliedSequenceNumber(),
		},
	)

	localStore.SetWalWriter(walWriter)

	localStore.SubscribeToWalEntries()

	if currentNodeConfig.Role == "primary" {
		localStore.StartSnapshotManager(time.Second*1, 64) // 64MB
	}

	if currentNodeConfig.Role == "primary" {
		for i := range currentShardGroup.Nodes {
			go func() {
				for entry := range walWriter.Subscribe() {
					if currentShardGroup.Nodes[i].ID == currentNodeConfig.ID {
						continue
					}

					rpcClient, err := grpcClientManager.GetOrCreateRpcClient(&rpc.RpcClientConfig{
						Address: currentShardGroup.Nodes[i].GrpcAddress,
					})

					if err != nil {
						log.Fatalf("failed to create rpc client %v", err)
					}

					var keyBytes []byte
					var valueBytes []byte

					if entry.KeyBytes == nil {
						keyBytes = make([]byte, 0)
					} else {
						keyBytes = make([]byte, len(*entry.KeyBytes))
						copy(keyBytes, *entry.KeyBytes)
					}

					if entry.ValueBytes == nil {
						valueBytes = make([]byte, 0)
					} else {
						valueBytes = make([]byte, len(*entry.ValueBytes))
						copy(valueBytes, *entry.ValueBytes)
					}

					rpcClient.AppendWalEntry(&rpc.AppendWalEntryRequest{
						WalEntry: &rpc.WalEntry{
							SequenceNumber: entry.SequenceNumber,
							OpType:         uint32(entry.OpType),
							KeyLength:      int32(entry.KeyLength),
							ValueLength:    int32(entry.ValueLength),
							KeyBytes:       keyBytes,
							ValueBytes:     valueBytes,
						},
					})
				}
			}()
		}
	}

	httpServer := http_server.NewHttpServer(
		&http_server.HttpServerConfig{
			Address:          ":" + strings.Split(currentNodeConfig.HttpAddress, ":")[1],
			ConfigManager:    configurationManager,
			RpcClientManager: grpcClientManager,
		},
	)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("HTTP server running on port %s", currentNodeConfig.HttpAddress)

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start http server %v", err)
		}
	}()

	list, err := net.Listen("tcp", ":"+strings.Split(currentNodeConfig.GrpcAddress, ":")[1])

	if err != nil {
		log.Fatalf("failed to start grpc server %v", err)
	}

	log.Printf("GRPC server runnnig on port %s", currentNodeConfig.GrpcAddress)

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
