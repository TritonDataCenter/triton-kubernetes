package create

import (
	"errors"
	"fmt"

	"github.com/joyent/triton-kubernetes/backend"
	"github.com/joyent/triton-kubernetes/state"

	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

const (
	vSphereNodeKeyFormat                            = "module.node_vsphere_%s"
	vSphereRancherKubernetesHostTerraformModulePath = "terraform/modules/vsphere-rancher-k8s-host"
)

type vSphereNodeTerraformConfig struct {
	baseNodeTerraformConfig

	Host        string `json:"host"`
	BastionHost string `json:"bastion_host"`
	SSHUser     string `json:"ssh_user"`
	KeyPath     string `json:"key_path"`
}

// Adds new vSphere nodes to the given cluster and manager.
// Returns:
// - a slice of the hostnames added
// - the new state
// - error or nil
func newVSphereNode(selectedClusterManager, selectedCluster string, remoteBackend backend.Backend, currentState state.State) ([]string, error) {
	nonInteractiveMode := viper.GetBool("non-interactive")
	baseConfig, err := getBaseNodeTerraformConfig(vSphereRancherKubernetesHostTerraformModulePath, selectedCluster, currentState)
	if err != nil {
		return []string{}, err
	}

	cfg := vSphereNodeTerraformConfig{
		baseNodeTerraformConfig: baseConfig,
	}

	host := ""
	if viper.IsSet("host") {
		host = viper.GetString("host")
	} else if nonInteractiveMode {
		return []string{}, errors.New("host must be specified")
	} else {
		prompt := promptui.Prompt{
			Label: "Host",
			Validate: func(input string) error {
				if input == "" {
					return errors.New("Invalid host")
				}
				return nil
			},
		}
		result, err := prompt.Run()
		if err != nil {
			return nil, err
		}
		host = result
	}
	cfg.Host = host

	// Get existing node names
	nodes, err := currentState.Nodes(selectedCluster)
	if err != nil {
		return []string{}, err
	}
	existingNames := []string{}
	for nodeName := range nodes {
		existingNames = append(existingNames, nodeName)
	}

	// Determine what the hostnames should be for the new node(s)
	newHostnames := getNewHostnames(existingNames, cfg.Hostname, cfg.NodeCount)

	// Add new node to terraform config with the new hostnames
	for _, newHostname := range newHostnames {
		cfgCopy := cfg
		cfgCopy.Hostname = newHostname
		err = currentState.Add(fmt.Sprintf(vSphereNodeKeyFormat, newHostname), cfgCopy)
		if err != nil {
			return []string{}, err
		}
	}

	return newHostnames, nil
}
