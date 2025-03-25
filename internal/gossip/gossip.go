package gossip

import (
	"log"

	"github.com/ethan-stone/go-key-store/internal/configuration"
	"github.com/ethan-stone/go-key-store/internal/rpc"
)

type GossipClient struct {
	rpcClient rpc.RpcClient             // the client to the node to gossip with
	thisNode  *configuration.NodeConfig // the config of the current node
}

func (gossipClient *GossipClient) Gossip() ([]*configuration.NodeConfig, error) {
	r, err := gossipClient.rpcClient.Gossip(
		&rpc.GossipRequest{
			NodeId:         gossipClient.thisNode.ID,
			Address:        gossipClient.thisNode.Address,
			HashSlotsStart: uint32(gossipClient.thisNode.HashSlots[0]),
			HashSlotsEnd:   uint32(gossipClient.thisNode.HashSlots[1]),
		},
	)

	if err != nil {
		log.Printf("Failed to gossip with node %s", gossipClient.rpcClient.GetAddress())
		return nil, err
	}

	otherNodes := []*configuration.NodeConfig{}

	for i := range r.OtherNodes {
		otherNode := r.OtherNodes[i]
		otherNodes = append(otherNodes, &configuration.NodeConfig{
			ID:        otherNode.GetNodeId(),
			Address:   otherNode.GetAddress(),
			HashSlots: []int{int(otherNode.GetHashSlotsStart()), int(otherNode.GetHashSlotsEnd())},
		})
	}

	log.Printf("Successfully gossipped with node %s", gossipClient.rpcClient.GetAddress())

	return otherNodes, nil
}

type GossipClientConfig struct {
	RpcClient rpc.RpcClient
	ThisNode  *configuration.NodeConfig
}

func NewGossipClient(config *GossipClientConfig) *GossipClient {
	return &GossipClient{
		rpcClient: config.RpcClient,
		thisNode:  config.ThisNode,
	}
}
