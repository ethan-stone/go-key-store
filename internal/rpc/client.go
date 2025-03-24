package rpc

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RpcClient struct {
	conn    *grpc.ClientConn
	client  StoreServiceClient
	Address string
}

func NewRpcClient(address string, opts grpc.DialOption) (*RpcClient, error) {
	conn, err := grpc.NewClient(address, opts)

	if err != nil {
		return nil, err
	}

	client := NewStoreServiceClient(conn)

	return &RpcClient{
		conn:    conn,
		client:  client,
		Address: address,
	}, nil
}

func (rpcClient *RpcClient) Ping() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	r, err := rpcClient.client.Ping(ctx, &PingRequest{})

	if err != nil {
		return false, err
	}

	log.Printf("Ping result ok = %t", r.GetOk())

	return r.GetOk(), nil
}

func (rpcClient *RpcClient) Get(key string) (*GetResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	r, err := rpcClient.client.Get(ctx, &GetRequest{Key: key})

	if err != nil {
		return nil, err
	}

	log.Printf("Get result ok = %t", r.GetOk())

	return r, nil
}

func (rpcClient *RpcClient) Put(key string, val string) (*PutResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	r, err := rpcClient.client.Put(ctx, &PutRequest{
		Key: key,
		Val: val,
	})

	if err != nil {
		return nil, err
	}

	log.Printf("Put result ok = %t", r.GetOk())

	return r, nil
}

func (rpcClient *RpcClient) Delete(key string) (*DeleteResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	r, err := rpcClient.client.Delete(ctx, &DeleteRequest{
		Key: key,
	})

	if err != nil {
		return nil, err
	}

	log.Printf("Delete result ok = %t", r.GetOk())

	return r, nil
}

func (rpcClient *RpcClient) Gossip(req *GossipRequest) (*GossipResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	r, err := rpcClient.client.Gossip(ctx, req)

	if err != nil {
		return nil, err
	}

	log.Printf("Gossip result ok = %t", r.GetOk())

	jsonData, err := json.Marshal(r)

	if err != nil {
		return nil, err
	}

	log.Printf("Response = %s", jsonData)

	return r, nil
}

// map of node ID to rpc client
var rpcClients map[string]*RpcClient = make(map[string]*RpcClient)

type RpcClientConfig struct {
	Address string
}

func GetOrCreateRpcClient(config *RpcClientConfig) (*RpcClient, error) {
	existingClient, ok := rpcClients[config.Address]

	if ok && existingClient != nil {
		return existingClient, nil
	}

	newClient, err := NewRpcClient(config.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}

	rpcClients[config.Address] = newClient

	go func() {
		for range time.NewTicker(time.Second * 5).C {

			r, err := newClient.Ping()

			if err != nil || !r {
				log.Fatalf("Could not ping server %v", err)
			}
		}
	}()

	return newClient, nil
}
