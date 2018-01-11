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
	"github.com/spf13/viper"
)

const tritonClusterKeyFormat = "cluster_triton_%s"

// This struct represents the definition of a Terraform .tf file.
// Marshalled into json this struct can be passed directly to Terraform.
type tritonClusterTerraformConfig struct {
	Source string `json:"source"`

	Name string `json:"name"`

	EtcdNodeCount          string `json:"etcd_node_count"`
	OrchestrationNodeCount string `json:"orchestration_node_count"`
	ComputeNodeCount       string `json:"compute_node_count"`

	KubernetesPlaneIsolation string `json:"k8s_plane_isolation"`

	RancherAPIURL    string `json:"rancher_api_url"`
	RancherAccessKey string `json:"rancher_access_key"`
	RancherSecretKey string `json:"rancher_secret_key"`

	TritonAccount string `json:"triton_account"`
	TritonKeyPath string `json:"triton_key_path"`
	TritonKeyID   string `json:"triton_key_id"`
	TritonURL     string `json:"triton_url,omitempty"`

	RancherRegistry         string `json:"rancher_registry,omitempty"`
	RancherRegistryUsername string `json:"rancher_registry_username,omitempty"`
	RancherRegistryPassword string `json:"rancher_registry_password,omitempty"`

	KubernetesRegistry         string `json:"k8s_registry,omitempty"`
	KubernetesRegistryUsername string `json:"k8s_registry_username,omitempty"`
	KubernetesRegistryPassword string `json:"k8s_registry_password,omitempty"`
}

func newTritonCluster(selectedClusterManager string, remoteClusterManagerState remote.RemoteClusterManagerStateManta, tritonAccount, tritonKeyPath, tritonKeyID, tritonURL, mantaURL string) error {
	cfg := tritonClusterTerraformConfig{}

	cfg.TritonAccount = tritonAccount
	cfg.TritonKeyPath = tritonKeyPath
	cfg.TritonKeyID = tritonKeyID
	cfg.TritonURL = tritonURL

	baseSource := "github.com/joyent/triton-kubernetes"
	if viper.IsSet("source_url") {
		baseSource = viper.GetString("source_url")
	}

	cfg.Source = fmt.Sprintf("%s//terraform/modules/triton-rancher-k8s", baseSource)

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
