package create

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joyent/triton-kubernetes/backend"
	"github.com/joyent/triton-kubernetes/state"
	"github.com/joyent/triton-kubernetes/util"

	"github.com/Azure/azure-sdk-for-go/arm/compute"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/manifoldco/promptui"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

const (
	azureRancherKubernetesHostTerraformModulePath = "terraform/modules/azure-rancher-k8s-host"
)

type azureNodeTerraformConfig struct {
	baseNodeTerraformConfig

	AzureSubscriptionID string `json:"azure_subscription_id"`
	AzureClientID       string `json:"azure_client_id"`
	AzureClientSecret   string `json:"azure_client_secret"`
	AzureTenantID       string `json:"azure_tenant_id"`
	AzureEnvironment    string `json:"azure_environment"`

	AzureLocation               string `json:"azure_location"`
	AzureResourceGroupName      string `json:"azure_resource_group_name"`
	AzureNetworkSecurityGroupID string `json:"azure_network_security_group_id"`
	AzureSubnetID               string `json:"azure_subnet_id"`

	AzureSize           string `json:"azure_size"`
	AzureImagePublisher string `json:"azure_image_publisher,omitempty"`
	AzureImageOffer     string `json:"azure_image_offer,omitempty"`
	AzureImageSKU       string `json:"azure_image_sku,omitempty"`
	AzureImageVersion   string `json:"azure_image_version,omitempty"`
	AzureSSHUser        string `json:"azure_ssh_user"`
	AzurePublicKeyPath  string `json:"azure_public_key_path"`

	AzureDiskMountPath string `json:"azure_disk_mount_path"`
	AzureDiskSize      string `json:"azure_disk_size"`
}

// Adds new Azure nodes to the given cluster and manager.
// Returns:
// - a slice of the hostnames added
// - the new state
// - error or nil
func newAzureNode(selectedClusterManager, selectedCluster string, remoteBackend backend.Backend, currentState state.State) ([]string, error) {
	nonInteractiveMode := viper.GetBool("non-interactive")
	baseConfig, err := getBaseNodeTerraformConfig(azureRancherKubernetesHostTerraformModulePath, selectedCluster, currentState)
	if err != nil {
		return []string{}, err
	}

	cfg := azureNodeTerraformConfig{
		baseNodeTerraformConfig: baseConfig,

		// Grab variables from cluster config
		AzureSubscriptionID: currentState.Get(fmt.Sprintf("module.%s.azure_subscription_id", selectedCluster)),
		AzureClientID:       currentState.Get(fmt.Sprintf("module.%s.azure_client_id", selectedCluster)),
		AzureClientSecret:   currentState.Get(fmt.Sprintf("module.%s.azure_client_secret", selectedCluster)),
		AzureTenantID:       currentState.Get(fmt.Sprintf("module.%s.azure_tenant_id", selectedCluster)),
		AzureEnvironment:    currentState.Get(fmt.Sprintf("module.%s.azure_environment", selectedCluster)),
		AzureLocation:       currentState.Get(fmt.Sprintf("module.%s.azure_location", selectedCluster)),

		// Reference terraform output variables from cluster module
		AzureResourceGroupName:      fmt.Sprintf("${module.%s.azure_resource_group_name}", selectedCluster),
		AzureNetworkSecurityGroupID: fmt.Sprintf("${module.%s.azure_network_security_group_id}", selectedCluster),
		AzureSubnetID:               fmt.Sprintf("${module.%s.azure_subnet_id}", selectedCluster),
	}

	// Terraform expects public/government/german/china for azure environment
	// Azure SDK expects `Azure{Environment}Cloud`
	azureEnv, err := azure.EnvironmentFromName(fmt.Sprintf("Azure%sCloud", cfg.AzureEnvironment))
	if err != nil {
		return []string{}, err
	}

	oauthConfig, err := adal.NewOAuthConfig(azureEnv.ActiveDirectoryEndpoint, cfg.AzureTenantID)
	if err != nil {
		return []string{}, err
	}

	azureSPT, err := adal.NewServicePrincipalToken(*oauthConfig, cfg.AzureClientID, cfg.AzureClientSecret, azureEnv.ResourceManagerEndpoint)
	if err != nil {
		return []string{}, err
	}

	azureVMSizesClient := compute.NewVirtualMachineSizesClientWithBaseURI(azureEnv.ResourceManagerEndpoint, cfg.AzureSubscriptionID)
	azureVMSizesClient.Authorizer = autorest.NewBearerAuthorizer(azureSPT)

	azureRawVMSizes, err := azureVMSizesClient.List(strings.Replace(strings.ToLower(cfg.AzureLocation), " ", "", -1))
	if err != nil {
		return []string{}, err
	}

	azureVMSizes := []string{}
	for _, size := range *azureRawVMSizes.Value {
		azureVMSizes = append(azureVMSizes, *size.Name)
	}

	// Azure Size
	if viper.IsSet("azure_size") {
		cfg.AzureSize = viper.GetString("azure_size")

		// Verify selected azure size exists
		found := false
		for _, size := range azureVMSizes {
			if cfg.AzureSize == size {
				found = true
				break
			}
		}
		if !found {
			return []string{}, fmt.Errorf("Invalid azure_size '%s', must be one of the following: %s", cfg.AzureSize, strings.Join(azureVMSizes, ", "))
		}
	} else if nonInteractiveMode {
		return []string{}, errors.New("azure_size must be specified")
	} else {
		prompt := promptui.Select{
			Label: "Azure Size",
			Items: azureVMSizes,
			Searcher: func(input string, index int) bool {
				name := strings.Replace(strings.ToLower(azureVMSizes[index]), " ", "", -1)
				input = strings.Replace(strings.ToLower(input), " ", "", -1)
				return strings.Contains(name, input)
			},
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}?",
				Active:   fmt.Sprintf(`%s {{ . | underline }}`, promptui.IconSelect),
				Inactive: `  {{ . }}`,
				Selected: fmt.Sprintf(`{{ "%s" | green }} {{ "Azure Size:" | bold}} {{ . }}`, promptui.IconGood),
			},
		}

		_, value, err := prompt.Run()
		if err != nil {
			return []string{}, err
		}

		cfg.AzureSize = value
	}

	azureImagesClient := compute.NewImagesClientWithBaseURI(azureEnv.ResourceManagerEndpoint, cfg.AzureSubscriptionID)
	azureImagesClient.Authorizer = autorest.NewBearerAuthorizer(azureSPT)

	imageResults, err := azureImagesClient.List()
	if err != nil {
		return []string{}, err
	}

	for _, x := range *imageResults.Value {
		fmt.Println(*x.Name)
		fmt.Print(*x.ID)
	}

	// TODO
	// Azure Image Publisher
	// Azure Image Offer
	// Azure Image SKU
	// Azure Image Version

	// Azure SSH User
	if viper.IsSet("azure_ssh_user") {
		cfg.AzureSSHUser = viper.GetString("azure_ssh_user")
	} else if nonInteractiveMode {
		return []string{}, errors.New("azure_ssh_user must be specified")
	} else {
		prompt := promptui.Prompt{
			Label:   "Azure SSH User",
			Default: "ubuntu",
		}

		result, err := prompt.Run()
		if err != nil {
			return []string{}, err
		}
		cfg.AzureSSHUser = result
	}

	// Azure Public Key Path
	if viper.IsSet("azure_public_key_path") {
		expandedPublicKeyPath, err := homedir.Expand(viper.GetString("azure_public_key_path"))
		if err != nil {
			return []string{}, err
		}

		cfg.AzurePublicKeyPath = expandedPublicKeyPath

	} else if nonInteractiveMode {
		return []string{}, errors.New("azure_public_key_path must be specified")
	} else {
		prompt := promptui.Prompt{
			Label: "Azure Public Key Path",
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
			Default: "~/.ssh/id_rsa.pub",
		}

		result, err := prompt.Run()
		if err != nil {
			return []string{}, err
		}

		expandedPublicKeyPath, err := homedir.Expand(result)
		if err != nil {
			return []string{}, err
		}

		cfg.AzurePublicKeyPath = expandedPublicKeyPath
	}

	// Azure Disk
	diskMountPathIsSet := viper.IsSet("azure_disk_mount_path")
	diskSizeIsSet := viper.IsSet("azure_disk_size")
	if nonInteractiveMode {
		if diskMountPathIsSet && !diskSizeIsSet {
			return nil, errors.New("If azure_disk_mount_path is set, then azure_disk_size must also be set.")
		} else if diskMountPathIsSet {
			cfg.AzureDiskMountPath = viper.GetString("azure_disk_mount_path")
			cfg.AzureDiskSize = viper.GetString("azure_disk_size")
		}
	} else {
		shouldCreateDisk, err := util.PromptForConfirmation("Create a disk for this node", "Disk created")
		if err != nil {
			return nil, err
		}

		if shouldCreateDisk {
			// GCP Disk Mount path
			if diskMountPathIsSet {
				cfg.AzureDiskMountPath = viper.GetString("azure_disk_mount_path")
			} else {
				prompt := promptui.Prompt{
					Label: "Azure Disk Mount Path",
				}

				result, err := prompt.Run()
				if err != nil {
					return nil, err
				}
				cfg.AzureDiskMountPath = result
			}

			if diskSizeIsSet {
				cfg.AzureDiskSize = viper.GetString("azure_disk_size")
			} else {
				prompt := promptui.Prompt{
					Label: "Azure Disk Size in GB",
					Validate: func(input string) error {
						num, err := strconv.ParseInt(input, 10, 64)
						if err != nil {
							return errors.New("Invalid number")
						}
						if num <= 0 {
							return errors.New("Number must be greater than 0")
						}
						return nil
					},
				}
				result, err := prompt.Run()
				if err != nil {
					return nil, err
				}
				cfg.AzureDiskSize = result
			}
		}
	}

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
