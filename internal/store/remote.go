package store

import (
	"fmt"
	"log"

	"github.com/ethan-stone/go-key-store/internal/configuration"
	"github.com/ethan-stone/go-key-store/internal/rpc"
)

type RemoteKeyValueStore struct {
	rpcClient *rpc.RpcClient
}

func (store *RemoteKeyValueStore) Get(key string) (string, error) {
	r, err := store.rpcClient.Get(key)

	if err != nil {
		return "", err
	}

	if !r.GetOk() {
		return "", fmt.Errorf("could not find key \"%s\"", key)
	}

	return r.GetVal(), nil
}

func (store *RemoteKeyValueStore) Put(key string, val string) error {
	r, err := store.rpcClient.Put(key, val)

	if err != nil {
		return err
	}

	if !r.GetOk() {
		return fmt.Errorf("could put key \"%s\"", key)
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

func InitializeRemoteStores(clusterConfig *configuration.ClusterConfig) {
	for i := range clusterConfig.OtherNodes {
		address := clusterConfig.OtherNodes[i].Address

		if address == clusterConfig.ThisNode.Address {
			continue
		}

		client, err := rpc.GetOrCreateRpcClient(&rpc.RpcClientConfig{
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
