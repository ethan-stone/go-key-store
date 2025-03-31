package cluster

import (
	create_cluster "github.com/ethan-stone/go-key-store/internal/cli/cluster/create"
	"github.com/spf13/cobra"
)

var ClusterCommand = &cobra.Command{
	Use:   "cluster",
	Short: "Perform cluster related operations.",
}

func init() {
	ClusterCommand.AddCommand(create_cluster.CreateClusterCommand)
}
