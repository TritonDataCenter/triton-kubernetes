package create

import (
	"errors"
	"fmt"
	"os"

	"github.com/joyent/triton-kubernetes/backend"
	"github.com/joyent/triton-kubernetes/state"

	"github.com/manifoldco/promptui"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

const (
	bareMetalRancherKubernetesHostTerraformModulePath = "terraform/modules/bare-metal-rancher-k8s-host"
)

type bareMetalNodeTerraformConfig struct {
	baseNodeTerraformConfig

	Host        string `json:"host"`
	BastionHost string `json:"bastion_host"`
	SSHUser     string `json:"ssh_user"`
	KeyPath     string `json:"key_path"`
}

// Adds new Bare Metal nodes to the given cluster and manager.
// Returns:
// - a slice of the hostnames added
// - the new state
// - error or nil
func newBareMetalNode(selectedClusterManager, selectedCluster string, remoteBackend backend.Backend, currentState state.State) ([]string, error) {
	nonInteractiveMode := viper.GetBool("non-interactive")
	baseConfig, err := getBaseNodeTerraformConfig(bareMetalRancherKubernetesHostTerraformModulePath, selectedCluster, currentState)
	if err != nil {
		return []string{}, err
	}

	cfg := bareMetalNodeTerraformConfig{
		baseNodeTerraformConfig: baseConfig,
	}

	ssh_user := ""
	if viper.IsSet("ssh_user") {
		ssh_user = viper.GetString("ssh_user")
	} else if nonInteractiveMode {
		return []string{}, errors.New("ssh_user must be specified")
	} else {
		prompt := promptui.Prompt{
			Label:   "SSH User",
			Default: "ubuntu",
			Validate: func(input string) error {
				if input == "" {
					return errors.New("Invalid SSH User")
				}
				return nil
			},
		}
		result, err := prompt.Run()
		if err != nil {
			return []string{}, err
		}
		ssh_user = result
	}
	cfg.SSHUser = ssh_user

	bastion_host := ""
	if viper.IsSet("bastion_host") {
		bastion_host = viper.GetString("bastion_host")
	} else if nonInteractiveMode {
		return []string{}, errors.New("bastion_host must be specified")
	} else {
		prompt := promptui.Prompt{
			Label:   "Bastion Host",
			Default: "None",
			Validate: func(input string) error {
				if input == "" {
					return errors.New("Invalid Bastion Host")
				}
				return nil
			},
		}
		result, err := prompt.Run()
		if err != nil {
			return []string{}, err
		}
		if result != "None" {
			bastion_host = result
		}
	}
	cfg.BastionHost = bastion_host

	key_path := ""
	if viper.IsSet("key_path") {
		key_path = viper.GetString("key_path")
	} else if nonInteractiveMode {
		return []string{}, errors.New("key_path must be specified")
	} else {
		prompt := promptui.Prompt{
			Label:   "Key Path",
			Default: "~/.ssh/id_rsa",
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
		}
		result, err := prompt.Run()
		if err != nil {
			return []string{}, err
		}
		key_path = result
	}
	cfg.KeyPath = key_path

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

	// Bare metal node creation requires 1 host/ip address per node.
	hosts := []string{}
	if viper.IsSet("hosts") {
		hosts = viper.GetStringSlice("hosts")
	} else if nonInteractiveMode {
		return []string{}, errors.New("hosts must be specified")
	} else {
		for i := 0; i < cfg.NodeCount; i++ {
			prompt := promptui.Prompt{
				Label: fmt.Sprintf("Host/IP for %s", newHostnames[i]),
				Validate: func(input string) error {
					if input == "" {
						return errors.New("Invalid host/ip")
					}
					return nil
				},
			}
			result, err := prompt.Run()
			if err != nil {
				return []string{}, err
			}
			hosts = append(hosts, result)
		}
	}

	if len(hosts) != cfg.NodeCount {
		return []string{}, errors.New("not enough hosts")
	}

	// Add new node to terraform config with the new hostnames
	for i, newHostname := range newHostnames {
		cfgCopy := cfg
		cfgCopy.Hostname = newHostname
		cfgCopy.Host = hosts[i]
		err = currentState.AddNode(selectedCluster, newHostname, cfgCopy)
		if err != nil {
			return []string{}, err
		}
	}

	return newHostnames, nil
}
