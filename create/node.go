package create

import (
	"fmt"
	"strings"

	"github.com/joyent/triton-kubernetes/remote"
	"github.com/joyent/triton-kubernetes/util"

	"github.com/Jeffail/gabs"
	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

type rancherHostLabelsConfig struct {
	Orchestration string `json:"orchestration,omitempty"`
	Etcd          string `json:"etcd,omitempty"`
	Compute       string `json:"compute,omitempty"`
}

func NewNode() error {
	tritonAccount, tritonKeyPath, tritonKeyID, tritonURL, mantaURL, err := util.GetTritonAccountVariables()
	if err != nil {
		return err
	}

	remoteClusterManagerState, err := remote.NewRemoteClusterManagerStateManta(tritonAccount, tritonKeyPath, tritonKeyID, tritonURL, mantaURL)
	if err != nil {
		return err
	}

	clusterManagers, err := remoteClusterManagerState.List()
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

	// Get existing clusters
	clusterManagerTerraformConfigBytes, err := remoteClusterManagerState.GetTerraformConfig(selectedClusterManager)
	if err != nil {
		return err
	}

	clusterManagerTerraformConfig, err := gabs.ParseJSON(clusterManagerTerraformConfigBytes)
	if err != nil {
		return err
	}

	clusterOptions, err := util.GetClusterOptions(clusterManagerTerraformConfig)
	if err != nil {
		return err
	}

	selectedClusterKey := ""
	if viper.IsSet("cluster_name") {
		clusterName := viper.GetString("cluster_name")
		for _, option := range clusterOptions {
			if clusterName == option.ClusterName {
				selectedClusterKey = option.ClusterKey
				break
			}
		}

		if selectedClusterKey == "" {
			return fmt.Errorf("A cluster named '%s', does not exist.", clusterName)
		}
	} else {
		prompt := promptui.Select{
			Label: "Cluster to create node in",
			Items: clusterOptions,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}?",
				Active:   fmt.Sprintf("%s {{ .ClusterName | underline }}", promptui.IconSelect),
				Inactive: " {{ .ClusterName }}",
				Selected: fmt.Sprintf(`{{ "%s" | green }} {{ "Cluster:" | bold}} {{ .ClusterName }}`, promptui.IconGood),
			},
		}

		i, _, err := prompt.Run()
		if err != nil {
			return err
		}
		selectedClusterKey = clusterOptions[i].ClusterKey
	}

	// Determine which cloud the selected cluster is in and call the appropriate newNode func
	parts := strings.Split(selectedClusterKey, "_")
	if len(parts) < 3 {
		// clusterKey is `cluster_{provider}_{hostname}`
		return fmt.Errorf("Could not determine cloud provider for cluster '%s'", selectedClusterKey)
	}

	switch parts[1] {
	case "triton":
		err = newTritonNode(selectedClusterManager, selectedClusterKey, remoteClusterManagerState, tritonAccount, tritonKeyPath, tritonKeyID, tritonURL, mantaURL)
	case "aws":
		err = nil
	case "gcp":
		err = nil
	case "azure":
		err = nil
	default:
		return fmt.Errorf("Unsupported cloud provider '%s', cannot create node", parts[0])
	}

	if err != nil {
		return err
	}

	return nil
}
