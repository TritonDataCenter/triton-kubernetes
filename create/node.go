package create

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/joyent/triton-kubernetes/backend"
	"github.com/joyent/triton-kubernetes/shell"
	"github.com/joyent/triton-kubernetes/state"
	"github.com/joyent/triton-kubernetes/util"

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
	nonInteractiveMode := viper.GetBool("non-interactive")
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
	} else if nonInteractiveMode {
		return errors.New("cluster_manager must be specified")
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

	// Get existing clusters
	clusters, err := currentState.Clusters()
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
	} else if nonInteractiveMode {
		return errors.New("cluster_name must be specified")
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

	_, err = newNode(selectedClusterManager, selectedClusterKey, remoteBackend, currentState)
	if err != nil {
		return err
	}

	// Confirmation Prompt
	if !nonInteractiveMode {
		label := "Proceed with the node creation"
		selected := "Proceed"
		confirmed, err := util.PromptForConfirmation(label, selected)
		if err != nil {
			return err
		}
		if !confirmed {
			fmt.Println("Node creation canceled")
			return nil
		}
	}

	// Get the new state and run terraform apply
	err = shell.RunTerraformApplyWithState(currentState)
	if err != nil {
		return err
	}

	// After terraform succeeds, commit state
	err = remoteBackend.PersistState(currentState)
	if err != nil {
		return err
	}

	return nil
}

func newNode(selectedClusterManager, selectedClusterKey string, remoteBackend backend.Backend, currentState state.State) ([]string, error) {
	// Determine which cloud the selected cluster is in and call the appropriate newNode func
	parts := strings.Split(selectedClusterKey, "_")
	if len(parts) < 3 {
		// clusterKey is `cluster_{provider}_{hostname}`
		return []string{}, fmt.Errorf("Could not determine cloud provider for cluster '%s'", selectedClusterKey)
	}

	switch parts[1] {
	case "triton":
		return newTritonNode(selectedClusterManager, selectedClusterKey, remoteBackend, currentState)
	case "aws":
		return newAWSNode(selectedClusterManager, selectedClusterKey, remoteBackend, currentState)
	case "gcp":
		return newGCPNode(selectedClusterManager, selectedClusterKey, remoteBackend, currentState)
	case "azure":
		return newAzureNode(selectedClusterManager, selectedClusterKey, remoteBackend, currentState)
	case "baremetal":
		return newBareMetalNode(selectedClusterManager, selectedClusterKey, remoteBackend, currentState)
	default:
		return []string{}, fmt.Errorf("Unsupported cloud provider '%s', cannot create node", parts[1])
	}
}

func getBaseNodeTerraformConfig(terraformModulePath, selectedCluster string, currentState state.State) (baseNodeTerraformConfig, error) {
	cfg := baseNodeTerraformConfig{
		RancherAPIURL:    "${module.cluster-manager.rancher_url}",
		RancherAccessKey: "${module.cluster-manager.rancher_access_key}",
		RancherSecretKey: "${module.cluster-manager.rancher_secret_key}",

		RancherEnvironmentID: fmt.Sprintf("${module.%s.rancher_environment_id}", selectedCluster),

		// Grab registry variables from cluster config
		RancherRegistry:         currentState.Get(fmt.Sprintf("module.%s.rancher_registry", selectedCluster)),
		RancherRegistryUsername: currentState.Get(fmt.Sprintf("module.%s.rancher_registry_username", selectedCluster)),
		RancherRegistryPassword: currentState.Get(fmt.Sprintf("module.%s.rancher_registry_password", selectedCluster)),
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
	} else if cfg.RancherHostLabels.Compute == "true" {
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
			return baseNodeTerraformConfig{}, err
		}
		countInput = result
	} else {
		nodeCountOptions := []string{"1", "3", "5", "7"}

		prompt := promptui.Select{
			Label: "Number of nodes to create",
			Items: nodeCountOptions,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}?",
				Active:   fmt.Sprintf("%s {{ . | underline }}", promptui.IconSelect),
				Inactive: "  {{.}}",
				Selected: "  Number of nodes to create? {{.}}",
			},
		}

		i, _, err := prompt.Run()
		if err != nil {
			return baseNodeTerraformConfig{}, err
		}

		countInput = nodeCountOptions[i]
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
			Label: "Hostname prefix",
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
