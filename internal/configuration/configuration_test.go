package configuration

import (
	"reflect"
	"testing"
)

func TestBaseConfigurationManager_SetClusterConfig(t *testing.T) {
	node1 := &NodeConfig{ID: "node1", Address: "addr1", HashSlots: []int{1, 2}}
	node2 := &NodeConfig{ID: "node2", Address: "addr2", HashSlots: []int{3, 4}}
	node3 := &NodeConfig{ID: "node3", Address: "addr3", HashSlots: []int{5, 6}}

	tests := []struct {
		name           string
		inputConfig    *ClusterConfig
		expectedConfig *ClusterConfig
	}{
		{
			name: "No Other Nodes",
			inputConfig: &ClusterConfig{
				ThisNode:   node1,
				OtherNodes: []*NodeConfig{},
			},
			expectedConfig: &ClusterConfig{
				ThisNode:   node1,
				OtherNodes: []*NodeConfig{},
			},
		},
		{
			name: "Single Other Node",
			inputConfig: &ClusterConfig{
				ThisNode:   node1,
				OtherNodes: []*NodeConfig{node2},
			},
			expectedConfig: &ClusterConfig{
				ThisNode:   node1,
				OtherNodes: []*NodeConfig{node2},
			},
		},
		{
			name: "Multiple Other Nodes",
			inputConfig: &ClusterConfig{
				ThisNode:   node1,
				OtherNodes: []*NodeConfig{node2, node3},
			},
			expectedConfig: &ClusterConfig{
				ThisNode:   node1,
				OtherNodes: []*NodeConfig{node2, node3},
			},
		},
		{
			name: "Duplicate Other Nodes",
			inputConfig: &ClusterConfig{
				ThisNode:   node1,
				OtherNodes: []*NodeConfig{node2, node2},
			},
			expectedConfig: &ClusterConfig{
				ThisNode:   node1,
				OtherNodes: []*NodeConfig{node2},
			},
		},
		{
			name: "This Node in Other Nodes",
			inputConfig: &ClusterConfig{
				ThisNode:   node1,
				OtherNodes: []*NodeConfig{node1, node2},
			},
			expectedConfig: &ClusterConfig{
				ThisNode:   node1,
				OtherNodes: []*NodeConfig{node2},
			},
		},
		{
			name: "This Node and Duplicate in Other Nodes",
			inputConfig: &ClusterConfig{
				ThisNode:   node1,
				OtherNodes: []*NodeConfig{node1, node1, node2},
			},
			expectedConfig: &ClusterConfig{
				ThisNode:   node1,
				OtherNodes: []*NodeConfig{node2},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := NewBaseConfigurationManager(nil) // Initialize with nil or an initial config
			cm.SetClusterConfig(tt.inputConfig)

			config := cm.GetClusterConfig()

			if !reflect.DeepEqual(config, tt.expectedConfig) {
				t.Errorf("SetClusterConfig() = %v, want %v", config,
					tt.expectedConfig)
			}
		})
	}
}

func TestBaseConfigurationManager_GetClusterConfig(t *testing.T) {
	t.Run("Successfully Get", func(t *testing.T) {
		expectedConfig := &ClusterConfig{
			ThisNode:   &NodeConfig{ID: "test-node", Address: "test-address"},
			OtherNodes: []*NodeConfig{},
		}
		cm := NewBaseConfigurationManager(expectedConfig)
		cfg := cm.GetClusterConfig()

		if !reflect.DeepEqual(cfg, expectedConfig) {
			t.Errorf("GetClusterConfig() = %v, want %v", cfg, expectedConfig)
		}
	})

	t.Run("Nil Config", func(t *testing.T) {
		cm := NewBaseConfigurationManager(nil)
		cfg := cm.GetClusterConfig()

		if cfg != nil {
			t.Errorf("GetClusterConfig() should return nil when not set, got %v",
				cfg)
		}
	})
}
