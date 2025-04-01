package gossip

import (
	"log"
	"math/rand"
	"time"

	"github.com/ethan-stone/go-key-store/internal/configuration"
	"github.com/ethan-stone/go-key-store/internal/rpc"
)

type GossipClient struct {
	rpcClientManager rpc.RpcClientManager
	configManager    configuration.ConfigurationManager
}

func (gossipClient *GossipClient) Gossip() {
	go func() {
		for range time.NewTicker(time.Second * 5).C {
			// pick 3 random nodes from the cluster config and gossip.
			// just in case, only try to generate 3 random indexes 6 times to not get stuck

			clusterConfig := gossipClient.configManager.GetClusterConfig()

			seenIndexes := make(map[int]bool)

			if len(clusterConfig.OtherNodes) == 0 {
				log.Println("No OtherNodes configured. Skipping...")
				continue
			}

			for range 6 {
				idx := rand.Intn(len(clusterConfig.OtherNodes))

				_, ok := seenIndexes[idx]

				// we've seen this index before so try again.
				if ok {
					continue
				}

				seenIndexes[idx] = true

				if len(seenIndexes) == 3 {
					break
				}
			}

			otherNodes := []*configuration.NodeConfig{}

			for j := range clusterConfig.OtherNodes {
				// is this index of the OtherNodes array in the set of the indexes we randomly generated?

				_, ok := seenIndexes[j]

				if !ok {
					continue
				}

				otherNode := clusterConfig.OtherNodes[j]

				client, err := gossipClient.rpcClientManager.GetOrCreateRpcClient(&rpc.RpcClientConfig{Address: otherNode.Address})

				if err != nil {
					continue
				}

				r, err := client.Gossip(&rpc.GossipRequest{
					NodeId:         clusterConfig.ThisNode.ID,
					Address:        clusterConfig.ThisNode.Address,
					HashSlotsStart: uint32(clusterConfig.ThisNode.HashSlots[0]),
					HashSlotsEnd:   uint32(clusterConfig.ThisNode.HashSlots[1]),
				})

				if err != nil {
					log.Printf("Failed to gossip with node %s", client.GetAddress())
					continue
				}

				for i := range r.OtherNodes {
					otherNode := r.OtherNodes[i]
					otherNodes = append(otherNodes, &configuration.NodeConfig{
						ID:        otherNode.GetNodeId(),
						Address:   otherNode.GetAddress(),
						HashSlots: []int{int(otherNode.GetHashSlotsStart()), int(otherNode.GetHashSlotsEnd())},
					})
				}

				log.Printf("Successfully gossipped with node %s", client.GetAddress())
			}

			gossipClient.configManager.SetClusterConfig(&configuration.ClusterConfig{
				ThisNode:   clusterConfig.ThisNode,
				OtherNodes: otherNodes,
			})
		}
	}()
}

type GossipClientConfig struct {
	RpcClientManager rpc.RpcClientManager
	ConfigManager    configuration.ConfigurationManager
}

func NewGossipClient(config *GossipClientConfig) *GossipClient {
	return &GossipClient{
		rpcClientManager: config.RpcClientManager,
		configManager:    config.ConfigManager,
	}
}
