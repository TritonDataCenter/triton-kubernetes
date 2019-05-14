package create

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/mesoform/triton-kubernetes/backend"
	"github.com/mesoform/triton-kubernetes/state"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2016-06-01/subscriptions"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

const (
	azureRancherKubernetesTerraformModulePath = "terraform/modules/azure-rancher-k8s"
)

// This struct represents the definition of a Terraform .tf file.
// Marshalled into json this struct can be passed directly to Terraform.
type azureClusterTerraformConfig struct {
	baseClusterTerraformConfig

	AzureSubscriptionID string `json:"azure_subscription_id"`
	AzureClientID       string `json:"azure_client_id"`
	AzureClientSecret   string `json:"azure_client_secret"`
	AzureTenantID       string `json:"azure_tenant_id"`
	AzureEnvironment    string `json:"azure_environment"`
	AzureLocation       string `json:"azure_location"`
}

// Returns the name of the cluster that was created and the new state.
func newAzureCluster(remoteBackend backend.Backend, currentState state.State) (string, error) {
	nonInteractiveMode := viper.GetBool("non-interactive")
	baseConfig, err := getBaseClusterTerraformConfig(azureRancherKubernetesTerraformModulePath)
	if err != nil {
		return "", err
	}

	cfg := azureClusterTerraformConfig{
		baseClusterTerraformConfig: baseConfig,
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

	// Add new cluster to terraform config
	err = currentState.AddCluster("azure", cfg.Name, &cfg)
	if err != nil {
		return "", err
	}

	return cfg.Name, nil
}
