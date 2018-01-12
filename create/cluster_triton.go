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
	tritonClusterKeyFormat                     = "cluster_triton_%s"
	tritonRancherKubernetesTerraformModulePath = "terraform/modules/triton-rancher-k8s"
)

// This struct represents the definition of a Terraform .tf file.
// Marshalled into json this struct can be passed directly to Terraform.
type tritonClusterTerraformConfig struct {
	baseClusterTerraformConfig

	TritonAccount string `json:"triton_account"`
	TritonKeyPath string `json:"triton_key_path"`
	TritonKeyID   string `json:"triton_key_id"`
	TritonURL     string `json:"triton_url,omitempty"`
}

func newTritonCluster(selectedClusterManager string, remoteClusterManagerState remote.RemoteClusterManagerStateManta, tritonAccount, tritonKeyPath, tritonKeyID, tritonURL, mantaURL string) error {
	baseConfig, err := getBaseClusterTerraformConfig(tritonRancherKubernetesTerraformModulePath)
	if err != nil {
		return err
	}

	cfg := tritonClusterTerraformConfig{
		baseClusterTerraformConfig: baseConfig,

		TritonAccount: tritonAccount,
		TritonKeyPath: tritonKeyPath,
		TritonKeyID:   tritonKeyID,
		TritonURL:     tritonURL,
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
	clusterKey := fmt.Sprintf(tritonClusterKeyFormat, cfg.Name)
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
