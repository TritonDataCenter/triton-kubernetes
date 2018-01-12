package destroy

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/joyent/triton-kubernetes/remote"
	"github.com/joyent/triton-kubernetes/shell"
	"github.com/joyent/triton-kubernetes/util"

	"github.com/Jeffail/gabs"
	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

func DeleteNode() error {
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
		return fmt.Errorf("No cluster managers, please create a cluster manager before creating a kubernetes node.")
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
			Label: "Cluster to delete",
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

	nodeOptions, err := util.GetNodeOptions(clusterManagerTerraformConfig, selectedClusterKey)
	if err != nil {
		return err
	}

	selectedNodeKey := ""
	if viper.IsSet("hostname") {
		nodeHostname := viper.GetString("hostname")
		for _, option := range nodeOptions {
			if nodeHostname == option.Hostname {
				selectedNodeKey = option.NodeKey
				break
			}
		}

		if selectedNodeKey == "" {
			return fmt.Errorf("A node named '%s', does not exist.", nodeHostname)
		}
	} else {
		prompt := promptui.Select{
			Label: "Node to delete",
			Items: nodeOptions,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}?",
				Active:   fmt.Sprintf("%s {{ .Hostname | underline }}", promptui.IconSelect),
				Inactive: " {{ .Hostname }}",
				Selected: fmt.Sprintf(`{{ "%s" | green }} {{ "Node:" | bold}} {{ .Hostname }}`, promptui.IconGood),
			},
		}

		i, _, err := prompt.Run()
		if err != nil {
			return err
		}
		selectedNodeKey = nodeOptions[i].NodeKey
	}

	// TODO: Prompt confirmation to delete node?

	// Create a temporary directory
	tempDir, err := ioutil.TempDir("", "triton-kubernetes-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	// Save the terraform config to the temporary directory
	jsonBytes := []byte(clusterManagerTerraformConfig.StringIndent("", "\t"))
	jsonPath := fmt.Sprintf("%s/%s", tempDir, "main.tf.json")
	err = ioutil.WriteFile(jsonPath, jsonBytes, 0644)
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
	err = clusterManagerTerraformConfig.Delete("module", selectedNodeKey)
	if err != nil {
		return err
	}

	// After terraform succeeds, commit state
	err = remoteClusterManagerState.CommitTerraformConfig(selectedClusterManager, clusterManagerTerraformConfig.BytesIndent("", "\t"))
	if err != nil {
		return err
	}

	return nil
}
