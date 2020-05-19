package create

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/joyent/triton-kubernetes/backend"
	"github.com/joyent/triton-kubernetes/shell"
	"github.com/joyent/triton-kubernetes/util"

	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

type baseManagerTerraformConfig struct {
	Source string `json:"source"`

	Name string `json:"name"`

	RancherAdminPassword    string `json:"rancher_admin_password,omitempty"`
	RancherServerImage      string `json:"rancher_server_image,omitempty"`
	RancherAgentImage       string `json:"rancher_agent_image,omitempty"`
	RancherRegistry         string `json:"rancher_registry,omitempty"`
	RancherRegistryUsername string `json:"rancher_registry_username,omitempty"`
	RancherRegistryPassword string `json:"rancher_registry_password,omitempty"`
}

func NewManager(remoteBackend backend.Backend) error {
	nonInteractiveMode := viper.GetBool("non-interactive")

	selectedCloudProvider := ""
	if viper.IsSet("manager_cloud_provider") {
		selectedCloudProvider = viper.GetString("manager_cloud_provider")
	} else if nonInteractiveMode {
		return errors.New("manager_cloud_provider must be specified")
	} else {
		prompt := promptui.Select{
			Label: "Create Manager in which Cloud Provider",
			Items: []string{"Triton", "AWS", "GCP", "Azure", "BareMetal"},
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}?",
				Active:   fmt.Sprintf(`%s {{ . | underline }}`, promptui.IconSelect),
				Inactive: `  {{ . }}`,
				Selected: fmt.Sprintf(`{{ "%s" | green }} {{ "Cloud Provider:" | bold}} {{ . }}`, promptui.IconGood),
			},
		}

		_, value, err := prompt.Run()
		if err != nil {
			return err
		}

		selectedCloudProvider = strings.ToLower(value)
	}

	// Name
	name := ""
	if viper.IsSet("name") {
		name = viper.GetString("name")
	} else if nonInteractiveMode {
		return errors.New("name must be specified")
	} else {
		prompt := promptui.Prompt{
			Label: "Cluster Manager Name",
			Validate: func(input string) error {
				if input == "" {
					return errors.New("manager name cannot be blank")
				}

				return nil
			},
		}

		result, err := prompt.Run()
		if err != nil {
			return err
		}
		name = result
	}

	if name == "" {
		return errors.New("Invalid Cluster Manager Name")
	}

	// Validate that a cluster manager with the same name doesn't already exist.
	existingClusterManagers, err := remoteBackend.States()
	if err != nil {
		return err
	}

	found := false
	for _, clusterManagerName := range existingClusterManagers {
		if name == clusterManagerName {
			found = true
			break
		}
	}
	if found {
		return fmt.Errorf("A Cluster Manager with the name '%s' already exists.", name)
	}

	currentState, err := remoteBackend.State(name)
	if err != nil {
		return err
	}

	switch selectedCloudProvider {
	case "triton":
		err = newTritonManager(currentState, name)
	case "aws":
		err = newAWSManager(currentState, name)
	case "gcp":
		err = newGCPManager(currentState, name)
	case "azure":
		err = newAzureManager(currentState, name)
	case "baremetal":
		err = newBareMetalManager(currentState, name)
	// case "vsphere":
	default:
		return fmt.Errorf("Unsupported cloud provider '%s', cannot create manager", selectedCloudProvider)
	}
	if err != nil {
		return err
	}

	if !nonInteractiveMode {
		label := "Proceed with the manager creation"
		selected := "Proceed"
		confirmed, err := util.PromptForConfirmation(label, selected)
		if err != nil {
			return err
		}
		if !confirmed {
			fmt.Println("Manager creation canceled.")
			return nil
		}
	}

	currentState.SetTerraformBackendConfig(remoteBackend.StateTerraformConfig(name))

	err = shell.RunTerraformApplyWithState(currentState)
	if err != nil {
		return err
	}

	// After terraform succeeds, commit state
	err = remoteBackend.PersistState(currentState)
	if err != nil {
		return err
	}

	return nil
}

func getBaseManagerTerraformConfig(terraformModulePath, name string) (baseManagerTerraformConfig, error) {
	nonInteractiveMode := viper.GetBool("non-interactive")
	cfg := baseManagerTerraformConfig{}

	baseSource := defaultSourceURL
	if viper.IsSet("source_url") {
		baseSource = viper.GetString("source_url")
	}

	baseSourceRef := defaultSourceRef
	if viper.IsSet("source_ref") {
		baseSourceRef = viper.GetString("source_ref")
	}

	_, err := os.Stat(baseSource)
	if err != nil {
		// Module Source location e.g. github.com/joyent/triton-kubernetes//terraform/modules/triton-rancher?ref=master
		cfg.Source = fmt.Sprintf("%s//%s?ref=%s", baseSource, terraformModulePath, baseSourceRef)
	} else {
		// This is a local file, ignore ref
		cfg.Source = fmt.Sprintf("%s//%s", baseSource, terraformModulePath)
	}

	cfg.Name = name

	// Rancher Docker Registry
	if viper.IsSet("private_registry") {
		cfg.RancherRegistry = viper.GetString("private_registry")
	} else if !nonInteractiveMode {
		prompt := promptui.Prompt{
			Label:   "Private Registry",
			Default: "None",
		}

		result, err := prompt.Run()
		if err != nil {
			return baseManagerTerraformConfig{}, err
		}

		if result != "None" {
			cfg.RancherRegistry = result
		}
	}

	// Ask for rancher registry username/password only if rancher registry is given
	if cfg.RancherRegistry != "" {
		// Rancher Registry Username
		if viper.IsSet("private_registry_username") {
			cfg.RancherRegistryUsername = viper.GetString("private_registry_username")
		} else if nonInteractiveMode {
			return baseManagerTerraformConfig{}, errors.New("private_registry_username must be specified")
		} else {
			prompt := promptui.Prompt{
				Label: "Private Registry Username",
			}

			result, err := prompt.Run()
			if err != nil {
				return baseManagerTerraformConfig{}, err
			}
			cfg.RancherRegistryUsername = result
		}

		// Rancher Registry Password
		if viper.IsSet("private_registry_password") {
			cfg.RancherRegistryPassword = viper.GetString("private_registry_password")
		} else if nonInteractiveMode {
			return baseManagerTerraformConfig{}, errors.New("private_registry_password must be specified")
		} else {
			prompt := promptui.Prompt{
				Label: "Private Registry Password",
				Mask:  '*',
			}

			result, err := prompt.Run()
			if err != nil {
				return baseManagerTerraformConfig{}, err
			}
			cfg.RancherRegistryPassword = result
		}
	}

	// Rancher Server Image
	if viper.IsSet("rancher_server_image") {
		cfg.RancherServerImage = viper.GetString("rancher_server_image")
	} else if !nonInteractiveMode {
		prompt := promptui.Prompt{
			Label:   "Rancher Server Image",
			Default: "Default",
		}

		result, err := prompt.Run()
		if err != nil {
			return baseManagerTerraformConfig{}, err
		}

		if result != "Default" {
			cfg.RancherServerImage = result
		}
	}

	// Rancher Agent Image
	if viper.IsSet("rancher_agent_image") {
		cfg.RancherAgentImage = viper.GetString("rancher_agent_image")
	} else if !nonInteractiveMode {
		prompt := promptui.Prompt{
			Label:   "Rancher Agent Image",
			Default: "Default",
		}

		result, err := prompt.Run()
		if err != nil {
			return baseManagerTerraformConfig{}, err
		}

		if result != "Default" {
			cfg.RancherAgentImage = result
		}
	}

	// Rancher Admin Password
	if viper.IsSet("rancher_admin_password") {
		cfg.RancherAdminPassword = viper.GetString("rancher_admin_password")
	} else if nonInteractiveMode {
		return baseManagerTerraformConfig{}, errors.New("UI Admin Password must be specified")
	} else {
		prompt := promptui.Prompt{
			Label: "Set UI Admin Password",
			Mask:  '*',
			Validate: func(input string) error {
				if input == "" {
					return errors.New("password cannot be blank")
				}

				return nil
			},
		}

		result, err := prompt.Run()
		if err != nil {
			return baseManagerTerraformConfig{}, err
		}
		cfg.RancherAdminPassword = result
	}

	if cfg.RancherAdminPassword == "" {
		return baseManagerTerraformConfig{}, errors.New("Invalid UI Admin password")
	}

	return cfg, nil
}
