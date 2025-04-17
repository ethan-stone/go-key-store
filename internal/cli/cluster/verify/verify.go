package verify_cluster

import (
	"github.com/spf13/cobra"
)

var VerifyClusterCommand = &cobra.Command{
	Use:   "verify",
	Short: "Verify all hash slots in a cluster are covered.",
	RunE: func(cmd *cobra.Command, args []string) error {

		return nil
	},
}

var nodeAddress string

func init() {
	VerifyClusterCommand.Flags().StringVar(&nodeAddress, "address", "", "The address of any node in the cluster (e.g., --address=localhost:8081)")
	VerifyClusterCommand.MarkFlagRequired("address")
}
