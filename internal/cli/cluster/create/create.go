package create_cluster

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ethan-stone/go-key-store/internal/hash"
	"github.com/ethan-stone/go-key-store/internal/rpc"
	"github.com/spf13/cobra"
)

var CreateClusterCommand = &cobra.Command{
	Use:   "create",
	Short: "Configure a set of nodes to be in a cluster.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. [x] Attempt to establish connections to all nodes. Error if can't.
		// 2. [x] Create suggested config. Divide hash slots evenly.
		// 3. [x] Ask for confirmation of config.
		// 4. [x] If yes, apply config.

		rpcClientManager := rpc.NewGrpcClientManager()

		hashSlotRanges := hash.CalculateHashSlotRanges(len(nodeAddresses), 16384)

		fmt.Println("hash slots", hashSlotRanges)

		for i := range nodeAddresses {
			address := nodeAddresses[i]

			_, err := rpcClientManager.GetOrCreateRpcClient(&rpc.RpcClientConfig{
				Address: address,
			})

			if err != nil {
				return err
			}
		}

		nodes := []*rpc.NodeConfig{}

		for i := range nodeAddresses {
			address := nodeAddresses[i]

			client, err := rpcClientManager.GetOrCreateRpcClient(&rpc.RpcClientConfig{
				Address: address,
			})

			if err != nil {
				return err
			}

			getClusterConfigResponse, err := client.GetClusterConfig(&rpc.GetClusterConfigRequest{})

			if err != nil || !getClusterConfigResponse.GetOk() {
				return err
			}

			if getClusterConfigResponse.GetOtherNodes() != nil && len(getClusterConfigResponse.GetOtherNodes()) > 0 {
				return fmt.Errorf("node %s is already a part of a cluster", address)
			}

			hashSlotRange := hashSlotRanges[i+1]

			// when creating a cluster it is assumed all the nodes are independently running
			nodes = append(nodes, &rpc.NodeConfig{
				NodeId:         getClusterConfigResponse.GetThisNode().GetNodeId(),
				Address:        getClusterConfigResponse.GetThisNode().GetAddress(),
				HashSlotsStart: uint32(hashSlotRange[0]),
				HashSlotsEnd:   uint32(hashSlotRange[1]),
			})
		}

		for i := range nodes {
			node := nodes[i]

			fmt.Printf("  %s -> Slots: %d to %d\n", node.Address, node.HashSlotsStart, node.HashSlotsEnd)
		}

		confirmed := confirm("Are you sure you want to apply this configuration?")

		if !confirmed {
			fmt.Println("\nConfiguration not applied")
			return nil
		}

		for i := range nodes {
			node := nodes[i]

			client, err := rpcClientManager.GetOrCreateRpcClient(&rpc.RpcClientConfig{
				Address: node.Address,
			})

			if err != nil {
				return err
			}

			client.SetClusterConfig(&rpc.SetClusterConfigRequest{
				ThisNode: &rpc.SetNodeConfigOptions{
					HashSlotsStart: node.HashSlotsStart,
					HashSlotsEnd:   node.HashSlotsEnd,
				},
				OtherNodes: nodes,
			})
		}

		return nil
	},
}

var nodeAddresses []string

func init() {
	CreateClusterCommand.Flags().StringSliceVar(&nodeAddresses, "addresses", []string{}, "A list of node addresses, separated by commas (e.g., --addresses=localhost:8080,localhost:8081)")
	CreateClusterCommand.MarkFlagRequired("addresses")
}

func confirm(s string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/n]: ", s)

		response, err := reader.ReadString('\n')
		if err != nil {
			return false // Assume no confirmation on error
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}
