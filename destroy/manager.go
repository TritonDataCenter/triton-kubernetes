package destroy

import (
	"errors"
	"fmt"
	"sort"

	"github.com/mesoform/triton-kubernetes/backend"
	"github.com/mesoform/triton-kubernetes/shell"
	"github.com/mesoform/triton-kubernetes/util"

	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

func DeleteManager(remoteBackend backend.Backend) error {
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
		sort.Strings(clusterManagers)
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

	if !nonInteractiveMode {
		// Confirmation
		label := fmt.Sprintf("Are you sure you want to destroy %q", selectedClusterManager)
		selected := fmt.Sprintf("Destroy %q", selectedClusterManager)
		confirmed, err := util.PromptForConfirmation(label, selected)
		if err != nil {
			return err
		}
		if !confirmed {
			fmt.Println("Destroy manager canceled.")
			return nil
		}
	}

	// Run Terraform destroy
	err = shell.RunTerraformDestroyWithState(state, []string{})
	if err != nil {
		return err
	}

	// After terraform succeeds, delete remote state
	err = remoteBackend.DeleteState(selectedClusterManager)
	if err != nil {
		return err
	}

	return nil
}
