package remote_store

import (
	"fmt"
	"log"

	"github.com/ethan-stone/go-key-store/internal/configuration"
	"github.com/ethan-stone/go-key-store/internal/rpc"
	"github.com/ethan-stone/go-key-store/internal/service"
)

type RemoteKeyValueStore struct {
	RpcClient rpc.RpcClient
}

func (store *RemoteKeyValueStore) Get(key string) (*service.GetResult, error) {
	r, err := store.RpcClient.Get(key)

	if err != nil {
		return nil, err
	}

	if !r.GetOk() {
		return &service.GetResult{
			Ok:  false,
			Val: "",
		}, nil
	}

	return &service.GetResult{
		Ok:  true,
		Val: r.GetVal(),
	}, nil
}

func (store *RemoteKeyValueStore) Put(key string, val string) error {
	r, err := store.RpcClient.Put(key, val)

	if err != nil {
		return err
	}

	if !r.GetOk() {
		return fmt.Errorf("could not put key \"%s\"", key)
	}

	return nil
}

func (store *RemoteKeyValueStore) Delete(key string) error {
	r, err := store.RpcClient.Delete(key)

	if err != nil {
		return err
	}

	if !r.GetOk() {
		return fmt.Errorf("could delete key \"%s\"", key)
	}

	return nil
}

var remoteKeyValueStores map[string]*RemoteKeyValueStore = make(map[string]*RemoteKeyValueStore)

func InitializeRemoteStores(configManager configuration.ConfigurationManager, rpcClientManager rpc.RpcClientManager) {

	currentNode, _, clusterConfig, err := configManager.GetCurrentNodeConfig()

	if err != nil {
		log.Fatalf("Failed to get current node config %v", err)
	}

	for i := range clusterConfig.ShardGroups {
		shardGroup := clusterConfig.ShardGroups[i]
		for j := range shardGroup.Nodes {
			node := shardGroup.Nodes[j]
			if node.Role == "primary" && node.ID != currentNode.ID {
				client, err := rpcClientManager.GetOrCreateRpcClient(&rpc.RpcClientConfig{
					Address: node.GrpcAddress,
				})
				if err != nil {
					log.Fatalf("Failed to make grpc client %v", err)
				}

				remoteKeyValueStore := &RemoteKeyValueStore{
					RpcClient: client,
				}
				remoteKeyValueStores[node.GrpcAddress] = remoteKeyValueStore
			}
		}
	}

}
