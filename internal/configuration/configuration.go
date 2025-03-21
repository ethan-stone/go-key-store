package configuration

import (
	"encoding/json"
	"io"
	"log"
	"os"

	"github.com/google/uuid"
)

type ClusterConfigFile struct {
	Nodes []NodeConfig `json:"nodes"`
}

type ClusterConfig struct {
	ThisNode   NodeConfig
	OtherNodes []NodeConfig
}

type NodeConfig struct {
	Address   string `json:"address"`
	HashSlots []int  `json:"hashSlots"` // first element is the start of the range, second element is the end of the range.
}

func GenerateNodeID() string {
	return uuid.New().String()
}

func LoadClusterConfigFromFile(path string, thisNodesAddress string) (*ClusterConfig, error) {
	configFile, err := os.Open("cluster-config.json")

	if err != nil {
		return nil, err
	}

	defer configFile.Close()

	byteResult, _ := io.ReadAll(configFile)

	var clusterConfigFile ClusterConfigFile

	err = json.Unmarshal(byteResult, &clusterConfigFile)

	if err != nil {
		return nil, err
	}

	var thisNode *NodeConfig

	otherNodes := []NodeConfig{}

	for i := range clusterConfigFile.Nodes {
		if clusterConfigFile.Nodes[i].Address == thisNodesAddress {
			thisNode = &NodeConfig{
				Address:   clusterConfigFile.Nodes[i].Address,
				HashSlots: clusterConfigFile.Nodes[i].HashSlots,
			}
		} else {
			otherNodes = append(otherNodes, NodeConfig{
				Address:   clusterConfigFile.Nodes[i].Address,
				HashSlots: clusterConfigFile.Nodes[i].HashSlots,
			})
		}
	}

	if thisNode == nil {
		log.Fatalf("This nodes address is not in the cluster configuration file.")
	}

	return &ClusterConfig{
		ThisNode:   *thisNode,
		OtherNodes: otherNodes,
	}, nil
}
