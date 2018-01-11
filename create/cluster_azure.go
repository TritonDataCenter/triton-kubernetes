package create

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/joyent/triton-kubernetes/remote"
	"github.com/joyent/triton-kubernetes/shell"

	"github.com/Azure/azure-sdk-for-go/arm/resources/subscriptions"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Jeffail/gabs"
	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

// This struct represents the definition of a Terraform .tf file.
// Marshalled into json this struct can be passed directly to Terraform.
type azureClusterTerraformConfig struct {
	Source string `json:"source"`

	Name string `json:"name"`

	EtcdNodeCount          string `json:"etcd_node_count"`
	OrchestrationNodeCount string `json:"orchestration_node_count"`
	ComputeNodeCount       string `json:"compute_node_count"`

	KubernetesPlaneIsolation string `json:"k8s_plane_isolation"`

	RancherAPIURL    string `json:"rancher_api_url"`
	RancherAccessKey string `json:"rancher_access_key"`
	RancherSecretKey string `json:"rancher_secret_key"`

	AzureSubscriptionID string `json:"azure_subscription_id`
	AzureClientID       string `json:"azure_client_id"`
	AzureClientSecret   string `json:"azure_client_secret"`
	AzureTenantID       string `json:"azure_tenant_id"`
	AzureEnvironment    string `json:"azure_environment"`

	AzureLocation      string `json:"azure_location"`
	AzureSSHUser       string `json:"azure_ssh_user"`
	AzurePublicKeyPath string `json:"azure_public_key_path"`

	RancherRegistry         string `json:"rancher_registry,omitempty"`
	RancherRegistryUsername string `json:"rancher_registry_username,omitempty"`
	RancherRegistryPassword string `json:"rancher_registry_password,omitempty"`

	KubernetesRegistry         string `json:"k8s_registry,omitempty"`
	KubernetesRegistryUsername string `json:"k8s_registry_username,omitempty"`
	KubernetesRegistryPassword string `json:"k8s_registry_password,omitempty"`
}

func newAzureCluster(selectedClusterManager string, remoteClusterManagerState remote.RemoteClusterManagerStateManta) error {
	cfg := azureClusterTerraformConfig{}

	baseSource := "github.com/joyent/triton-kubernetes"
	if viper.IsSet("source_url") {
		baseSource = viper.GetString("source_url")
	}

	cfg.Source = fmt.Sprintf("%s//terraform/modules/azure-rancher-k8s", baseSource)

	// Rancher API URL
	cfg.RancherAPIURL = "http://${element(module.cluster-manager.masters, 0)}:8080"

	// Set node counts to 0 since we manage nodes individually in triton-kubernetes cli
	cfg.EtcdNodeCount = "0"
	cfg.OrchestrationNodeCount = "0"
	cfg.ComputeNodeCount = "0"

	// Name
	if viper.IsSet("name") {
		cfg.Name = viper.GetString("name")
	} else {
		prompt := promptui.Prompt{
			Label: "Cluster Name",
		}

		result, err := prompt.Run()
		if err != nil {
			return err
		}
		cfg.Name = result
	}

	if cfg.Name == "" {
		return errors.New("Invalid Cluster Name")
	}

	// Kubernetes Plane Isolation
	if viper.IsSet("k8s_plane_isolation") {
		cfg.KubernetesPlaneIsolation = viper.GetString("k8s_plane_isolation")
	} else {
		prompt := promptui.Select{
			Label: "Kubernetes Plane Isolation",
			Items: []string{"required", "none"},
		}

		_, value, err := prompt.Run()
		if err != nil {
			return err
		}

		cfg.KubernetesPlaneIsolation = value
	}

	// Verify selected plane isolation is valid
	if cfg.KubernetesPlaneIsolation != "required" && cfg.KubernetesPlaneIsolation != "none" {
		return fmt.Errorf("Invalid k8s_plane_isolation '%s', must be 'required' or 'none'", cfg.KubernetesPlaneIsolation)
	}

	// Azure Subscription ID
	if viper.IsSet("azure_subscription_id") {
		cfg.AzureSubscriptionID = viper.GetString("azure_subscription_id")
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
			return err
		}
		cfg.AzureSubscriptionID = result
	}

	// Azure Client ID
	if viper.IsSet("azure_client_id") {
		cfg.AzureClientID = viper.GetString("azure_client_id")
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
			return err
		}
		cfg.AzureClientID = result
	}

	// Azure Client Secret
	if viper.IsSet("azure_client_secret") {
		cfg.AzureClientSecret = viper.GetString("azure_client_secret")
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
			return err
		}
		cfg.AzureClientSecret = result
	}

	// Azure Tenant ID
	if viper.IsSet("azure_tenant_id") {
		cfg.AzureTenantID = viper.GetString("azure_tenant_id")
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
			return err
		}
		cfg.AzureTenantID = result
	}

	// Azure Environment
	if viper.IsSet("azure_environment") {
		cfg.AzureEnvironment = viper.GetString("azure_environment")
	} else {
		prompt := promptui.Select{
			Label: "Azure Environment",
			Items: []string{"public", "government", "german", "china"},
		}

		_, value, err := prompt.Run()
		if err != nil {
			return err
		}

		cfg.AzureEnvironment = value
	}

	// Verify selected azure environment is valid
	if cfg.AzureEnvironment != "public" && cfg.AzureEnvironment != "government" && cfg.AzureEnvironment != "german" && cfg.AzureEnvironment != "china" {
		return fmt.Errorf("Invalid azure_environment '%s', must be one of the following: 'public', 'government', 'german', or 'china'", cfg.AzureEnvironment)
	}

	// Terraform expects public/government/german/china for azure environment
	// Azure SDK expects `Azure{Environment}Cloud`
	azureEnv, err := azure.EnvironmentFromName(fmt.Sprintf("Azure%sCloud", cfg.AzureEnvironment))
	if err != nil {
		return err
	}

	// We now have enough information to init an azure client
	oauthConfig, err := adal.NewOAuthConfig(azureEnv.ActiveDirectoryEndpoint, cfg.AzureTenantID)
	if err != nil {
		return err
	}

	azureSPT, err := adal.NewServicePrincipalToken(*oauthConfig, cfg.AzureClientID, cfg.AzureClientSecret, azureEnv.ResourceManagerEndpoint)
	if err != nil {
		return err
	}

	azureGroupClient := subscriptions.NewGroupClientWithBaseURI(azureEnv.ResourceManagerEndpoint)
	azureGroupClient.Authorizer = autorest.NewBearerAuthorizer(azureSPT)

	azureRawLocations, err := azureGroupClient.ListLocations(cfg.AzureSubscriptionID)
	if err != nil {
		return err
	}

	azureLocations := []string{}
	for _, loc := range *azureRawLocations.Value {
		azureLocations = append(azureLocations, *loc.DisplayName)
	}

	// Azure Location
	if viper.IsSet("azure_location") {
		cfg.AzureLocation = viper.GetString("azure_location")
	} else {
		prompt := promptui.Select{
			Label: "Azure Location",
			Items: azureLocations,
			Searcher: func(input string, index int) bool {
				name := strings.Replace(strings.ToLower(azureLocations[index]), " ", "", -1)
				input = strings.Replace(strings.ToLower(input), " ", "", -1)
				return strings.Contains(name, input)
			},
		}

		_, value, err := prompt.Run()
		if err != nil {
			return err
		}

		cfg.AzureLocation = value
	}

	// Verify selected auzre location exists
	found := false
	for _, location := range azureLocations {
		if cfg.AzureLocation == location {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("Invalid azure_location '%s', must be one of the following: %s", cfg.AzureLocation, strings.Join(azureLocations, ", "))
	}

	// Load current cluster manager config
	clusterManagerTerraformConfigBytes, err := remoteClusterManagerState.GetTerraformConfig(selectedClusterManager)
	if err != nil {
		return err
	}

	clusterManagerTerraformConfig, err := gabs.ParseJSON(clusterManagerTerraformConfigBytes)
	if err != nil {
		return err
	}

	// Add new cluster to terraform config
	clusterKey := fmt.Sprintf("cluster_%s", cfg.Name)
	clusterManagerTerraformConfig.SetP(&cfg, fmt.Sprintf("module.%s", clusterKey))

	// Create a temporary directory
	tempDir, err := ioutil.TempDir("", "triton-kubernetes-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	// Save the terraform config to the temporary directory
	jsonBytes := []byte(clusterManagerTerraformConfig.StringIndent("", "\t"))
	jsonPath := fmt.Sprintf("%s/%s", tempDir, "main.tf.json")
	err = ioutil.WriteFile(jsonPath, jsonBytes, 0644)
	if err != nil {
		return err
	}

	// Use temporary directory as working directory
	shellOptions := shell.ShellOptions{
		WorkingDir: tempDir,
	}

	// Run terraform init
	err = shell.RunShellCommand(&shellOptions, "terraform", "init", "-force-copy")
	if err != nil {
		return err
	}

	// Run terraform apply
	err = shell.RunShellCommand(&shellOptions, "terraform", "apply", "-auto-approve")
	if err != nil {
		return err
	}

	// After terraform succeeds, commit state
	err = remoteClusterManagerState.CommitTerraformConfig(selectedClusterManager, jsonBytes)
	if err != nil {
		return err
	}

	return nil
}
