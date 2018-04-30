package create

import (
	"errors"
	"fmt"
	"os"

	"github.com/joyent/triton-kubernetes/backend"
	"github.com/joyent/triton-kubernetes/state"
	homedir "github.com/mitchellh/go-homedir"

	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

const (
	vSphereRancherKubernetesHostTerraformModulePath = "terraform/modules/vsphere-rancher-k8s-host"
)

type vSphereNodeTerraformConfig struct {
	baseNodeTerraformConfig

	VSphereUser     string `json:"vsphere_user,omitempty"`
	VSpherePassword string `json:"vsphere_password,omitempty"`
	VSphereServer   string `json:"vsphere_server,omitempty"`

	VSphereDatacenterName   string `json:"vsphere_datacenter_name,omitempty"`
	VSphereDatastoreName    string `json:"vsphere_datastore_name,omitempty"`
	VSphereResourcePoolName string `json:"vsphere_resource_pool_name,omitempty"`
	VSphereNetworkName      string `json:"vsphere_network_name,omitempty"`
	VSphereTemplateName     string `json:"vsphere_template_name,omitempty"`

	SSHUser string `json:"ssh_user"`
	KeyPath string `json:"key_path"`
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

		// Grab variables from cluster config
		VSphereUser:     currentState.Get(fmt.Sprintf("module.%s.vsphere_user", selectedCluster)),
		VSpherePassword: currentState.Get(fmt.Sprintf("module.%s.vsphere_password", selectedCluster)),
		VSphereServer:   currentState.Get(fmt.Sprintf("module.%s.vsphere_server", selectedCluster)),

		// Reference terraform output variables from cluster module
		VSphereDatacenterName:   fmt.Sprintf("${module.%s.vsphere_datacenter_name}", selectedCluster),
		VSphereDatastoreName:    fmt.Sprintf("${module.%s.vsphere_datastore_name}", selectedCluster),
		VSphereResourcePoolName: fmt.Sprintf("${module.%s.vsphere_resource_pool_name}", selectedCluster),
		VSphereNetworkName:      fmt.Sprintf("${module.%s.vsphere_network_name}", selectedCluster),
	}

	if viper.IsSet("vsphere_template_name") {
		cfg.VSphereTemplateName = viper.GetString("vsphere_template_name")
	} else if nonInteractiveMode {
		return []string{}, errors.New("vsphere_template_name must be specified")
	} else {
		prompt := promptui.Prompt{
			Label: "VM Template Name",
			Validate: func(input string) error {
				if input == "" {
					return errors.New("Invalid template name")
				}
				return nil
			},
		}
		result, err := prompt.Run()
		if err != nil {
			return nil, err
		}
		cfg.VSphereTemplateName = result
	}

	// SSH User
	if viper.IsSet("ssh_user") {
		cfg.SSHUser = viper.GetString("ssh_user")
	} else if nonInteractiveMode {
		return []string{}, errors.New("ssh_user must be specified")
	} else {
		prompt := promptui.Prompt{
			Label: "SSH User",
			Validate: func(input string) error {
				if input == "" {
					return errors.New("Invalid SSH user")
				}
				return nil
			},
		}
		result, err := prompt.Run()
		if err != nil {
			return nil, err
		}
		cfg.SSHUser = result
	}

	// Private Key Path
	rawKeyPath := ""
	if viper.IsSet("key_path") {
		rawKeyPath = viper.GetString("key_path")
	} else if nonInteractiveMode {
		return []string{}, errors.New("key_path must be specified")
	} else {
		prompt := promptui.Prompt{
			Label: "Private Key Path",
			Validate: func(input string) error {
				expandedPath, err := homedir.Expand(input)
				if err != nil {
					return err
				}

				_, err = os.Stat(expandedPath)
				if err != nil {
					if os.IsNotExist(err) {
						return errors.New("File not found")
					}
				}
				return nil
			},
			Default: "~/.ssh/id_rsa",
		}

		result, err := prompt.Run()
		if err != nil {
			return nil, err
		}
		rawKeyPath = result
	}

	expandedKeyPath, err := homedir.Expand(rawKeyPath)
	if err != nil {
		return nil, err
	}
	cfg.KeyPath = expandedKeyPath

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
		err = currentState.AddNode(selectedCluster, newHostname, cfgCopy)
		if err != nil {
			return []string{}, err
		}
	}

	return newHostnames, nil
}
