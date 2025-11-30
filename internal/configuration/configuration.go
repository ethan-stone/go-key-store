package configuration

import (
	"encoding/json"
	"fmt"
	"os"
)

// import (
// 	"encoding/json"
// 	"io"
// 	"log"
// 	"os"

// 	"github.com/google/uuid"
// )

// type NodeBootstrapConfig struct {
// 	GrpcPort          string   `json:"grpcPort"`
// 	HttpPort          string   `json:"httpPort"`
// 	SeedNodeAddresses []string `json:"seedNodeAddresses"`
// 	HashSlots         []int    `json:"hashSlots"`
// }

// type ClusterConfig struct {
// 	ThisNode     *NodeConfig
// 	OtherNodes   []*NodeConfig
// 	ReplicaNodes []*NodeConfig
// }

// type NodeConfig struct {
// 	ID        string `json:"id"`
// 	Address   string `json:"address"`
// 	HashSlots []int  `json:"hashSlots"` // First element is the start of the range, second element is the end of the range. Both sides are inclusive.
// }

// func GenerateNodeID() string {
// 	return uuid.New().String()
// }

type ConfigurationManager interface {
	LoadClusterConfig(path string) error
	SetCurrentNodeID(id string) error
	GetCurrentNodeConfig() (*NodeConfig, *ShardGroup, *ClusterConfig, error)
}

type BaseConfigurationManager struct {
	currentNodeID string // the id of this particular node
	clusterConfig *ClusterConfig
}

func (cm *BaseConfigurationManager) LoadClusterConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var cluster ClusterConfig
	if err := json.Unmarshal(data, &cluster); err != nil {
		return err
	}
	cm.clusterConfig = &cluster
	return nil
}

func (cm *BaseConfigurationManager) SetCurrentNodeID(id string) error {
	cm.currentNodeID = id
	return nil
}

func (cm *BaseConfigurationManager) GetCurrentNodeConfig() (*NodeConfig, *ShardGroup, *ClusterConfig, error) {
	for _, shardGroup := range cm.clusterConfig.ShardGroups {
		for _, node := range shardGroup.Nodes {
			if node.ID == cm.currentNodeID {
				return &node, &shardGroup, cm.clusterConfig, nil
			}
		}
	}
	return nil, nil, nil, fmt.Errorf("node %s not found", cm.currentNodeID)
}

func (cm *BaseConfigurationManager) GetClusterConfig() *ClusterConfig {
	return cm.clusterConfig
}

type NodeConfig struct {
	ID          string `json:"id"`
	Role        string `json:"role"` // "leader" or "follower"
	HttpAddress string `json:"http_address"`
	GrpcAddress string `json:"grpc_address"`
	DataDir     string `json:"data_dir"`
}

type ShardGroup struct {
	ID        int          `json:"id"`
	HashSlots []int        `json:"hash_slots"`
	Nodes     []NodeConfig `json:"nodes"`
}

type ClusterConfig struct {
	HashSlots   int          `json:"hash_slots"`
	ShardGroups []ShardGroup `json:"shard_groups"`
}

func NewBaseConfigurationManager() *BaseConfigurationManager {
	return &BaseConfigurationManager{}
}
