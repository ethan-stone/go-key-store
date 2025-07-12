package verify_cluster

import (
	"fmt"

	"github.com/ethan-stone/go-key-store/internal/hash"
	"github.com/ethan-stone/go-key-store/internal/rpc"
	"github.com/spf13/cobra"
)

var VerifyClusterCommand = &cobra.Command{
	Use:   "verify",
	Short: "Verify all hash slots in a cluster are covered.",
	RunE: func(cmd *cobra.Command, args []string) error {
		rpcClientManager := rpc.NewGrpcClientManager()

		client, err := rpcClientManager.GetOrCreateRpcClient(&rpc.RpcClientConfig{
			Address: nodeAddress,
		})

		if err != nil {
			return err
		}

		clusterConfig, err := client.GetClusterConfig(&rpc.GetClusterConfigRequest{})

		if err != nil {
			return err
		}

		totalHashSlotsCovered := clusterConfig.ThisNode.HashSlotsEnd - clusterConfig.ThisNode.HashSlotsStart + 1 // +1 because the range is inclusive

		for _, node := range clusterConfig.OtherNodes {
			totalHashSlotsCovered += (node.HashSlotsEnd - node.HashSlotsStart + 1) // +1 because the range is inclusive
		}

		if totalHashSlotsCovered != hash.NumHashSlots {
			return fmt.Errorf("total hash slots covered (%d) does not match expected (%d)", totalHashSlotsCovered, hash.NumHashSlots)
		}

		fmt.Println("Cluster is valid")

		return nil
	},
}

var nodeAddress string

func init() {
	VerifyClusterCommand.Flags().StringVar(&nodeAddress, "address", "", "The address of any node in the cluster (e.g., --address=localhost:8081)")
	VerifyClusterCommand.MarkFlagRequired("address")
}
