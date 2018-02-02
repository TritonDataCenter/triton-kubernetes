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

func DeleteNode(remoteBackend backend.Backend) error {
	silentMode := viper.GetBool("silent")
	clusterManagers, err := remoteBackend.States()
	if err != nil {
		return err
	}

	if len(clusterManagers) == 0 {
		return fmt.Errorf("No cluster managers, please create a cluster manager before creating a kubernetes node.")
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
	if viper.IsSet("cluster_name") {
		clusterName := viper.GetString("cluster_name")
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
		selectedClusterKey = clusters[value]
	}

	// Get existing nodes
	nodes, err := state.Nodes(selectedClusterKey)
	if err != nil {
		return err
	}

	selectedNodeKey := ""
	nodeHostname := ""
	if viper.IsSet("hostname") {
		nodeHostname = viper.GetString("hostname")
		nodeKey, ok := nodes[nodeHostname]
		if !ok {
			return fmt.Errorf("A node named '%s', does not exist.", nodeHostname)
		}

		selectedNodeKey = nodeKey
	} else if silentMode {
		return errors.New("hostname must be specified")
	} else {
		nodeNames := make([]string, 0, len(nodes))
		for name := range nodes {
			nodeNames = append(nodeNames, name)
		}
		sort.Strings(nodeNames)
		prompt := promptui.Select{
			Label: "Node to delete",
			Items: nodeNames,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}?",
				Active:   fmt.Sprintf("%s {{ . | underline }}", promptui.IconSelect),
				Inactive: " {{ . }}",
				Selected: fmt.Sprintf(`{{ "%s" | green }} {{ "Node:" | bold}} {{ . }}`, promptui.IconGood),
			},
		}

		_, value, err := prompt.Run()
		if err != nil {
			return err
		}
		nodeHostname = value
		selectedNodeKey = nodes[value]
	}

	// Confirmation
	label := fmt.Sprintf("Are you sure you want to destroy %q", nodeHostname)
	selected := fmt.Sprintf("Destroy %q", nodeHostname)
	confirmed, err := util.PromptForConfirmation(label, selected)
	if err != nil {
		return err
	}
	if !confirmed {
		fmt.Println("Destroy node canceled.")
		return nil
	}

	// Run terraform destroy
	targetArg := fmt.Sprintf("-target=module.%s", selectedNodeKey)
	err = shell.RunTerraformDestroyWithState(state, []string{targetArg})
	if err != nil {
		return err
	}

	// Remove node from terraform config
	err = state.Delete(fmt.Sprintf("module.%s", selectedNodeKey))
	if err != nil {
		return err
	}

	// After terraform succeeds, commit state
	err = remoteBackend.PersistState(state)
	if err != nil {
		return err
	}

	return nil
}
