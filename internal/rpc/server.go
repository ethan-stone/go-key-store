package rpc

import (
	"context"
	"log"

	"github.com/ethan-stone/go-key-store/internal/configuration"
	"github.com/ethan-stone/go-key-store/internal/service"
	"google.golang.org/grpc"
)

type RpcServer struct {
	UnimplementedStoreServiceServer
	storeService     service.StoreService
	rpcClientManager RpcClientManager
	configManager    configuration.ConfigurationManager
}

func (s *RpcServer) Ping(_ context.Context, req *PingRequest) (*PingResponse, error) {
	log.Println("Ping request received.")
	return &PingResponse{Ok: true}, nil
}

func (s *RpcServer) Get(_ context.Context, req *GetRequest) (*GetResponse, error) {
	log.Printf("Get request received for key %s", req.GetKey())

	result, err := s.storeService.Get(req.GetKey())

	if err != nil {
		return nil, err
	}

	if !result.Ok {
		return &GetResponse{
			Key: req.GetKey(),
			Val: "",
			Ok:  false,
		}, nil
	}

	return &GetResponse{
		Key: req.GetKey(),
		Val: result.Val,
		Ok:  true,
	}, nil
}

func (s *RpcServer) Put(_ context.Context, req *PutRequest) (*PutResponse, error) {
	log.Printf("Put request received for key %s", req.GetKey())

	err := s.storeService.Put(req.GetKey(), req.GetVal())

	if err != nil {
		return nil, err
	}

	return &PutResponse{
		Ok: true,
	}, nil
}

func (s *RpcServer) Delete(_ context.Context, req *DeleteRequest) (*DeleteResponse, error) {
	log.Printf("Delete request received for key %s", req.GetKey())

	err := s.storeService.Delete(req.GetKey())

	if err != nil {
		return nil, err
	}

	return &DeleteResponse{
		Ok: true,
	}, nil
}

func (s *RpcServer) Gossip(_ context.Context, req *GossipRequest) (*GossipResponse, error) {
	log.Printf("Received Gossip request from node %s", req.GetNodeId())

	clusterConfig := s.configManager.GetClusterConfig()

	clusterConfig.OtherNodes = append(clusterConfig.OtherNodes, &configuration.NodeConfig{
		ID:        req.GetNodeId(),
		Address:   req.GetAddress(),
		HashSlots: []int{int(req.GetHashSlotsStart()), int(req.GetHashSlotsEnd())},
	})

	otherNodes := []*NodeConfig{}

	for i := range clusterConfig.OtherNodes {
		otherNode := clusterConfig.OtherNodes[i]

		otherNodes = append(otherNodes, &NodeConfig{
			NodeId:         otherNode.ID,
			Address:        otherNode.Address,
			HashSlotsStart: uint32(otherNode.HashSlots[0]),
			HashSlotsEnd:   uint32(otherNode.HashSlots[1]),
		})
	}

	otherNodes = append(otherNodes, &NodeConfig{
		NodeId:         clusterConfig.ThisNode.ID,
		Address:        clusterConfig.ThisNode.Address,
		HashSlotsStart: uint32(clusterConfig.ThisNode.HashSlots[0]),
		HashSlotsEnd:   uint32(clusterConfig.ThisNode.HashSlots[1]),
	})

	_, err := s.rpcClientManager.GetOrCreateRpcClient(&RpcClientConfig{
		Address: req.GetAddress(),
	})

	if err != nil {
		return nil, err
	}

	s.configManager.SetClusterConfig(clusterConfig)

	return &GossipResponse{
		OtherNodes: otherNodes,
		Ok:         true,
	}, nil
}

func (s *RpcServer) SetClusterConfig(_ context.Context, req *SetClusterConfigRequest) (*SetClusterConfigResponse, error) {
	log.Println("Received SetClusterConfig request")

	clusterConfig := s.configManager.GetClusterConfig()

	otherNodes := []*configuration.NodeConfig{}

	for i := range req.OtherNodes {
		otherNode := req.OtherNodes[i]

		otherNodes = append(otherNodes, &configuration.NodeConfig{
			ID:        otherNode.NodeId,
			Address:   otherNode.Address,
			HashSlots: []int{int(otherNode.HashSlotsStart), int(otherNode.HashSlotsEnd)},
		})
	}

	s.configManager.SetClusterConfig(&configuration.ClusterConfig{
		ThisNode:   clusterConfig.ThisNode,
		OtherNodes: otherNodes,
	})

	return &SetClusterConfigResponse{
		Ok: true,
	}, nil
}

func (s *RpcServer) GetNodeConfig(_ context.Context, req *GetNodeConfigRequest) (*GetNodeConfigResponse, error) {
	log.Println("Received GetNodeConfig request")

	clusterConfig := s.configManager.GetClusterConfig()

	return &GetNodeConfigResponse{
		Ok: true,
		NodeConfig: &NodeConfig{
			NodeId:         clusterConfig.ThisNode.ID,
			Address:        clusterConfig.ThisNode.Address,
			HashSlotsStart: uint32(clusterConfig.ThisNode.HashSlots[0]),
			HashSlotsEnd:   uint32(clusterConfig.ThisNode.HashSlots[1]),
		},
	}, nil
}

func (s *RpcServer) SetNodeConfig(_ context.Context, req *SetNodeConfigRequest) (*SetNodeConfigResponse, error) {
	log.Println("Received SetClusterConfig request")

	clusterConfig := s.configManager.GetClusterConfig()

	s.configManager.SetClusterConfig(&configuration.ClusterConfig{
		ThisNode: &configuration.NodeConfig{
			ID:        clusterConfig.ThisNode.ID,
			Address:   clusterConfig.ThisNode.Address,
			HashSlots: []int{int(req.GetNodeConfig().GetHashSlotsStart()), int(req.GetNodeConfig().GetHashSlotsEnd())},
		},
		OtherNodes: clusterConfig.OtherNodes,
	})

	return &SetNodeConfigResponse{
		Ok: true,
	}, nil
}

func NewRpcServer(storeService service.StoreService, rpcClientManager RpcClientManager) *grpc.Server {
	grpcServer := grpc.NewServer()

	RegisterStoreServiceServer(grpcServer, &RpcServer{
		storeService:     storeService,
		rpcClientManager: rpcClientManager,
	})

	return grpcServer
}
