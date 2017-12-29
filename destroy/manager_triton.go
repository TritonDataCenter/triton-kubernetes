package destroy

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	triton "github.com/joyent/triton-go"
	"github.com/joyent/triton-go/authentication"
	"github.com/joyent/triton-go/storage"
	"github.com/joyent/triton-kubernetes/shell"
	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

func DeleteTritonManager() error {
	var tritonAccount string
	var tritonKeyPath string
	var tritonKeyID string
	var tritonURL string
	var targetManager string

	// Triton Account
	if viper.IsSet("triton_account") {
		tritonAccount = viper.GetString("triton_account")
	} else {
		prompt := promptui.Prompt{
			Label: "Triton Account Name (usually your email)",
			Validate: func(input string) error {
				if len(input) == 0 {
					return errors.New("Invalid Triton Account")
				}
				return nil
			},
		}

		result, err := prompt.Run()
		if err != nil {
			return err
		}
		tritonAccount = result
	}

	// Triton Key Path
	if viper.IsSet("triton_key_path") {
		tritonKeyPath = viper.GetString("triton_key_path")
	} else {
		prompt := promptui.Prompt{
			Label: "Triton Key Path",
			Validate: func(input string) error {
				_, err := os.Stat(input)
				if err != nil {
					if os.IsNotExist(err) {
						return errors.New("File not found")
					}
				}
				return nil
			},
			Default: "~/.ssh/id_rsa",
		}

		result, err := prompt.Run()
		if err != nil {
			return err
		}
		tritonKeyPath = result
	}

	// Triton Key ID
	if viper.IsSet("triton_key_id") {
		tritonKeyID = viper.GetString("triton_key_id")
	} else {
		fingerprint, err := shell.GetPublicKeyFingerprintFromPrivateKey(tritonKeyPath)
		if err != nil {
			return err
		}
		tritonKeyID = fingerprint
	}

	// Triton URL
	if viper.IsSet("triton_url") {
		tritonURL = viper.GetString("triton_url")
	} else {
		prompt := promptui.Prompt{
			Label:   "Triton URL",
			Default: "https://us-east-1.api.joyent.com",
		}

		result, err := prompt.Run()
		if err != nil {
			return err
		}
		tritonURL = result
	}

	// We now have enough information to init a triton client
	keyMaterial, err := ioutil.ReadFile(tritonKeyPath)
	if err != nil {
		return err
	}

	sshKeySigner, err := authentication.NewPrivateKeySigner(tritonKeyID, keyMaterial, tritonAccount)
	if err != nil {
		return err
	}

	// Create manta client
	config := &triton.ClientConfig{
		TritonURL:   tritonURL,
		MantaURL:    "https://us-east.manta.joyent.com", // TODO: Make this configurable
		AccountName: tritonAccount,
		Signers:     []authentication.Signer{sshKeySigner},
	}
	tritonStorageClient, err := storage.NewClient(config)
	if err != nil {
		return err
	}

	input := storage.ListDirectoryInput{
		DirectoryName: fmt.Sprintf("/stor/%s", "triton-kubernetes"),
		Limit:         100,
	}

	result, err := tritonStorageClient.Dir().List(context.Background(), &input)
	if err != nil {
		return err
	}

	// Prompt for cluster manager
	prompt := promptui.Select{
		Label: "Cluster Manager to delete",
		Items: result.Entries,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}?",
			Active:   fmt.Sprintf("%s {{ .Name | underline }}", promptui.IconSelect),
			Inactive: " {{.Name}}",
			Selected: fmt.Sprintf(`{{ "%s" | green }} {{ "Cluster Manager:" | bold}} {{ .Name }}`, promptui.IconGood),
		},
	}

	i, _, err := prompt.Run()
	if err != nil {
		return err
	}

	targetManager = result.Entries[i].Name

	// Create a temporary directory
	tempDir, err := ioutil.TempDir("", "triton-kubernetes-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	// Use temporary directory as working directory
	shellOptions := shell.ShellOptions{
		WorkingDir: tempDir,
	}

	// Download main.tf.json for correct manager into temporary directory
	getObjectInput := &storage.GetObjectInput{
		ObjectPath: fmt.Sprintf("/stor/%s/%s/%s", "triton-kubernetes", targetManager, "main.tf.json"),
	}
	output, err := tritonStorageClient.Objects().Get(context.Background(), getObjectInput)
	if err != nil {
		return err
	}

	// Save the main.tf.json to file on disk
	jsonPath := fmt.Sprintf("%s/%s", tempDir, "main.tf.json")
	jsonBytes, err := ioutil.ReadAll(output.ObjectReader)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(jsonPath, jsonBytes, 0644)
	if err != nil {
		return err
	}

	// Run terraform init
	err = shell.RunShellCommand(&shellOptions, "terraform", []string{"init", "-force-copy"}...)
	if err != nil {
		return err
	}

	// Run terraform destroy
	err = shell.RunShellCommand(&shellOptions, "terraform", []string{"destroy", "-force"}...)
	if err != nil {
		return err
	}

	err = deleteTerraformBackendFromManta(tritonStorageClient, targetManager)
	if err != nil {
		return err
	}

	return nil
}

// Deletes the terraform backend for the given cluster manager
// Note: For some reason, attempting to force delete the entire directory only deletes the json file.
// As a workaround, this function explicitly deletes the json file and tfstate file before deleting the directory.
func deleteTerraformBackendFromManta(tritonStorageClient *storage.StorageClient, targetManager string) error {
	objClient := tritonStorageClient.Objects()
	managerFolderPath := fmt.Sprintf("/stor/%s/%s", "triton-kubernetes", targetManager)

	// Deleting the main.tf.json file
	jsonMantaPath := fmt.Sprintf("%s/%s", managerFolderPath, "main.tf.json")
	deleteObjInput := &storage.DeleteObjectInput{
		ObjectPath: jsonMantaPath,
	}
	err := objClient.Delete(context.Background(), deleteObjInput)
	if err != nil {
		return err
	}

	// Deleting the terraform.tfstate file
	tfStatePath := fmt.Sprintf("%s/%s", managerFolderPath, "terraform.tfstate")
	deleteObjInput = &storage.DeleteObjectInput{
		ObjectPath: tfStatePath,
	}
	err = objClient.Delete(context.Background(), deleteObjInput)
	if err != nil {
		return err
	}

	// Deleting the cluster manager directory
	deleteDirInput := &storage.DeleteDirectoryInput{
		DirectoryName: managerFolderPath,
	}
	err = tritonStorageClient.Dir().Delete(context.Background(), deleteDirInput)
	if err != nil {
		return err
	}

	return nil
}
