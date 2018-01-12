package create

import (
	"fmt"
	"strings"

	"github.com/joyent/triton-kubernetes/remote"
	"github.com/joyent/triton-kubernetes/util"

	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

func NewCluster() error {
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

	// Ask user what cloud provider the new cluster should be created in
	selectedCloudProvider := ""
	if viper.IsSet("cluster_cloud_provider") {
		selectedCloudProvider = viper.GetString("cluster_cloud_provider")
	} else {
		prompt := promptui.Select{
			Label: "Create Cluster in which Cloud Provider",
			Items: []string{"Triton", "AWS", "GCP", "Azure"},
		}

		_, value, err := prompt.Run()
		if err != nil {
			return err
		}

		selectedCloudProvider = strings.ToLower(value)
	}

	switch selectedCloudProvider {
	case "triton":
		// We pass the same Triton credentials used to get the cluster manager state to create the cluster.
		err = newTritonCluster(selectedClusterManager, remoteClusterManagerState, tritonAccount, tritonKeyPath, tritonKeyID, tritonURL, mantaURL)
	case "aws":
		err = newAWSCluster(selectedClusterManager, remoteClusterManagerState, tritonAccount, tritonKeyPath, tritonKeyID, tritonURL, mantaURL)
	case "gcp":
		err = nil
	case "azure":
		err = newAzureCluster(selectedClusterManager, remoteClusterManagerState)
	default:
		return fmt.Errorf("Unsupported cloud provider '%s', cannot create cluster", selectedCloudProvider)
	}

	if err != nil {
		return err
	}

	// TODO: Ask user if they'd like to create nodes in their new cluster

	return nil
}
