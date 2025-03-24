package store

import (
	"fmt"
	"log"
	"time"

	"github.com/ethan-stone/go-key-store/internal/configuration"
	"github.com/ethan-stone/go-key-store/internal/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

var RemoteKeyValueStores map[string]*RemoteKeyValueStore = make(map[string]*RemoteKeyValueStore)

func InitializeRemoteStores(clusterConfig *configuration.ClusterConfig) {
	for i := range clusterConfig.OtherNodes {
		address := clusterConfig.OtherNodes[i].Address

		client, err := rpc.NewRpcClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))

		if err != nil {
			log.Fatalf("Failed to make grpc client %v", err)
		}

		remoteKeyValueStore := &RemoteKeyValueStore{
			rpcClient: client,
		}

		// TODO use node ID
		RemoteKeyValueStores[address] = remoteKeyValueStore

		go func() {
			for range time.NewTicker(time.Second * 5).C {

				r, err := client.Ping()

				if err != nil || !r {
					log.Fatalf("Could not ping server %v", err)
				}
			}
		}()
	}
}
