package client

import (
	"log"

	"github.com/ethan-stone/go-key-store/internal/cli/cluster"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "go-store",
	Short: "CLI for interacting with go-store",
	Long:  "This is the CLI application for interacting with go-store. You can perform basic operations and configure clusters.",
}

var versionCommand = &cobra.Command{
	Use:   "version",
	Short: "Print the version of go-store",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("version command used.")
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(versionCommand)
	rootCmd.AddCommand(cluster.ClusterCommand)
}
