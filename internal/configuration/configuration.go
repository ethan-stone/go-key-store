package configuration

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/google/uuid"
)

type NodeBootstrapConfig struct {
	GrpcPort          string   `json:"grpcPort"`
	HttpPort          string   `json:"httpPort"`
	SeedNodeAddresses []string `json:"seedNodeAddresses"`
	HashSlots         []int    `json:"hashSlots"`
}

type ClusterConfig struct {
	ThisNode   *NodeConfig
	OtherNodes []*NodeConfig
}

type NodeConfig struct {
	ID        string
	Address   string `json:"address"`
	HashSlots []int  `json:"hashSlots"` // first element is the start of the range, second element is the end of the range.
}

func GenerateNodeID() string {
	return uuid.New().String()
}

var clusterConfig *ClusterConfig

func SetClusterConfig(config *ClusterConfig) {
	filteredNodes := []*NodeConfig{}
	seenIDs := make(map[string]bool)

	// Add this node's ID to the seen IDs to filter it out from other nodes.
	seenIDs[config.ThisNode.ID] = true

	for _, node := range config.OtherNodes {
		if _, seen := seenIDs[node.ID]; !seen {
			filteredNodes = append(filteredNodes, node)
			seenIDs[node.ID] = true
		}
	}

	config.OtherNodes = filteredNodes
	clusterConfig = config
}

func GetClusterConfig() (*ClusterConfig, error) {
	if clusterConfig == nil {
		return nil, fmt.Errorf("cluster config is not set")
	}

	return clusterConfig, nil
}

func LoadNodeBootstrapConfigFromFile(path string) (*NodeBootstrapConfig, error) {
	configFile, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	defer configFile.Close()

	byteResult, _ := io.ReadAll(configFile)

	var nodeBootstrapConfig NodeBootstrapConfig

	err = json.Unmarshal(byteResult, &nodeBootstrapConfig)

	if err != nil {
		return nil, err
	}

	return &nodeBootstrapConfig, nil
}
