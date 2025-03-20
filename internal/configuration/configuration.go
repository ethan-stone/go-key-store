package configuration

import "github.com/google/uuid"

type ClusterConfig struct {
	NodeID     string
	HashSlots  []int                 // The range hash slots this node is the leader for. First element is the start of the range. Second element is the end of the range. The end of the range is non-inclusive.
	OtherNodes map[string]NodeConfig // A map of node IDs to their configuration.
}

type NodeConfig struct {
	Address string
}

func GenerateNodeID() string {
	return uuid.New().String()
}
