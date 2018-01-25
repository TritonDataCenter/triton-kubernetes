package create

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/joyent/triton-kubernetes/backend"
	"github.com/joyent/triton-kubernetes/state"

	"github.com/Azure/azure-sdk-for-go/arm/compute"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/manifoldco/promptui"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

const (
	azureNodeKeyFormat                            = "module.node_azure_%s"
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
}

// Adds new Azure nodes to the given cluster and manager.
// Returns:
// - a slice of the hostnames added
// - the new state
// - error or nil
func newAzureNode(selectedClusterManager, selectedCluster string, remoteBackend backend.Backend, state state.State) ([]string, state.State, error) {
	baseConfig, err := getBaseNodeTerraformConfig(azureRancherKubernetesHostTerraformModulePath, selectedCluster, state)
	if err != nil {
		return []string{}, state, err
	}

	cfg := azureNodeTerraformConfig{
		baseNodeTerraformConfig: baseConfig,

		// Grab variables from cluster config
		AzureSubscriptionID: state.Get(fmt.Sprintf("module.%s.azure_subscription_id", selectedCluster)),
		AzureClientID:       state.Get(fmt.Sprintf("module.%s.azure_client_id", selectedCluster)),
		AzureClientSecret:   state.Get(fmt.Sprintf("module.%s.azure_client_secret", selectedCluster)),
		AzureTenantID:       state.Get(fmt.Sprintf("module.%s.azure_tenant_id", selectedCluster)),
		AzureEnvironment:    state.Get(fmt.Sprintf("module.%s.azure_environment", selectedCluster)),
		AzureLocation:       state.Get(fmt.Sprintf("module.%s.azure_location", selectedCluster)),

		// Reference terraform output variables from cluster module
		AzureResourceGroupName:      fmt.Sprintf("${module.%s.azure_resource_group_name}", selectedCluster),
		AzureNetworkSecurityGroupID: fmt.Sprintf("${module.%s.azure_network_security_group_id}", selectedCluster),
		AzureSubnetID:               fmt.Sprintf("${module.%s.azure_subnet_id}", selectedCluster),
	}

	// Terraform expects public/government/german/china for azure environment
	// Azure SDK expects `Azure{Environment}Cloud`
	azureEnv, err := azure.EnvironmentFromName(fmt.Sprintf("Azure%sCloud", cfg.AzureEnvironment))
	if err != nil {
		return []string{}, state, err
	}

	oauthConfig, err := adal.NewOAuthConfig(azureEnv.ActiveDirectoryEndpoint, cfg.AzureTenantID)
	if err != nil {
		return []string{}, state, err
	}

	azureSPT, err := adal.NewServicePrincipalToken(*oauthConfig, cfg.AzureClientID, cfg.AzureClientSecret, azureEnv.ResourceManagerEndpoint)
	if err != nil {
		return []string{}, state, err
	}

	azureVMSizesClient := compute.NewVirtualMachineSizesClientWithBaseURI(azureEnv.ResourceManagerEndpoint, cfg.AzureSubscriptionID)
	azureVMSizesClient.Authorizer = autorest.NewBearerAuthorizer(azureSPT)

	azureRawVMSizes, err := azureVMSizesClient.List(strings.Replace(strings.ToLower(cfg.AzureLocation), " ", "", -1))
	if err != nil {
		return []string{}, state, err
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
			return []string{}, state, fmt.Errorf("Invalid azure_size '%s', must be one of the following: %s", cfg.AzureSize, strings.Join(azureVMSizes, ", "))
		}
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
			return []string{}, state, err
		}

		cfg.AzureSize = value
	}

	azureImagesClient := compute.NewImagesClientWithBaseURI(azureEnv.ResourceManagerEndpoint, cfg.AzureSubscriptionID)
	azureImagesClient.Authorizer = autorest.NewBearerAuthorizer(azureSPT)

	imageResults, err := azureImagesClient.List()
	if err != nil {
		return []string{}, state, err
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
	} else {
		prompt := promptui.Prompt{
			Label:   "Azure SSH User",
			Default: "root",
		}

		result, err := prompt.Run()
		if err != nil {
			return []string{}, state, err
		}
		cfg.AzureSSHUser = result
	}

	// Azure Public Key Path
	if viper.IsSet("azure_public_key_path") {
		expandedPublicKeyPath, err := homedir.Expand(viper.GetString("azure_public_key_path"))
		if err != nil {
			return []string{}, state, err
		}

		cfg.AzurePublicKeyPath = expandedPublicKeyPath

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
			return []string{}, state, err
		}

		expandedPublicKeyPath, err := homedir.Expand(result)
		if err != nil {
			return []string{}, state, err
		}

		cfg.AzurePublicKeyPath = expandedPublicKeyPath
	}

	// Get existing node names
	nodes, err := state.Nodes(selectedCluster)
	if err != nil {
		return []string{}, state, err
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
		err = state.Add(fmt.Sprintf(azureNodeKeyFormat, newHostname), cfgCopy)
		if err != nil {
			return []string{}, state, err
		}
	}

	return newHostnames, state, nil
}
