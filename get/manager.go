package get

import (
	"errors"
	"fmt"

	"github.com/mesoform/triton-kubernetes/backend"
	"github.com/mesoform/triton-kubernetes/shell"

	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

func GetManager(remoteBackend backend.Backend) error {
	nonInteractiveMode := viper.GetBool("non-interactive")
	clusterManagers, err := remoteBackend.States()
	if err != nil {
		return err
	}

	if len(clusterManagers) == 0 {
		return fmt.Errorf("No cluster managers.")
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

	err = shell.RunTerraformOutputWithState(state, "cluster-manager")
	if err != nil {
		return err
	}

	return nil
}
