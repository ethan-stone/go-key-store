package store

import (
	"fmt"
	"log"

	"github.com/ethan-stone/go-key-store/internal/configuration"
	"github.com/ethan-stone/go-key-store/internal/rpc"
	"github.com/ethan-stone/go-key-store/internal/service"
)

type RemoteKeyValueStore struct {
	rpcClient rpc.RpcClient
}

func (store *RemoteKeyValueStore) Get(key string) (*service.GetResult, error) {
	r, err := store.rpcClient.Get(key)

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
	r, err := store.rpcClient.Put(key, val)

	if err != nil {
		return err
	}

	if !r.GetOk() {
		return fmt.Errorf("could not put key \"%s\"", key)
	}

	return nil
}

func (store *RemoteKeyValueStore) Delete(key string) error {
	r, err := store.rpcClient.Delete(key)

	if err != nil {
		return err
	}

	if !r.GetOk() {
		return fmt.Errorf("could delete key \"%s\"", key)
	}

	return nil
}

var remoteKeyValueStores map[string]*RemoteKeyValueStore = make(map[string]*RemoteKeyValueStore)

func InitializeRemoteStores(clusterConfig *configuration.ClusterConfig, rpcClientManager rpc.RpcClientManager) {
	for i := range clusterConfig.OtherNodes {
		address := clusterConfig.OtherNodes[i].Address

		if address == clusterConfig.ThisNode.Address {
			continue
		}

		client, err := rpcClientManager.GetOrCreateRpcClient(&rpc.RpcClientConfig{
			Address: address,
		})

		if err != nil {
			log.Fatalf("Failed to make grpc client %v", err)
		}

		remoteKeyValueStore := &RemoteKeyValueStore{
			rpcClient: client,
		}

		remoteKeyValueStores[address] = remoteKeyValueStore

	}
}
