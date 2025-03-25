package store

import (
	"testing"

	"github.com/ethan-stone/go-key-store/internal/configuration"
	"github.com/ethan-stone/go-key-store/internal/rpc"
)

type MockRpcClientManager struct {
	MockGetOrCreateRpcClient func(config *rpc.RpcClientConfig) (rpc.RpcClient, error)
}

func (m *MockRpcClientManager) GetOrCreateRpcClient(config *rpc.RpcClientConfig) (rpc.RpcClient, error) {
	return m.MockGetOrCreateRpcClient(config)
}

type MockRpcClient struct {
}

func (m *MockRpcClient) Ping() (bool, error) {
	return true, nil
}
func (m *MockRpcClient) Get(key string) (*rpc.GetResponse, error) {
	return &rpc.GetResponse{
		Key: "a",
		Val: "b",
		Ok:  true,
	}, nil
}
func (m *MockRpcClient) Put(key string, val string) (*rpc.PutResponse, error) {
	return &rpc.PutResponse{
		Ok: true,
	}, nil
}
func (m *MockRpcClient) Delete(key string) (*rpc.DeleteResponse, error) {
	return &rpc.DeleteResponse{
		Ok: true,
	}, nil
}
func (m *MockRpcClient) Gossip(req *rpc.GossipRequest) (*rpc.GossipResponse, error) {
	return &rpc.GossipResponse{
		Ok:         true,
		OtherNodes: nil,
	}, nil
}
func (m *MockRpcClient) GetAddress() string {
	return "localhost:8081"
}

// a hashes to slot 15939
// b hashes to slot 12281
// c hashes to slot 8047
// d hashes to slot 2764

var node1 = &configuration.NodeConfig{
	Address:   "localhost:8081",
	HashSlots: []int{0, 4095},
}

var node2 = &configuration.NodeConfig{
	Address:   "localhost:8083",
	HashSlots: []int{4096, 8191},
}

var node3 = &configuration.NodeConfig{
	Address:   "localhost:8085",
	HashSlots: []int{8192, 12287},
}

var node4 = &configuration.NodeConfig{
	Address:   "localhost:8087",
	HashSlots: []int{12288, 16383},
}

func TestReturnsLocalStore(t *testing.T) {
	key := "a"

	// current node is node4, and a DOES fall into node1's hash slots
	clusterConfig := &configuration.ClusterConfig{
		ThisNode: node4,
		OtherNodes: []*configuration.NodeConfig{
			node1, node2, node3,
		},
	}

	mockRpcClientManager := &MockRpcClientManager{
		MockGetOrCreateRpcClient: func(config *rpc.RpcClientConfig) (rpc.RpcClient, error) {
			return &MockRpcClient{}, nil
		},
	}

	InitializeLocalKeyValueStore(clusterConfig)

	store, err := GetStore(key, clusterConfig, mockRpcClientManager)

	if err != nil {
		t.Fatalf("Did not expect an error when getting store %v", err)
	}

	_, ok := store.(*LocalKeyValueStore)

	if !ok {
		t.Errorf("Expected *LocalKeyValueStore, got %T", store)
	}
}

func TestReturnsRemoteStore(t *testing.T) {
	key := "a"

	// current node is node1, and a DOES NOT fall into node1's hash slots.
	clusterConfig := &configuration.ClusterConfig{
		ThisNode: node1,
		OtherNodes: []*configuration.NodeConfig{
			node2, node3, node4,
		},
	}

	mockRpcClientManager := &MockRpcClientManager{
		MockGetOrCreateRpcClient: func(config *rpc.RpcClientConfig) (rpc.RpcClient, error) {
			return &MockRpcClient{}, nil
		},
	}

	InitializeLocalKeyValueStore(clusterConfig)

	store, err := GetStore(key, clusterConfig, mockRpcClientManager)

	if err != nil {
		t.Fatalf("Did not expect an error when getting store %v", err)
	}

	_, ok := store.(*RemoteKeyValueStore)

	if !ok {
		t.Errorf("Expected *RemoteKeyValueStore, got %T", store)
	}
}
