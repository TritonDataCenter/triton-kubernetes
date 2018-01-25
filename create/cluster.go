package create

import (
	"errors"
	"fmt"
	"strings"

	"github.com/joyent/triton-kubernetes/shell"
	"github.com/joyent/triton-kubernetes/state"

	"github.com/joyent/triton-kubernetes/backend"

	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

const (
	defaultSourceURL = "github.com/joyent/triton-kubernetes"
	defaultSourceRef = "master"
)

type baseClusterTerraformConfig struct {
	Source string `json:"source"`

	Name string `json:"name"`

	EtcdNodeCount          string `json:"etcd_node_count"`
	OrchestrationNodeCount string `json:"orchestration_node_count"`
	ComputeNodeCount       string `json:"compute_node_count"`

	KubernetesPlaneIsolation string `json:"k8s_plane_isolation"`

	RancherAPIURL    string `json:"rancher_api_url"`
	RancherAccessKey string `json:"rancher_access_key"`
	RancherSecretKey string `json:"rancher_secret_key"`

	RancherRegistry         string `json:"rancher_registry,omitempty"`
	RancherRegistryUsername string `json:"rancher_registry_username,omitempty"`
	RancherRegistryPassword string `json:"rancher_registry_password,omitempty"`

	KubernetesRegistry         string `json:"k8s_registry,omitempty"`
	KubernetesRegistryUsername string `json:"k8s_registry_username,omitempty"`
	KubernetesRegistryPassword string `json:"k8s_registry_password,omitempty"`
}

func NewCluster(remoteBackend backend.Backend) error {
	clusterManagers, err := remoteBackend.States()
	if err != nil {
		return err
	}

	if len(clusterManagers) == 0 {
		return fmt.Errorf("No cluster managers, please create a cluster manager before creating a kubernetes cluster.")
	}

	selectedClusterManager := ""
	if viper.IsSet("cluster_manager") {
		selectedClusterManager = viper.GetString("cluster_manager")
	} else {
		prompt := promptui.Select{
			Label: "Cluster Manager",
			Items: clusterManagers,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}?",
				Active:   fmt.Sprintf(`%s {{ . | underline }}`, promptui.IconSelect),
				Inactive: `  {{ . }}`,
				Selected: fmt.Sprintf(`{{ "%s" | green }} {{ "Cluster Manager:" | bold}} {{ . }}`, promptui.IconGood),
			},
		}

		_, value, err := prompt.Run()
		if err != nil {
			return err
		}

		selectedClusterManager = value
	}

	// Verify selected cluster manager exists
	found := false
	for _, clusterManager := range clusterManagers {
		if selectedClusterManager == clusterManager {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("Selected cluster manager '%s' does not exist.", selectedClusterManager)
	}

	currentState, err := remoteBackend.State(selectedClusterManager)
	if err != nil {
		return err
	}

	// Ask user what cloud provider the new cluster should be created in
	selectedCloudProvider := ""
	if viper.IsSet("cluster_cloud_provider") {
		selectedCloudProvider = viper.GetString("cluster_cloud_provider")
	} else {
		prompt := promptui.Select{
			Label: "Create Cluster in which Cloud Provider",
			Items: []string{"Triton", "AWS", "GCP", "Azure"},
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

	var clusterName string
	switch selectedCloudProvider {
	case "triton":
		// We pass the same Triton credentials used to get the cluster manager state to create the cluster.
		clusterName, err = newTritonCluster(remoteBackend, currentState)
	case "aws":
		clusterName, err = newAWSCluster(remoteBackend, currentState)
	case "gcp":
		clusterName, err = newGCPCluster(remoteBackend, currentState)
	case "azure":
		clusterName, err = newAzureCluster(remoteBackend, currentState)
	default:
		return fmt.Errorf("Unsupported cloud provider '%s', cannot create cluster", selectedCloudProvider)
	}

	// TODO: Find a fix - state.Clusters() doesn't return any clusters added via state.Add().
	// However, the new clusters appear in the result of state.Bytes(). The current workaround
	// is to create a new state object that has the same bytes as the previous state object.
	currentState, err = state.New(currentState.Name, currentState.Bytes())
	if err != nil {
		return err
	}

	// Get the new cluster key given the cluster name
	clusterMap, err := currentState.Clusters()
	if err != nil {
		return err
	}
	clusterKey, ok := clusterMap[clusterName]
	if !ok {
		return fmt.Errorf("Couldn't find cluster key for cluster '%s'.\n", clusterName)
	}

	// Ask user if they'd like to create a node for this cluster
	createNodeOptions := []struct {
		Name  string
		Value bool
	}{
		{"Yes", true},
		{"No", false},
	}

	createNodePrompt := promptui.Select{
		Label: "Would you like to create nodes for this cluster",
		Items: createNodeOptions,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}?",
			Active:   fmt.Sprintf("%s {{ .Name | underline }}", promptui.IconSelect),
			Inactive: "  {{.Name}}",
			Selected: "  Create new node? {{.Name}}",
		},
	}

	i, _, err := createNodePrompt.Run()
	if err != nil {
		return err
	}

	shouldCreateNode := createNodeOptions[i].Value
	createNodePrompt.Label = "Would you like to create more nodes for this cluster"

	for shouldCreateNode {
		// Add new nodes to the state
		_, err = newNode(selectedClusterManager, clusterKey, remoteBackend, currentState)
		if err != nil {
			return err
		}

		// Ask if user would like to create more nodes
		i, _, err := createNodePrompt.Run()
		if err != nil {
			return err
		}
		shouldCreateNode = createNodeOptions[i].Value
	}

	// Run terraform apply with state
	err = shell.RunTerraformApplyWithState(currentState)

	// After terraform succeeds, commit state
	err = remoteBackend.PersistState(currentState)
	if err != nil {
		return err
	}

	return nil
}

func getBaseClusterTerraformConfig(terraformModulePath string) (baseClusterTerraformConfig, error) {
	cfg := baseClusterTerraformConfig{
		RancherAPIURL: "http://${element(module.cluster-manager.masters, 0)}:8080",

		// Set node counts to 0 since we manage nodes individually in triton-kubernetes cli
		EtcdNodeCount:          "0",
		OrchestrationNodeCount: "0",
		ComputeNodeCount:       "0",
	}

	baseSource := defaultSourceURL
	if viper.IsSet("source_url") {
		baseSource = viper.GetString("source_url")
	}

	baseSourceRef := defaultSourceRef
	if viper.IsSet("source_ref") {
		baseSourceRef = viper.GetString("source_ref")
	}

	// Module Source location e.g. github.com/joyent/triton-kubernetes//terraform/modules/azure-rancher-k8s?ref=master
	cfg.Source = fmt.Sprintf("%s//%s?ref=%s", baseSource, terraformModulePath, baseSourceRef)

	// Name
	if viper.IsSet("name") {
		cfg.Name = viper.GetString("name")
	} else {
		prompt := promptui.Prompt{
			Label: "Cluster Name",
		}

		result, err := prompt.Run()
		if err != nil {
			return baseClusterTerraformConfig{}, err
		}
		cfg.Name = result
	}

	if cfg.Name == "" {
		return baseClusterTerraformConfig{}, errors.New("Invalid Cluster Name")
	}

	// Kubernetes Plane Isolation
	if viper.IsSet("k8s_plane_isolation") {
		cfg.KubernetesPlaneIsolation = viper.GetString("k8s_plane_isolation")
	} else {
		prompt := promptui.Select{
			Label: "Kubernetes Plane Isolation",
			Items: []string{"required", "none"},
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}?",
				Active:   fmt.Sprintf(`%s {{ . | underline }}`, promptui.IconSelect),
				Inactive: `  {{ . }}`,
				Selected: fmt.Sprintf(`{{ "%s" | green }} {{ "k8s Plane Isolation:" | bold}} {{ . }}`, promptui.IconGood),
			},
		}

		_, value, err := prompt.Run()
		if err != nil {
			return baseClusterTerraformConfig{}, err
		}

		cfg.KubernetesPlaneIsolation = value
	}

	// Verify selected plane isolation is valid
	if cfg.KubernetesPlaneIsolation != "required" && cfg.KubernetesPlaneIsolation != "none" {
		return baseClusterTerraformConfig{}, fmt.Errorf("Invalid k8s_plane_isolation '%s', must be 'required' or 'none'", cfg.KubernetesPlaneIsolation)
	}

	// Rancher Docker Registry
	if viper.IsSet("rancher_registry") {
		cfg.RancherRegistry = viper.GetString("rancher_registry")
	} else {
		prompt := promptui.Prompt{
			Label:   "Rancher Registry",
			Default: "None",
		}

		result, err := prompt.Run()
		if err != nil {
			return baseClusterTerraformConfig{}, err
		}

		if result != "None" {
			cfg.RancherRegistry = result
		}
	}

	// Ask for rancher registry username/password only if rancher registry is given
	if cfg.RancherRegistry != "" {
		// Rancher Registry Username
		if viper.IsSet("rancher_registry_username") {
			cfg.RancherRegistryUsername = viper.GetString("rancher_registry_username")
		} else {
			prompt := promptui.Prompt{
				Label: "Rancher Registry Username",
			}

			result, err := prompt.Run()
			if err != nil {
				return baseClusterTerraformConfig{}, err
			}
			cfg.RancherRegistryUsername = result
		}

		// Rancher Registry Password
		if viper.IsSet("rancher_registry_password") {
			cfg.RancherRegistryPassword = viper.GetString("rancher_registry_password")
		} else {
			prompt := promptui.Prompt{
				Label: "Rancher Registry Password",
				Mask:  '*',
			}

			result, err := prompt.Run()
			if err != nil {
				return baseClusterTerraformConfig{}, err
			}
			cfg.RancherRegistryPassword = result
		}
	}

	// k8s Docker Registry
	if viper.IsSet("k8s_registry") {
		cfg.KubernetesRegistry = viper.GetString("k8s_registry")
	} else {
		prompt := promptui.Prompt{
			Label:   "k8s Registry",
			Default: "None",
		}

		result, err := prompt.Run()
		if err != nil {
			return baseClusterTerraformConfig{}, err
		}

		if result != "None" {
			cfg.KubernetesRegistry = result
		}
	}

	// Ask for k8s registry username/password only if k8s registry is given
	if cfg.KubernetesRegistry != "" {
		// k8s Registry Username
		if viper.IsSet("k8s_registry_username") {
			cfg.KubernetesRegistryUsername = viper.GetString("k8s_registry_username")
		} else {
			prompt := promptui.Prompt{
				Label: "k8s Registry Username",
			}

			result, err := prompt.Run()
			if err != nil {
				return baseClusterTerraformConfig{}, err
			}
			cfg.KubernetesRegistryUsername = result
		}

		// Rancher Registry Password
		if viper.IsSet("k8s_registry_password") {
			cfg.KubernetesRegistryPassword = viper.GetString("k8s_registry_password")
		} else {
			prompt := promptui.Prompt{
				Label: "k8s Registry Password",
				Mask:  '*',
			}

			result, err := prompt.Run()
			if err != nil {
				return baseClusterTerraformConfig{}, err
			}
			cfg.KubernetesRegistryPassword = result
		}
	}

	return cfg, nil
}
