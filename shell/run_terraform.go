package shell

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/joyent/triton-kubernetes/state"

	getter "github.com/hashicorp/go-getter"
)

type thirdPartyPlugin struct {
	BaseURL    string
	SHA256Sums map[string]string
}

var thirdPartyPlugins = []thirdPartyPlugin{
	thirdPartyPlugin{
		BaseURL: "https://github.com/yamamoto-febc/terraform-provider-rke/releases/download/0.4.0/terraform-provider-rke_0.4.0_%s-%s.zip",
		SHA256Sums: map[string]string{
			"terraform-provider-rke_0.4.0_darwin-386.zip":    "b8b3085b06307619a98b83dd9901f8504169e83948843dc1af79aa058dfc03ee",
			"terraform-provider-rke_0.4.0_darwin-amd64.zip":  "c6abfdee7c4db0bd8d068172c9c97beee4f382e1578ad0588f0347a725fd5562",
			"terraform-provider-rke_0.4.0_linux-386.zip":     "4b8b3d76efd4e64813de820342c79cf60bffe6726d6a15a2306d636f5ae1dbfe",
			"terraform-provider-rke_0.4.0_linux-amd64.zip":   "3c9bc09b389d00a4e130e4674c5aae6a5284417dbbc4de3c7e14a9a20eabe51c",
			"terraform-provider-rke_0.4.0_windows-386.zip":   "b8a4594edadf9489b331939453b61cdbd383f380645f50f1411b6796155b7f73",
			"terraform-provider-rke_0.4.0_windows-amd64.zip": "1975fa24b2aa5830d884649a34adee6136e213cf5edd021513232d77eb50820c",
		},
	},
}

// A simplistic third party plugin installer.
// This func will install terraform plugins regardless of whether they're needed for this particular execution.
// Terraform doesn't currently support automatically installing third party plugins, issue being tracked here
// https://github.com/hashicorp/terraform/issues/17154
func installThirdPartyProviders(workingDirectory string) error {
	fmt.Println("Installing third party plugins...")

	pluginInstallDirectoryPath := filepath.Join(workingDirectory, "terraform.d", "plugins", fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH))

	// Download/Verify/Unzip plugins
	for _, plugin := range thirdPartyPlugins {
		pluginURL := fmt.Sprintf(plugin.BaseURL, runtime.GOOS, runtime.GOARCH)
		_, pluginFileName := path.Split(pluginURL)

		pluginURL = fmt.Sprintf("%s?checksum=sha256:%s", pluginURL, plugin.SHA256Sums[pluginFileName])

		// Download/Verify/Unzip
		err := getter.Get(pluginInstallDirectoryPath, pluginURL)
		if err != nil {
			return err
		}
	}

	fmt.Println("Finished installing third party plugins...")

	return nil
}

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

	// Install third party providers
	err = installThirdPartyProviders(tempDir)
	if err != nil {
		return err
	}

	// Run terraform init
	err = runShellCommand(&shellOptions, "terraform", "init", "-force-copy")
	if err != nil {
		return err
	}

	// Run terraform apply
	err = runShellCommand(&shellOptions, "terraform", "apply", "-auto-approve")
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

	// Install third party providers
	err = installThirdPartyProviders(tempDir)
	if err != nil {
		return err
	}

	// Run terraform init
	err = runShellCommand(&shellOptions, "terraform", "init", "-force-copy")
	if err != nil {
		return err
	}

	// Run terraform destroy
	allArgs := append([]string{"destroy", "-force"}, args...)
	err = runShellCommand(&shellOptions, "terraform", allArgs...)
	if err != nil {
		return err
	}

	return nil
}

func RunTerraformOutputWithState(state state.State, moduleName string) error {
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

	// Install third party providers
	err = installThirdPartyProviders(tempDir)
	if err != nil {
		return err
	}

	// Run terraform init
	err = runShellCommand(&shellOptions, "terraform", "init", "-force-copy")
	if err != nil {
		return err
	}

	// Run terraform output
	err = runShellCommand(&shellOptions, "terraform", "output", "-module", moduleName)
	if err != nil {
		return err
	}

	return nil
}
