package create

import (
	"errors"
	"os"

	"github.com/mesoform/triton-kubernetes/state"

	"github.com/manifoldco/promptui"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

const (
	bareMetalRancherTerraformModulePath = "terraform/modules/bare-metal-rancher"
)

// This struct represents the definition of a Terraform .tf file.
// Marshalled into json this struct can be passed directly to Terraform.
type bareMetalManagerTerraformConfig struct {
	baseManagerTerraformConfig

	Host        string `json:"host"`
	BastionHost string `json:"bastion_host,omitempty"`
	SSHUser     string `json:"ssh_user,omitempty"`
	KeyPath     string `json:"key_path,omitempty"`
}

func newBareMetalManager(currentState state.State, name string) error {
	nonInteractiveMode := viper.GetBool("non-interactive")

	baseConfig, err := getBaseManagerTerraformConfig(bareMetalRancherTerraformModulePath, name)
	if err != nil {
		return err
	}

	cfg := bareMetalManagerTerraformConfig{
		baseManagerTerraformConfig: baseConfig,
	}

	host := ""
	if viper.IsSet("host") {
		host = viper.GetString("host")
	} else if nonInteractiveMode {
		return errors.New("host must be specified")
	} else {
		prompt := promptui.Prompt{
			Label: "Host/IP for cluster manager",
			Validate: func(input string) error {
				if input == "" {
					return errors.New("Invalid host/ip")
				}
				return nil
			},
		}
		result, err := prompt.Run()
		if err != nil {
			return err
		}

		host = result
	}
	cfg.Host = host

	ssh_user := ""
	if viper.IsSet("ssh_user") {
		ssh_user = viper.GetString("ssh_user")
	} else if nonInteractiveMode {
		return errors.New("ssh_user must be specified")
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
			return err
		}
		ssh_user = result
	}
	cfg.SSHUser = ssh_user

	bastion_host := ""
	if viper.IsSet("bastion_host") {
		bastion_host = viper.GetString("bastion_host")
	} else if nonInteractiveMode {
		return errors.New("bastion_host must be specified")
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
			return err
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
		return errors.New("key_path must be specified")
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
			return err
		}
		key_path = result
	}
	cfg.KeyPath = key_path

	currentState.SetManager(&cfg)

	return nil
}
