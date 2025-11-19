package add_node

import (
	"fmt"

	"github.com/ethan-stone/go-key-store/internal/rpc"
	"github.com/spf13/cobra"
)

var AddNodeCommand = &cobra.Command{
	Use:   "add_node",
	Short: "Add a node to a cluster. The node will not handle any hash slots until you reshard the cluster.",
	RunE: func(cmd *cobra.Command, args []string) error {
		rpcClientManager := rpc.NewGrpcClientManager(rpc.NewRpcClient)

		clusterNodeClient, err := rpcClientManager.GetOrCreateRpcClient(&rpc.RpcClientConfig{
			Address: clusterNodeAddress,
		})

		if err != nil {
			return err
		}

		clusterNodeClusterConfig, err := clusterNodeClient.GetClusterConfig(&rpc.GetClusterConfigRequest{})

		if err != nil {
			return err
		}

		clusterContainsNewNode := false

		for _, node := range clusterNodeClusterConfig.OtherNodes {
			if node.Address == newNodeAddress {
				clusterContainsNewNode = true
				break
			}
		}

		if clusterContainsNewNode {
			return fmt.Errorf("new node %s is already a part of the cluster", newNodeAddress)
		}

		newNodeClient, err := rpcClientManager.GetOrCreateRpcClient(&rpc.RpcClientConfig{
			Address: newNodeAddress,
		})

		if err != nil {
			return err
		}

		newNodeClusterConfig, err := newNodeClient.GetClusterConfig(&rpc.GetClusterConfigRequest{})

		if err != nil {
			return err
		}

		if len(newNodeClusterConfig.OtherNodes) > 0 {
			return fmt.Errorf("new node %s is already a part of a cluster", newNodeAddress)
		}

		allNodes := []*rpc.NodeConfig{}

		allNodes = append(allNodes, clusterNodeClusterConfig.ThisNode)

		allNodes = append(allNodes, clusterNodeClusterConfig.OtherNodes...)

		for _, node := range allNodes {
			nodeClient, err := rpcClientManager.GetOrCreateRpcClient(&rpc.RpcClientConfig{
				Address: node.Address,
			})

			if err != nil {
				return err
			}

			nodeClient.SetClusterConfig(&rpc.SetClusterConfigRequest{
				ThisNode: &rpc.SetNodeConfigOptions{
					HashSlotsStart: node.HashSlotsStart,
					HashSlotsEnd:   node.HashSlotsEnd,
				},
				OtherNodes: allNodes,
			})
		}

		return nil
	},
}

var newNodeAddress string     // address of the node to add to the cluster
var clusterNodeAddress string // address of any node already in the cluster

func init() {
	AddNodeCommand.Flags().StringVar(&newNodeAddress, "new-node-address", "", "Address of the node to add to the cluster")
	AddNodeCommand.Flags().StringVar(&clusterNodeAddress, "cluster-node-address", "", "Address of any node already in the cluster")
	AddNodeCommand.MarkFlagRequired("new-node-address")
	AddNodeCommand.MarkFlagRequired("cluster-node-address")
}
