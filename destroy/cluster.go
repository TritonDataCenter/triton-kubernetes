package destroy

import (
	"errors"
	"fmt"
	"sort"

	"github.com/joyent/triton-kubernetes/backend"
	"github.com/joyent/triton-kubernetes/shell"
	"github.com/joyent/triton-kubernetes/util"

	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

func DeleteCluster(remoteBackend backend.Backend, silentMode bool) error {
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
	} else if silentMode {
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
	clusterName := ""
	if viper.IsSet("cluster_name") {
		clusterName = viper.GetString("cluster_name")
		clusterKey, ok := clusters[clusterName]
		if !ok {
			return fmt.Errorf("A cluster named '%s', does not exist.", clusterName)
		}

		selectedClusterKey = clusterKey
	} else if silentMode {
		return errors.New("cluster_name must be specified")
	} else {
		clusterNames := make([]string, 0, len(clusters))
		for name := range clusters {
			clusterNames = append(clusterNames, name)
		}
		sort.Strings(clusterNames)
		prompt := promptui.Select{
			Label: "Cluster to delete",
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
		clusterName = value
		selectedClusterKey = clusters[value]
	}

	// Confirmation
	if !silentMode {
		label := fmt.Sprintf("Are you sure you want to destroy %q", clusterName)
		selected := fmt.Sprintf("Destroy %q", clusterName)
		confirmed, err := util.PromptForConfirmation(label, selected)
		if err != nil {
			return err
		}
		if !confirmed {
			fmt.Println("Destroy cluster canceled.")
			return nil
		}
	}

	nodes, err := state.Nodes(selectedClusterKey)
	if err != nil {
		return err
	}

	args := []string{
		fmt.Sprintf("-target=module.%s", selectedClusterKey),
	}

	// Delete all nodes in the selected cluster
	for _, node := range nodes {
		args = append(args, fmt.Sprintf("-target=module.%s", node))
	}

	// Run terraform destroy
	err = shell.RunTerraformDestroyWithState(state, args)
	if err != nil {
		return err
	}

	// Remove cluster from terraform config
	err = state.Delete(fmt.Sprintf("module.%s", selectedClusterKey))
	if err != nil {
		return err
	}

	// Remove all nodes associated to this cluster from terraform config
	for _, node := range nodes {
		err = state.Delete(fmt.Sprintf("module.%s", node))
		if err != nil {
			return err
		}
	}

	// After terraform succeeds, commit state
	err = remoteBackend.PersistState(state)
	if err != nil {
		return err
	}

	return nil
}
