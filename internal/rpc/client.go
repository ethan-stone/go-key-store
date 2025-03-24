package rpc

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
)

type RpcClient struct {
	conn   *grpc.ClientConn
	client StoreServiceClient
}

func NewRpcClient(address string, opts grpc.DialOption) (*RpcClient, error) {
	conn, err := grpc.NewClient(address, opts)

	if err != nil {
		return nil, err
	}

	client := NewStoreServiceClient(conn)

	return &RpcClient{
		conn:   conn,
		client: client,
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

	return rpcClient.client.Get(ctx, &GetRequest{Key: key})
}

func (rpcClient *RpcClient) Put(key string, val string) (*PutResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	return rpcClient.client.Put(ctx, &PutRequest{
		Key: key,
		Val: val,
	})
}

func (rpcClient *RpcClient) Delete(key string) (*DeleteResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	return rpcClient.client.Delete(ctx, &DeleteRequest{
		Key: key,
	})
}
