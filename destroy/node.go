package destroy

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/joyent/triton-kubernetes/backend"
	"github.com/joyent/triton-kubernetes/shell"

	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

func DeleteNode(remoteBackend backend.Backend) error {

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
	if viper.IsSet("hostname") {
		nodeHostname := viper.GetString("hostname")
		nodeKey, ok := nodes[nodeHostname]
		if !ok {
			return fmt.Errorf("A node named '%s', does not exist.", nodeHostname)
		}

		selectedNodeKey = nodeKey
	} else {
		nodeNames := make([]string, 0, len(nodes))
		for name := range nodes {
			nodeNames = append(nodeNames, name)
		}
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
		selectedNodeKey = nodes[value]
	}

	// TODO: Prompt confirmation to delete node?

	// Create a temporary directory
	tempDir, err := ioutil.TempDir("", "triton-kubernetes-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	// Save the terraform config to the temporary directory
	jsonPath := fmt.Sprintf("%s/%s", tempDir, "main.tf.json")
	err = ioutil.WriteFile(jsonPath, state.Bytes(), 0644)
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

	// Run terraform destroy
	targetArg := fmt.Sprintf("-target=module.%s", selectedNodeKey)
	err = shell.RunShellCommand(&shellOptions, "terraform", "destroy", "-force", targetArg)
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
