package configuration

import (
	"reflect"
	"testing"
)

func TestSetClusterConfig(t *testing.T) {
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
			SetClusterConfig(tt.inputConfig)

			config, _ := GetClusterConfig()

			if !reflect.DeepEqual(config, tt.expectedConfig) {
				t.Errorf("SetClusterConfig() = %v, want %v", config, tt.expectedConfig)
			}

			// Reset clusterConfig for the next test.  Important!
			clusterConfig = nil
		})
	}
}

func TestGetClusterConfig(t *testing.T) {
	t.Run("Not Set", func(t *testing.T) {
		clusterConfig = nil // Ensure it's nil before the test
		_, err := GetClusterConfig()
		if err == nil {
			t.Errorf("GetClusterConfig() should return an error when not set")
		}
	})

	t.Run("Successfully Get", func(t *testing.T) {
		expectedConfig := &ClusterConfig{
			ThisNode:   &NodeConfig{ID: "test-node", Address: "test-address"},
			OtherNodes: []*NodeConfig{},
		}
		SetClusterConfig(expectedConfig)
		cfg, err := GetClusterConfig()

		if err != nil {
			t.Fatalf("GetClusterConfig() returned an error: %v", err)
		}

		if !reflect.DeepEqual(cfg, expectedConfig) {
			t.Errorf("GetClusterConfig() = %v, want %v", cfg, expectedConfig)
		}

		// Clean up for other tests
		clusterConfig = nil
	})
}
