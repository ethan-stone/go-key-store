package factory

import (
	"fmt"
	"log"

	"github.com/ethan-stone/go-key-store/internal/configuration"
	"github.com/ethan-stone/go-key-store/internal/hash"
	"github.com/ethan-stone/go-key-store/internal/local_store"
	"github.com/ethan-stone/go-key-store/internal/remote_store"
	"github.com/ethan-stone/go-key-store/internal/rpc"
	"github.com/ethan-stone/go-key-store/internal/service"
)

func GetStore(key string, configurationManager configuration.ConfigurationManager, rpcClientManager rpc.RpcClientManager) (service.StoreService, error) {
	hashSlot := hash.GetHashSlot(key)

	log.Printf("Key %s belongs to hash slot %d", key, hashSlot)

	_, currentShardGroup, clusterConfig, err := configurationManager.GetCurrentNodeConfig()

	if err != nil {
		return nil, err
	}

	// If the hash falls into this node, then get the local store.
	if hashSlot >= uint32(currentShardGroup.HashSlots[0]) && hashSlot <= uint32(currentShardGroup.HashSlots[1]) {
		log.Printf("Using local store")
		return local_store.Store, nil
	}

	var remoteKeyValueStore *remote_store.RemoteKeyValueStore

	// Find the node that key val belongs to.
	for i := range clusterConfig.ShardGroups {
		shardGroup := clusterConfig.ShardGroups[i]
		if hashSlot >= uint32(shardGroup.HashSlots[0]) && hashSlot <= uint32(shardGroup.HashSlots[1]) {
			for j := range shardGroup.Nodes {
				node := shardGroup.Nodes[j]

				if node.Role == "primary" {
					client, err := rpcClientManager.GetOrCreateRpcClient(&rpc.RpcClientConfig{
						Address: node.GrpcAddress,
					})
					if err != nil {
						return nil, err
					}

					remoteKeyValueStore = &remote_store.RemoteKeyValueStore{
						RpcClient: client,
					}
					break
				}
			}
		}
	}

	if remoteKeyValueStore == nil {
		return nil, fmt.Errorf("could not find remote key value store for hash slot %d", hashSlot)
	}

	log.Printf("Using remote store")

	return remoteKeyValueStore, nil
}
