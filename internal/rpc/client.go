package rpc

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RpcClient interface {
	Ping() (bool, error)
	Get(key string) (*GetResponse, error)
	Put(key string, val string) (*PutResponse, error)
	Delete(key string) (*DeleteResponse, error)
	Gossip(req *GossipRequest) (*GossipResponse, error)
	GetAddress() string
	SetClusterConfig(req *SetClusterConfigRequest) (*SetClusterConfigResponse, error)
	GetClusterConfig(req *GetClusterConfigRequest) (*GetClusterConfigResponse, error)
}

type GrpcClient struct {
	conn    *grpc.ClientConn
	client  StoreServiceClient
	Address string
}

// RpcClientCreator is a function type for creating RpcClient instances.
// It accepts variadic grpc.DialOption to allow flexible client configuration.
type RpcClientCreator func(address string, opts ...grpc.DialOption) (RpcClient, error)

// NewRpcClient creates a new GrpcClient.
// It accepts variadic grpc.DialOption for custom gRPC client configuration.
func NewRpcClient(address string, opts ...grpc.DialOption) (RpcClient, error) {
	conn, err := grpc.NewClient(address, opts...)

	if err != nil {
		return nil, err
	}

	client := NewStoreServiceClient(conn)

	grpcClient := &GrpcClient{
		conn:    conn,
		client:  client,
		Address: address,
	}

	// the tcp connection is actually only started when the first rpc call is made
	// so we send a ping here to make sure it actually can connect as early as possible
	pingResponse, err := grpcClient.Ping()

	if err != nil || !pingResponse {
		conn.Close() // Close connection if initial ping fails
		return nil, err
	}

	return grpcClient, nil
}

func (rpcClient *GrpcClient) GetAddress() string {
	return rpcClient.Address
}

func (rpcClient *GrpcClient) Ping() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	r, err := rpcClient.client.Ping(ctx, &PingRequest{})

	if err != nil {
		return false, err
	}

	log.Printf("Ping result ok = %t", r.GetOk())

	return r.GetOk(), nil
}

func (rpcClient *GrpcClient) Get(key string) (*GetResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	r, err := rpcClient.client.Get(ctx, &GetRequest{Key: key})

	if err != nil {
		return nil, err
	}

	log.Printf("Get result ok = %t", r.GetOk())

	return r, nil
}

func (rpcClient *GrpcClient) Put(key string, val string) (*PutResponse, error) {
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

func (rpcClient *GrpcClient) Delete(key string) (*DeleteResponse, error) {
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

func (rpcClient *GrpcClient) Gossip(req *GossipRequest) (*GossipResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	r, err := rpcClient.client.Gossip(ctx, req)

	if err != nil {
		return nil, err
	}

	log.Printf("Gossip result ok = %t", r.GetOk())

	return r, nil
}

func (rpcClient *GrpcClient) SetClusterConfig(req *SetClusterConfigRequest) (*SetClusterConfigResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	r, err := rpcClient.client.SetClusterConfig(ctx, req)

	if err != nil {
		return nil, err
	}

	log.Printf("SetClusterConfig result ok = %t", r.GetOk())

	return r, nil
}

func (rpcClient *GrpcClient) GetClusterConfig(req *GetClusterConfigRequest) (*GetClusterConfigResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	r, err := rpcClient.client.GetClusterConfig(ctx, req)

	if err != nil {
		return nil, err
	}

	log.Printf("GetClusterConfig result ok = %t", r.GetOk())

	return r, nil
}

type RpcClientConfig struct {
	Address string
}

type RpcClientManager interface {
	GetOrCreateRpcClient(config *RpcClientConfig) (RpcClient, error)
}

type GrpcClientManager struct {
	creator    RpcClientCreator // Dependency injected creator
	rpcClients map[string]*GrpcClient
}

func NewGrpcClientManager(creator RpcClientCreator) *GrpcClientManager {
	return &GrpcClientManager{
		creator:    creator,
		rpcClients: make(map[string]*GrpcClient),
	}
}

func (rpcClientManager *GrpcClientManager) GetOrCreateRpcClient(config *RpcClientConfig) (RpcClient, error) {
	existingClient, ok := rpcClientManager.rpcClients[config.Address]

	if ok && existingClient != nil {
		return existingClient, nil
	}

	// Use the injected creator here
	newClient, err := rpcClientManager.creator(config.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}

	rpcClientManager.rpcClients[config.Address] = newClient.(*GrpcClient)

	go func() {
		for range time.NewTicker(time.Second * 5).C {

			r, err := newClient.Ping()

			if err != nil || !r {
				// IMPORTANT: Change log.Fatalf to log.Printf in production code
				// to avoid crashing the entire application on a background ping failure.
				log.Fatalf("Could not ping server %v", err)
			}
		}
	}()

	return newClient, nil
}
