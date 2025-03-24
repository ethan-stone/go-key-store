package store

import (
	"fmt"
	"log"

	"github.com/ethan-stone/go-key-store/internal/configuration"
	"github.com/ethan-stone/go-key-store/internal/hash"
	"github.com/ethan-stone/go-key-store/internal/rpc"
	"github.com/ethan-stone/go-key-store/internal/service"
)

func GetStore(key string, clusterConfig *configuration.ClusterConfig) (service.StoreService, error) {
	hashSlot := hash.GetHashSlot(key)

	log.Printf("Key %s belongs to hash slot %d", key, hashSlot)

	// If the hash falls into this node, then get the local store.
	if hashSlot >= uint32(clusterConfig.ThisNode.HashSlots[0]) && hashSlot <= uint32(clusterConfig.ThisNode.HashSlots[1]) {
		log.Printf("Using local store")
		return Store, nil
	}

	var remoteKeyValueStore *RemoteKeyValueStore

	// Find the node that key val belongs to.
	for i := range clusterConfig.OtherNodes {
		otherNode := clusterConfig.OtherNodes[i]
		if hashSlot >= uint32(otherNode.HashSlots[0]) && hashSlot <= uint32(otherNode.HashSlots[1]) {
			client, err := rpc.GetOrCreateRpcClient(&rpc.RpcClientConfig{
				Address: otherNode.Address,
			})

			if err != nil {
				return nil, err
			}

			remoteKeyValueStore = &RemoteKeyValueStore{
				rpcClient: client,
			}
		}
	}

	if remoteKeyValueStore == nil {
		return nil, fmt.Errorf("could not find remote key value store for hash slot %d", hashSlot)
	}

	log.Printf("Using remote store")

	return remoteKeyValueStore, nil
}
