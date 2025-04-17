package cluster

import (
	create_cluster "github.com/ethan-stone/go-key-store/internal/cli/cluster/create"
	verify_cluster "github.com/ethan-stone/go-key-store/internal/cli/cluster/verify"
	"github.com/spf13/cobra"
)

var ClusterCommand = &cobra.Command{
	Use:   "cluster",
	Short: "Perform cluster related operations.",
}

func init() {
	ClusterCommand.AddCommand(create_cluster.CreateClusterCommand)
	ClusterCommand.AddCommand(verify_cluster.VerifyClusterCommand)
}
