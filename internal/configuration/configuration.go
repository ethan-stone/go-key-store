package configuration

import (
	"encoding/json"
	"io"
	"os"

	"github.com/google/uuid"
)

type ClusterConfig struct {
	Addresses []string `json:"addresses"`
}

type NodeConfig struct {
	Address string
}

func GenerateNodeID() string {
	return uuid.New().String()
}

func LoadClusterConfigFromFile(path string) (*ClusterConfig, error) {
	configFile, err := os.Open("cluster-config.json")

	if err != nil {
		return nil, err
	}

	defer configFile.Close()

	byteResult, _ := io.ReadAll(configFile)

	var clusterConfig ClusterConfig

	err = json.Unmarshal(byteResult, &clusterConfig)

	if err != nil {
		return nil, err
	}

	return &clusterConfig, nil
}
