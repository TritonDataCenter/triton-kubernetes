package state

import (
	"fmt"
	"strings"

	"github.com/Jeffail/gabs"
)

type State struct {
	Name       string
	configJSON *gabs.Container
}

func New(name string, raw []byte) (State, error) {
	config, err := gabs.ParseJSON(raw)
	if err != nil {
		return State{}, err
	}

	return State{
		Name:       name,
		configJSON: config,
	}, nil
}

func (state *State) Get(path string) string {
	value, ok := state.configJSON.Path(path).Data().(string)
	if !ok {
		return ""
	}

	return value
}

func (state *State) Add(path string, obj interface{}) error {
	_, err := state.configJSON.SetP(obj, path)
	if err != nil {
		return err
	}

	return nil
}

func (state *State) Delete(path string) error {
	err := state.configJSON.DeleteP(path)
	if err != nil {
		return err
	}

	return nil
}

func (state *State) Bytes() []byte {
	return state.configJSON.BytesIndent("", "\t")
}

// Returns map of cluster name to cluster key
// cluster keys are prefixed with 'cluster_'
func (state *State) Clusters() (map[string]string, error) {
	result := map[string]string{}

	children, err := state.configJSON.S("module").ChildrenMap()
	if err != nil {
		return nil, err
	}

	for key, child := range children {
		if strings.Index(key, "cluster_") == 0 {
			name, ok := child.Path("name").Data().(string)
			if !ok {
				continue
			}
			result[name] = key
		}
	}

	return result, nil
}

// Returns map of node name to node key for all nodes in a cluster
// node keys are prefixed with 'node_'
func (state *State) Nodes(clusterKey string) (map[string]string, error) {
	result := map[string]string{}

	parts := strings.Split(clusterKey, "_")
	if len(parts) < 3 {
		// clusterKey is `cluster_{provider}_{clusterName}`
		return result, fmt.Errorf("Could not determine cloud provider for cluster '%s'", clusterKey)
	}

	cloudProvider := parts[1]
	clusterName := parts[2]

	// Nodes are named `node_{provider}_{clusterName}-{nodeName}-{nodeNumber}
	// nodePrefix is `node_{provider}_{clusterName}`
	nodePrefix := fmt.Sprintf("node_%s_%s", cloudProvider, clusterName)

	children, err := state.configJSON.S("module").ChildrenMap()
	if err != nil {
		return nil, err
	}

	for key, child := range children {
		if strings.Index(key, nodePrefix) == 0 {
			// Retrieving hostname
			hostname, ok := child.Path("hostname").Data().(string)
			if !ok {
				continue
			}

			result[hostname] = key
		}
	}

	return result, nil
}
