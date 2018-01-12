package create

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/joyent/triton-kubernetes/remote"
	"github.com/joyent/triton-kubernetes/shell"

	"github.com/Jeffail/gabs"
)

const (
	azureNodeKeyFormat                            = "node_azure_%s"
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
	AzureImagePublisher string `json:"azure_image_publisher"`
	AzureImageOffer     string `json:"azure_image_offer"`
	AzureImageSKU       string `json:"azure_image_sku"`
	AzureImageVersion   string `json:"azure_image_version"`
	AzureSSHUser        string `json:"azure_ssh_user"`
	AzurePublicKeyPath  string `json:"azure_public_key_path"`
}

func newAzureNode(selectedClusterManager, selectedCluster string, remoteClusterManagerState remote.RemoteClusterManagerStateManta, clusterManagerTerraformConfig *gabs.Container) error {
	baseConfig, err := getBaseNodeTerraformConfig(azureRancherKubernetesHostTerraformModulePath, selectedCluster, clusterManagerTerraformConfig)
	if err != nil {
		return err
	}

	cfg := azureNodeTerraformConfig{
		baseNodeTerraformConfig: baseConfig,

		// Grab variables from cluster config
		AzureSubscriptionID: clusterManagerTerraformConfig.Path(fmt.Sprintf("module.%s.azure_subscription_id", selectedCluster)).Data().(string),
		AzureClientID:       clusterManagerTerraformConfig.Path(fmt.Sprintf("module.%s.azure_client_id", selectedCluster)).Data().(string),
		AzureClientSecret:   clusterManagerTerraformConfig.Path(fmt.Sprintf("module.%s.azure_client_secret", selectedCluster)).Data().(string),
		AzureTenantID:       clusterManagerTerraformConfig.Path(fmt.Sprintf("module.%s.azure_tenant_id", selectedCluster)).Data().(string),
		AzureEnvironment:    clusterManagerTerraformConfig.Path(fmt.Sprintf("module.%s.azure_environment", selectedCluster)).Data().(string),

		// Reference terraform output variables from cluster module
		AzureLocation:               fmt.Sprintf("${module.%s.azure_location}", selectedCluster),
		AzureResourceGroupName:      fmt.Sprintf("${module.%s.azure_resource_group_name}", selectedCluster),
		AzureNetworkSecurityGroupID: fmt.Sprintf("${module.%s.azure_network_security_group_id}", selectedCluster),
		AzureSubnetID:               fmt.Sprintf("${module.%s.azure_subnet_id}", selectedCluster),
	}

	// Azure Size
	// Azure Image Publisher
	// Azure Image Offer
	// Azure Image SKU
	// Azure Image Version
	// Azure SSH User
	// Azure Public Key Path

	// Add new node to terraform config
	nodeKey := fmt.Sprintf(azureNodeKeyFormat, cfg.Hostname)
	clusterManagerTerraformConfig.SetP(&cfg, fmt.Sprintf("module.%s", nodeKey))

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
