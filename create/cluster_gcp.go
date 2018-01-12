package create

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/joyent/triton-kubernetes/remote"
	"github.com/joyent/triton-kubernetes/shell"

	"github.com/Jeffail/gabs"
	"github.com/manifoldco/promptui"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

const (
	gcpClusterKeyFormat                     = "cluster_gcp_%s"
	gcpRancherKubernetesTerraformModulePath = "terraform/modules/gcp-rancher-k8s"
)

// This struct represents the definition of a Terraform .tf file.
// Marshalled into json this struct can be passed directly to Terraform.
type gcpClusterTerraformConfig struct {
	baseClusterTerraformConfig

	GCPPathToCredentials string `json:"gcp_path_to_credentials"`
	GCPProjectID         string `json:"gcp_project_id"`
	GCPComputeRegion     string `json:"gcp_compute_region"`
}

func newGCPCluster(selectedClusterManager string, remoteClusterManagerState remote.RemoteClusterManagerStateManta) error {
	baseConfig, err := getBaseClusterTerraformConfig(gcpRancherKubernetesTerraformModulePath)
	if err != nil {
		return err
	}

	cfg := gcpClusterTerraformConfig{
		baseClusterTerraformConfig: baseConfig,
	}

	// GCP path_to_credentials
	rawGCPPathToCredentials := ""
	if viper.IsSet("gcp_path_to_credentials") {
		rawGCPPathToCredentials = viper.GetString("gcp_path_to_credentials")
	} else {
		prompt := promptui.Prompt{
			Label: "Path to Google Cloud Platform Credentials File",
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
		rawGCPPathToCredentials = result
	}

	expandedGCPPathToCredentials, err := homedir.Expand(rawGCPPathToCredentials)
	if err != nil {
		return err
	}
	cfg.GCPPathToCredentials = expandedGCPPathToCredentials

	// GCP Project ID
	if viper.IsSet("gcp_project_id") {
		cfg.GCPProjectID = viper.GetString("gcp_project_id")
	} else {
		prompt := promptui.Prompt{
			Label: "GCP Project ID",
			Validate: func(input string) error {
				if len(input) == 0 {
					return errors.New("GCP Project ID")
				}
				return nil
			},
		}

		result, err := prompt.Run()
		if err != nil {
			return err
		}
		cfg.GCPProjectID = result
	}

	// TODO: GCP Compute Region

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
	clusterKey := fmt.Sprintf(azureClusterKeyFormat, cfg.Name)
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
