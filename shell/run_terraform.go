package shell

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/joyent/triton-kubernetes/state"
)

const terraformCmd = "/Users/chrisguevara/homdna-service/bin/terraform"

func RunTerraformApplyWithState(state state.State) error {
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
	shellOptions := ShellOptions{
		WorkingDir: tempDir,
	}

	// Run terraform init
	err = RunShellCommand(&shellOptions, terraformCmd, "init", "-force-copy")
	if err != nil {
		return err
	}

	// Run terraform apply
	err = RunShellCommand(&shellOptions, terraformCmd, "apply", "-auto-approve")
	if err != nil {
		return err
	}

	return nil
}

func RunTerraformDestroyWithState(currentState state.State, args []string) error {
	// Create a temporary directory
	tempDir, err := ioutil.TempDir("", "triton-kubernetes-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	// Save the terraform config to the temporary directory
	jsonPath := fmt.Sprintf("%s/%s", tempDir, "main.tf.json")
	err = ioutil.WriteFile(jsonPath, currentState.Bytes(), 0644)
	if err != nil {
		return err
	}

	// Use temporary directory as working directory
	shellOptions := ShellOptions{
		WorkingDir: tempDir,
	}

	// Run terraform init
	err = RunShellCommand(&shellOptions, terraformCmd, "init", "-force-copy")
	if err != nil {
		return err
	}

	// Run terraform destroy
	allArgs := append([]string{"destroy", "-force"}, args...)
	err = RunShellCommand(&shellOptions, terraformCmd, allArgs...)
	if err != nil {
		return err
	}

	return nil
}
