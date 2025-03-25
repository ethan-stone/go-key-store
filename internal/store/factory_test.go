package store

import (
	"testing"

	"github.com/ethan-stone/go-key-store/internal/rpc"
)

type MockRpcClientManager struct {
	MockGetOrCreateRpcClient func(config *rpc.RpcClientConfig) (rpc.RpcClient, error)
}

func (m *MockRpcClientManager) GetOrCreateRpcClient(config *rpc.RpcClientConfig) (rpc.RpcClient, error) {
	return m.MockGetOrCreateRpcClient(config)
}

func TestReturnsRemoteStore(t *testing.T) {
}
