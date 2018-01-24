package create

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/joyent/triton-kubernetes/backend"
	"github.com/joyent/triton-kubernetes/state"

	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

type baseNodeTerraformConfig struct {
	Source string `json:"source"`

	Hostname  string `json:"hostname"`
	NodeCount int    `json:"-"`

	RancherAPIURL        string                  `json:"rancher_api_url"`
	RancherAccessKey     string                  `json:"rancher_access_key"`
	RancherSecretKey     string                  `json:"rancher_secret_key"`
	RancherEnvironmentID string                  `json:"rancher_environment_id"`
	RancherHostLabels    rancherHostLabelsConfig `json:"rancher_host_labels"`

	RancherRegistry         string `json:"rancher_registry,omitempty"`
	RancherRegistryUsername string `json:"rancher_registry_username,omitempty"`
	RancherRegistryPassword string `json:"rancher_registry_password,omitempty"`
}

type rancherHostLabelsConfig struct {
	Orchestration string `json:"orchestration,omitempty"`
	Etcd          string `json:"etcd,omitempty"`
	Compute       string `json:"compute,omitempty"`
}

func NewNode(remoteBackend backend.Backend) error {
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

	state, err := remoteBackend.State(selectedClusterManager)
	if err != nil {
		return err
	}

	// Get existing clusters
	clusters, err := state.Clusters()
	if err != nil {
		return err
	}

	selectedClusterKey := ""
	if viper.IsSet("cluster_name") {
		clusterName := viper.GetString("cluster_name")
		clusterKey, ok := clusters[clusterName]
		if !ok {
			return fmt.Errorf("A cluster named '%s', does not exist.", clusterName)
		}

		selectedClusterKey = clusterKey
	} else {
		clusterNames := make([]string, 0, len(clusters))
		for name := range clusters {
			clusterNames = append(clusterNames, name)
		}
		sort.Strings(clusterNames)
		prompt := promptui.Select{
			Label: "Cluster to deploy node to",
			Items: clusterNames,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}?",
				Active:   fmt.Sprintf("%s {{ . | underline }}", promptui.IconSelect),
				Inactive: " {{ . }}",
				Selected: fmt.Sprintf(`{{ "%s" | green }} {{ "Cluster:" | bold}} {{ . }}`, promptui.IconGood),
			},
		}

		_, value, err := prompt.Run()
		if err != nil {
			return err
		}
		selectedClusterKey = clusters[value]
	}

	err = newNode(selectedClusterManager, selectedClusterKey, remoteBackend, state)
	if err != nil {
		return err
	}

	return nil
}

// Actually creates the new node
func newNode(selectedClusterManager, selectedClusterKey string, remoteBackend backend.Backend, state state.State) error {
	// Determine which cloud the selected cluster is in and call the appropriate newNode func
	parts := strings.Split(selectedClusterKey, "_")
	if len(parts) < 3 {
		// clusterKey is `cluster_{provider}_{hostname}`
		return fmt.Errorf("Could not determine cloud provider for cluster '%s'", selectedClusterKey)
	}

	var err error
	switch parts[1] {
	case "triton":
		err = newTritonNode(selectedClusterManager, selectedClusterKey, remoteBackend, state)
	case "aws":
		err = newAWSNode(selectedClusterManager, selectedClusterKey, remoteBackend, state)
	case "gcp":
		err = newGCPNode(selectedClusterManager, selectedClusterKey, remoteBackend, state)
	case "azure":
		err = newAzureNode(selectedClusterManager, selectedClusterKey, remoteBackend, state)
	default:
		return fmt.Errorf("Unsupported cloud provider '%s', cannot create node", parts[0])
	}
	if err != nil {
		return err
	}
	return nil
}

func getBaseNodeTerraformConfig(terraformModulePath, selectedCluster string, state state.State) (baseNodeTerraformConfig, error) {
	cfg := baseNodeTerraformConfig{
		RancherAPIURL:        "http://${element(module.cluster-manager.masters, 0)}:8080",
		RancherEnvironmentID: fmt.Sprintf("${module.%s.rancher_environment_id}", selectedCluster),

		// Grab registry variables from cluster config
		RancherRegistry:         state.Get(fmt.Sprintf("module.%s.rancher_registry", selectedCluster)),
		RancherRegistryUsername: state.Get(fmt.Sprintf("module.%s.rancher_registry_username", selectedCluster)),
		RancherRegistryPassword: state.Get(fmt.Sprintf("module.%s.rancher_registry_password", selectedCluster)),
	}

	baseSource := defaultSourceURL
	if viper.IsSet("source_url") {
		baseSource = viper.GetString("source_url")
	}

	baseSourceRef := defaultSourceRef
	if viper.IsSet("source_ref") {
		baseSourceRef = viper.GetString("source_ref")
	}

	cfg.Source = fmt.Sprintf("%s//%s?ref=%s", baseSource, terraformModulePath, baseSourceRef)

	// Rancher Host Label
	selectedHostLabel := ""
	hostLabelOptions := []string{
		"compute",
		"etcd",
		"orchestration",
	}
	if viper.IsSet("rancher_host_label") {
		selectedHostLabel = viper.GetString("rancher_host_label")
	} else {
		prompt := promptui.Select{
			Label: "Which type of node?",
			Items: hostLabelOptions,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}?",
				Active:   fmt.Sprintf("%s {{ . | underline }}", promptui.IconSelect),
				Inactive: "  {{ . }}",
				Selected: fmt.Sprintf(`{{ "%s" | green }} {{ "Host Type:" | bold}} {{ . }}`, promptui.IconGood),
			},
		}

		i, _, err := prompt.Run()
		if err != nil {
			return baseNodeTerraformConfig{}, err
		}

		selectedHostLabel = hostLabelOptions[i]
	}

	switch selectedHostLabel {
	case "compute":
		cfg.RancherHostLabels.Compute = "true"
	case "etcd":
		cfg.RancherHostLabels.Etcd = "true"
	case "orchestration":
		cfg.RancherHostLabels.Orchestration = "true"
	default:
		return baseNodeTerraformConfig{}, fmt.Errorf("Invalid rancher_host_label '%s', must be 'compute', 'etcd' or 'orchestration'", selectedHostLabel)
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
		}

		result, err := prompt.Run()
		if err != nil {
			return baseNodeTerraformConfig{}, err
		}

		countInput = result
	}

	// Verifying node count
	nodeCount, err := strconv.Atoi(countInput)
	if err != nil {
		return baseNodeTerraformConfig{}, fmt.Errorf("node_count must be a valid number. Found '%s'.", countInput)
	}
	if nodeCount <= 0 {
		return baseNodeTerraformConfig{}, fmt.Errorf("node_count must be greater than 0. Found '%d'.", nodeCount)
	}

	cfg.NodeCount = nodeCount

	// hostname
	if viper.IsSet("hostname") {
		cfg.Hostname = viper.GetString("hostname")
	} else {
		prompt := promptui.Prompt{
			Label: "Hostname",
		}

		result, err := prompt.Run()
		if err != nil {
			return baseNodeTerraformConfig{}, err
		}
		cfg.Hostname = result
	}

	if cfg.Hostname == "" {
		return baseNodeTerraformConfig{}, errors.New("Invalid Hostname")
	}

	return cfg, nil
}

// Returns the hostnames that should be used when adding new nodes. Prevents naming collisions.
func getNewHostnames(existingNames []string, nodeName string, nodesToAdd int) []string {
	if nodesToAdd < 1 {
		return []string{}
	}

	// Find the number at which the series of hostnames should start.
	startNum := 1
	targetPrefix := nodeName + "-"
	for _, existingName := range existingNames {
		if !strings.HasPrefix(existingName, targetPrefix) {
			continue
		}

		suffix := existingName[len(targetPrefix):]
		numSuffix, err := strconv.Atoi(suffix)
		if err != nil {
			continue
		}
		if numSuffix >= startNum {
			startNum = numSuffix + 1
		}
	}

	// Build the list of hostnames
	result := []string{}
	for i := 0; i < nodesToAdd; i++ {
		result = append(result, fmt.Sprintf("%s-%d", nodeName, startNum+i))
	}

	return result
}
