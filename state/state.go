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

func (state *State) AddManager(tfBackendPath string, tfBackendObj, obj interface{}) error {
	_, err := state.configJSON.SetP(obj, "module.cluster-manager")
	if err != nil {
		return err
	}

	_, err = state.configJSON.SetP(tfBackendObj, tfBackendPath)
	if err != nil {
		return err
	}

	return nil
}

// Clusters are stored at path `module.cluster_{provider}_{clusterName}`
func (state *State) AddCluster(provider, name string, obj interface{}) error {
	_, err := state.configJSON.SetP(obj, fmt.Sprintf("module.cluster_%s_%s", provider, name))
	if err != nil {
		return err
	}

	return nil
}

// Nodes are stored at path `module.node_{provider}_{clusterName}_{nodeName}`
func (state *State) AddNode(clusterKey, name string, obj interface{}) error {
	provider, clusterName, err := getClusterKeyParts(clusterKey)
	if err != nil {
		return err
	}

	_, err = state.configJSON.SetP(obj, fmt.Sprintf("module.node_%s_%s_%s", provider, clusterName, name))
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
// Clusters are stored at path `module.cluster_{provider}_{clusterName}`
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
// Nodes are stored at path `module.node_{provider}_{clusterName}_{nodeName}`
func (state *State) Nodes(clusterKey string) (map[string]string, error) {
	result := map[string]string{}

	provider, name, err := getClusterKeyParts(clusterKey)
	if err != nil {
		return nil, err
	}

	// Nodes are named `node_{provider}_{clusterName}_{nodeName}`
	// nodePrefix is `node_{provider}_{clusterName}_`
	nodePrefix := fmt.Sprintf("node_%s_%s_", provider, name)

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

func getClusterKeyParts(clusterKey string) (provider, name string, err error) {
	parts := strings.Split(clusterKey, "_")
	if len(parts) < 3 {
		err = fmt.Errorf("Could not get cluster key parts, cluster does not follow format `cluster_{provider}_{clusterName}` '%s'", clusterKey)
		return
	}

	provider = parts[1]
	name = parts[2]

	return
}
