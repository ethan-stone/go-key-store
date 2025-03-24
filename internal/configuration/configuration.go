package configuration

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/google/uuid"
)

type ClusterConfigFile struct {
	Nodes []NodeConfig `json:"nodes"`
}

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
