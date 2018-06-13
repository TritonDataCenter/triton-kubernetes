package create

import (
	"errors"
	"fmt"
	"os"
	"strings"

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

	// Bare metal node creation requires 1 host/ip address per node.
	hosts := []string{}
	if viper.IsSet("hosts") {
		hosts = viper.GetStringSlice("hosts")
	} else if nonInteractiveMode {
		return []string{}, errors.New("hosts must be specified")
	} else {
		shouldPrompt := true
		prompt := promptui.Prompt{
			Label: "Host/IP",
			Validate: func(input string) error {
				if input == "" {
					return errors.New("Invalid host/ip")
				}
				return nil
			},
		}

		continueOptions := []struct {
			Name  string
			Value bool
		}{
			{"Yes", true},
			{"No", false},
		}

		continuePrompt := promptui.Select{
			Label: "Add another Host/IP?",
			Items: continueOptions,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}?",
				Active:   fmt.Sprintf("%s {{ .Name | underline }}", promptui.IconSelect),
				Inactive: "  {{.Name}}",
				Selected: "  Add another Host/IP? {{.Name}}",
			},
		}

		for shouldPrompt {
			result, err := prompt.Run()
			if err != nil {
				return []string{}, err
			}
			hosts = append(hosts, result)

			// Continue Prompt
			i, _, err := continuePrompt.Run()
			if err != nil {
				return []string{}, err
			}
			shouldPrompt = continueOptions[i].Value
		}
	}

	// Add new node to terraform config with the host address as the unique name
	for _, host := range hosts {
		cfgCopy := cfg
		cfgCopy.Host = host
		cfgCopy.Hostname = host

		// Lets use the Host/IP as the unique name
		// State names can't contain . so replace with _
		hostname := strings.Replace(host, ".", "_", -1)

		err = currentState.AddNode(selectedCluster, hostname, cfgCopy)
		if err != nil {
			return []string{}, err
		}
	}

	return hosts, nil
}
