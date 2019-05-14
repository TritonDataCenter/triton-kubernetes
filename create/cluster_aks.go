package create

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/mesoform/triton-kubernetes/backend"
	"github.com/mesoform/triton-kubernetes/state"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-04-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2017-09-30/containerservice"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2016-06-01/subscriptions"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/manifoldco/promptui"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

const (
	aksRancherKubernetesTerraformModulePath = "terraform/modules/aks-rancher-k8s"
)

// This struct represents the definition of a Terraform .tf file.
// Marshalled into json this struct can be passed directly to Terraform.
type aksClusterTerraformConfig struct {
	Source string `json:"source"`

	Name string `json:"name"`

	RancherAPIURL    string `json:"rancher_api_url"`
	RancherAccessKey string `json:"rancher_access_key"`
	RancherSecretKey string `json:"rancher_secret_key"`

	AzureSubscriptionID string `json:"azure_subscription_id"`
	AzureClientID       string `json:"azure_client_id"`
	AzureClientSecret   string `json:"azure_client_secret"`
	AzureTenantID       string `json:"azure_tenant_id"`
	AzureEnvironment    string `json:"azure_environment"`
	AzureLocation       string `json:"azure_location"`

	AzureSize          string `json:"azure_size"`
	AzureSSHUser       string `json:"azure_ssh_user"`
	AzurePublicKeyPath string `json:"azure_public_key_path"`

	KubernetesVersion string `json:"k8s_version"`
	NodeCount         int    `json:"node_count"`
}

// Returns the name of the cluster that was created and the new state.
func newAKSCluster(remoteBackend backend.Backend, currentState state.State) (string, error) {
	nonInteractiveMode := viper.GetBool("non-interactive")
	cfg := aksClusterTerraformConfig{
		RancherAPIURL:    "${module.cluster-manager.rancher_url}",
		RancherAccessKey: "${module.cluster-manager.rancher_access_key}",
		RancherSecretKey: "${module.cluster-manager.rancher_secret_key}",
	}

	baseSource := defaultSourceURL
	if viper.IsSet("source_url") {
		baseSource = viper.GetString("source_url")
	}

	baseSourceRef := defaultSourceRef
	if viper.IsSet("source_ref") {
		baseSourceRef = viper.GetString("source_ref")
	}

	// Module Source location e.g. github.com/mesoform/triton-kubernetes//terraform/modules/azure-rancher-k8s?ref=master
	cfg.Source = fmt.Sprintf("%s//%s?ref=%s", baseSource, aksRancherKubernetesTerraformModulePath, baseSourceRef)

	// Name
	clusterNameRegexp := regexp.MustCompile("^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$")
	if viper.IsSet("name") {
		cfg.Name = viper.GetString("name")
	} else if nonInteractiveMode {
		return "", errors.New("name must be specified")
	} else {
		prompt := promptui.Prompt{
			Label: "Cluster Name",
			Validate: func(input string) error {
				if !clusterNameRegexp.MatchString(input) {
					return errors.New("A DNS-1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character")
				}

				return nil
			},
		}

		result, err := prompt.Run()
		if err != nil {
			return "", err
		}
		cfg.Name = result
	}

	if cfg.Name == "" || !clusterNameRegexp.MatchString(cfg.Name) {
		return "", errors.New("Invalid Cluster Name")
	}

	// Azure Subscription ID
	if viper.IsSet("azure_subscription_id") {
		cfg.AzureSubscriptionID = viper.GetString("azure_subscription_id")
	} else if nonInteractiveMode {
		return "", errors.New("azure_subscription_id must be specified")
	} else {
		prompt := promptui.Prompt{
			Label: "Azure Subscription ID",
			Validate: func(input string) error {
				if len(input) == 0 {
					return errors.New("Invalid Azure Subscription ID")
				}
				return nil
			},
		}

		result, err := prompt.Run()
		if err != nil {
			return "", err
		}
		cfg.AzureSubscriptionID = result
	}

	// Azure Client ID
	if viper.IsSet("azure_client_id") {
		cfg.AzureClientID = viper.GetString("azure_client_id")
	} else if nonInteractiveMode {
		return "", errors.New("azure_client_id must be specified")
	} else {
		prompt := promptui.Prompt{
			Label: "Azure Client ID",
			Validate: func(input string) error {
				if len(input) == 0 {
					return errors.New("Invalid Azure Client ID")
				}
				return nil
			},
		}

		result, err := prompt.Run()
		if err != nil {
			return "", err
		}
		cfg.AzureClientID = result
	}

	// Azure Client Secret
	if viper.IsSet("azure_client_secret") {
		cfg.AzureClientSecret = viper.GetString("azure_client_secret")
	} else if nonInteractiveMode {
		return "", errors.New("azure_client_secret must be specified")
	} else {
		prompt := promptui.Prompt{
			Label: "Azure Client Secret",
			Validate: func(input string) error {
				if len(input) == 0 {
					return errors.New("Invalid Azure Client Secret")
				}
				return nil
			},
		}

		result, err := prompt.Run()
		if err != nil {
			return "", err
		}
		cfg.AzureClientSecret = result
	}

	// Azure Tenant ID
	if viper.IsSet("azure_tenant_id") {
		cfg.AzureTenantID = viper.GetString("azure_tenant_id")
	} else if nonInteractiveMode {
		return "", errors.New("azure_tenant_id must be specified")
	} else {
		prompt := promptui.Prompt{
			Label: "Azure Tenant ID",
			Validate: func(input string) error {
				if len(input) == 0 {
					return errors.New("Invalid Azure Tenant ID")
				}
				return nil
			},
		}

		result, err := prompt.Run()
		if err != nil {
			return "", err
		}
		cfg.AzureTenantID = result
	}

	// Azure Environment
	if viper.IsSet("azure_environment") {
		cfg.AzureEnvironment = viper.GetString("azure_environment")
	} else if nonInteractiveMode {
		return "", errors.New("azure_environment must be specified")
	} else {
		prompt := promptui.Select{
			Label: "Azure Environment",
			Items: []string{"public", "government", "german", "china"},
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}?",
				Active:   fmt.Sprintf(`%s {{ . | underline }}`, promptui.IconSelect),
				Inactive: `  {{ . }}`,
				Selected: fmt.Sprintf(`{{ "%s" | green }} {{ "Azure Environment:" | bold}} {{ . }}`, promptui.IconGood),
			},
		}

		_, value, err := prompt.Run()
		if err != nil {
			return "", err
		}

		cfg.AzureEnvironment = value
	}

	// Verify selected azure environment is valid
	if cfg.AzureEnvironment != "public" && cfg.AzureEnvironment != "government" && cfg.AzureEnvironment != "german" && cfg.AzureEnvironment != "china" {
		return "", fmt.Errorf("Invalid azure_environment '%s', must be one of the following: 'public', 'government', 'german', or 'china'", cfg.AzureEnvironment)
	}

	// Terraform expects public/government/german/china for azure environment
	// Azure SDK expects `Azure{Environment}Cloud`
	azureEnv, err := azure.EnvironmentFromName(fmt.Sprintf("Azure%sCloud", cfg.AzureEnvironment))
	if err != nil {
		return "", err
	}

	// We now have enough information to init an azure client
	oauthConfig, err := adal.NewOAuthConfig(azureEnv.ActiveDirectoryEndpoint, cfg.AzureTenantID)
	if err != nil {
		return "", err
	}

	azureSPT, err := adal.NewServicePrincipalToken(*oauthConfig, cfg.AzureClientID, cfg.AzureClientSecret, azureEnv.ResourceManagerEndpoint)
	if err != nil {
		return "", err
	}

	azureGroupClient := subscriptions.NewClientWithBaseURI(azureEnv.ResourceManagerEndpoint)
	azureGroupClient.Authorizer = autorest.NewBearerAuthorizer(azureSPT)

	azureRawLocations, err := azureGroupClient.ListLocations(context.Background(), cfg.AzureSubscriptionID)
	if err != nil {
		return "", err
	}

	azureLocations := []string{}
	for _, loc := range *azureRawLocations.Value {
		azureLocations = append(azureLocations, *loc.DisplayName)
	}

	// Azure Location
	if viper.IsSet("azure_location") {
		cfg.AzureLocation = viper.GetString("azure_location")

		// Verify selected azure location exists
		found := false
		for _, location := range azureLocations {
			if cfg.AzureLocation == location {
				found = true
				break
			}
		}
		if !found {
			return "", fmt.Errorf("Invalid azure_location '%s', must be one of the following: %s", cfg.AzureLocation, strings.Join(azureLocations, ", "))
		}
	} else if nonInteractiveMode {
		return "", errors.New("azure_location must be specified")
	} else {
		prompt := promptui.Select{
			Label: "Azure Location",
			Items: azureLocations,
			Searcher: func(input string, index int) bool {
				name := strings.Replace(strings.ToLower(azureLocations[index]), " ", "", -1)
				input = strings.Replace(strings.ToLower(input), " ", "", -1)
				return strings.Contains(name, input)
			},
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}?",
				Active:   fmt.Sprintf(`%s {{ . | underline }}`, promptui.IconSelect),
				Inactive: `  {{ . }}`,
				Selected: fmt.Sprintf(`{{ "%s" | green }} {{ "Azure Location:" | bold}} {{ . }}`, promptui.IconGood),
			},
		}

		_, value, err := prompt.Run()
		if err != nil {
			return "", err
		}

		cfg.AzureLocation = value
	}

	azureVMSizesClient := compute.NewVirtualMachineSizesClientWithBaseURI(azureEnv.ResourceManagerEndpoint, cfg.AzureSubscriptionID)
	azureVMSizesClient.Authorizer = autorest.NewBearerAuthorizer(azureSPT)

	azureRawVMSizes, err := azureVMSizesClient.List(context.Background(), strings.Replace(strings.ToLower(cfg.AzureLocation), " ", "", -1))
	if err != nil {
		return "", err
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
			return "", fmt.Errorf("Invalid azure_size '%s', must be one of the following: %s", cfg.AzureSize, strings.Join(azureVMSizes, ", "))
		}
	} else if nonInteractiveMode {
		return "", errors.New("azure_size must be specified")
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
			return "", err
		}

		cfg.AzureSize = value
	}

	// Azure SSH User
	if viper.IsSet("azure_ssh_user") {
		cfg.AzureSSHUser = viper.GetString("azure_ssh_user")
	} else if nonInteractiveMode {
		return "", errors.New("azure_ssh_user must be specified")
	} else {
		prompt := promptui.Prompt{
			Label:   "Azure SSH User",
			Default: "ubuntu",
		}

		result, err := prompt.Run()
		if err != nil {
			return "", err
		}
		cfg.AzureSSHUser = result
	}

	// Azure Public Key Path
	if viper.IsSet("azure_public_key_path") {
		expandedPublicKeyPath, err := homedir.Expand(viper.GetString("azure_public_key_path"))
		if err != nil {
			return "", err
		}

		cfg.AzurePublicKeyPath = expandedPublicKeyPath

	} else if nonInteractiveMode {
		return "", errors.New("azure_public_key_path must be specified")
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
			return "", err
		}

		expandedPublicKeyPath, err := homedir.Expand(result)
		if err != nil {
			return "", err
		}

		cfg.AzurePublicKeyPath = expandedPublicKeyPath
	}

	aksClient := containerservice.NewContainerServicesClientWithBaseURI(azureEnv.ResourceManagerEndpoint, cfg.AzureSubscriptionID)
	aksClient.Authorizer = autorest.NewBearerAuthorizer(azureSPT)

	aksVersions, err := aksClient.ListOrchestrators(context.Background(), cfg.AzureLocation, "")
	if err != nil {
		return "", err
	}

	if aksVersions.Orchestrators == nil {
		return "", errors.New("Unable to retrieve available AKS Orchestrators")
	}

	availableVersions := []string{}
	for _, orch := range *aksVersions.Orchestrators {
		if orch.OrchestratorVersion == nil {
			continue
		}

		if orch.OrchestratorType == nil {
			continue
		}

		if *orch.OrchestratorType != string(containerservice.Kubernetes) {
			continue
		}

		availableVersions = append(availableVersions, *orch.OrchestratorVersion)
	}

	// Kubernetes Version
	if viper.IsSet("k8s_version") {
		cfg.KubernetesVersion = viper.GetString("k8s_version")
	} else if nonInteractiveMode {
		return "", errors.New("k8s_version must be specified")
	} else {
		prompt := promptui.Select{
			Label: "Kubernetes Version",
			Items: availableVersions,
			Searcher: func(input string, index int) bool {
				name := strings.Replace(strings.ToLower(availableVersions[index]), " ", "", -1)
				input = strings.Replace(strings.ToLower(input), " ", "", -1)
				return strings.Contains(name, input)
			},
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}?",
				Active:   fmt.Sprintf(`%s {{ . | underline }}`, promptui.IconSelect),
				Inactive: `  {{ . }}`,
				Selected: fmt.Sprintf(`{{ "%s" | green }} {{ "Kubernetes Version:" | bold}} {{ . }}`, promptui.IconGood),
			},
		}

		_, value, err := prompt.Run()
		if err != nil {
			return "", err
		}

		cfg.KubernetesVersion = value
	}

	// Allow user to specify number of nodes to be created.
	var countInput string
	if viper.IsSet("node_count") {
		countInput = viper.GetString("node_count")
	} else {
		prompt := promptui.Prompt{
			Label: "Number of nodes to create",
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
			Default: "3",
		}
		result, err := prompt.Run()
		if err != nil {
			return "", err
		}
		countInput = result
	}

	// Verifying node count
	nodeCount, err := strconv.Atoi(countInput)
	if err != nil {
		return "", fmt.Errorf("node_count must be a valid number. Found '%s'.", countInput)
	}
	if nodeCount <= 0 {
		return "", fmt.Errorf("node_count must be greater than 0. Found '%d'.", nodeCount)
	}

	cfg.NodeCount = nodeCount

	// Add new cluster to terraform config
	err = currentState.AddCluster("aks", cfg.Name, &cfg)
	if err != nil {
		return "", err
	}

	return cfg.Name, nil
}
