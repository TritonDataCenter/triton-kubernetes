package get

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

func GetManager() error {
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
		return fmt.Errorf("No cluster managers.")
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

	// Load current cluster manager config
	clusterManagerTerraformConfigBytes, err := remoteClusterManagerState.GetTerraformConfig(selectedClusterManager)
	if err != nil {
		return err
	}

	clusterManagerTerraformConfig, err := gabs.ParseJSON(clusterManagerTerraformConfigBytes)
	if err != nil {
		return err
	}

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

	// Run terraform output
	err = shell.RunShellCommand(&shellOptions, "terraform", "output", "-module", "cluster-manager")
	if err != nil {
		return err
	}

	return nil
}
