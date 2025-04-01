package configuration

import (
	"encoding/json"
	"io"
	"log"
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
	ID        string `json:"id"`
	Address   string `json:"address"`
	HashSlots []int  `json:"hashSlots"` // first element is the start of the range, second element is the end of the range.
}

func GenerateNodeID() string {
	return uuid.New().String()
}

var clusterConfig *ClusterConfig

type ConfigurationManager interface {
	SetClusterConfig(config *ClusterConfig)
	GetClusterConfig() *ClusterConfig
}

type BaseConfigurationManager struct {
	clusterConfig *ClusterConfig
}

func (cm *BaseConfigurationManager) SetClusterConfig(config *ClusterConfig) {
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
	cm.clusterConfig = config

	// Log the new cluster config in JSON format
	jsonConfig, err := json.Marshal(clusterConfig)
	if err != nil {
		log.Printf("Error marshaling cluster config to JSON: %v", err) // use log instead of fmt
		return                                                         // Important: Exit the function to prevent further errors
	}

	log.Printf("New cluster config: %s", jsonConfig) // use log instead of fmt to follow conventions.
}

func (cm *BaseConfigurationManager) GetClusterConfig() *ClusterConfig {
	return cm.clusterConfig
}

func NewBaseConfigurationManager(initialConfig *ClusterConfig) *BaseConfigurationManager {
	return &BaseConfigurationManager{
		clusterConfig: initialConfig,
	}
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
