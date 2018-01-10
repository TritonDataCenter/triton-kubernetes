package destroy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/Jeffail/gabs"
	triton "github.com/joyent/triton-go"
	"github.com/joyent/triton-go/authentication"
	"github.com/joyent/triton-go/storage"
	"github.com/joyent/triton-kubernetes/shell"
	"github.com/joyent/triton-kubernetes/util"
	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

func DeleteTritonNode() error {
	var tritonAccount string
	var tritonKeyPath string
	var tritonKeyID string
	var tritonURL string
	var clusterManager string
	var clusterKey string
	var nodeKey string

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

	// Stop if there are no cluster managers
	if result.ResultSetSize == 0 {
		fmt.Println("No cluster managers found.")
		return nil
	}

	// Prompt for cluster manager
	if viper.IsSet("cluster_manager") {
		clusterManager = viper.GetString("cluster_manager")
	} else {
		prompt := promptui.Select{
			Label: "Cluster Manager",
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
		clusterManager = result.Entries[i].Name
	}

	// Create a temporary directory
	tempDir, err := ioutil.TempDir("", "triton-kubernetes-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	// Load current tf config from manta
	tfJSONMantaPath := fmt.Sprintf("/stor/%s/%s/%s", "triton-kubernetes", clusterManager, "main.tf.json")
	getObjectInput := &storage.GetObjectInput{
		ObjectPath: tfJSONMantaPath,
	}
	output, err := tritonStorageClient.Objects().Get(context.Background(), getObjectInput)
	if err != nil {
		return err
	}

	currentConfigBytes, err := ioutil.ReadAll(output.ObjectReader)
	if err != nil {
		return err
	}

	parsedConfig, err := gabs.ParseJSON(currentConfigBytes)
	if err != nil {
		return err
	}

	// Get existing clusters
	clusterOptions, err := util.GetClusterOptions(parsedConfig)
	if err != nil {
		return err
	}

	if viper.IsSet("cluster_name") {
		clusterName := viper.GetString("cluster_name")
		// Find cluster key by cluster name
		for _, option := range clusterOptions {
			if clusterName == option.ClusterName {
				clusterKey = option.ClusterKey
				break
			}
		}
	} else {
		prompt := promptui.Select{
			Label: "Cluster",
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
		clusterKey = clusterOptions[i].ClusterKey
	}

	if clusterKey == "" {
		return errors.New("Invalid Cluster Name")
	}

	// Get existing nodes for the given cluster
	nodeOptions, err := util.GetNodeOptions(parsedConfig, clusterKey)
	if err != nil {
		return err
	}

	if viper.IsSet("hostname") {
		hostname := viper.GetString("hostname")
		// Find node by hostname
		for _, option := range nodeOptions {
			if hostname == option.Hostname {
				nodeKey = option.NodeKey
				break
			}
		}
	} else {
		prompt := promptui.Select{
			Label: "Node to delete",
			Items: nodeOptions,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}?",
				Active:   fmt.Sprintf("%s {{ .Hostname | underline }}", promptui.IconSelect),
				Inactive: " {{ .Hostname }}",
				Selected: fmt.Sprintf(`{{ "%s" | green }} {{ "Cluster:" | bold}} {{ .Hostname }}`, promptui.IconGood),
			},
		}

		i, _, err := prompt.Run()
		if err != nil {
			return err
		}
		nodeKey = nodeOptions[i].NodeKey
	}

	if nodeKey == "" {
		return errors.New("Invalid hostname")
	}

	jsonBytes := []byte(parsedConfig.StringIndent("", "\t"))

	// Save the main.tf.json to file on disk
	jsonPath := fmt.Sprintf("%s/%s", tempDir, "main.tf.json")
	err = ioutil.WriteFile(jsonPath, jsonBytes, 0644)
	if err != nil {
		return err
	}

	// Copying ./terraform folder to temporary directory
	// Need to remove this once terraform modules are hosted on github
	err = shell.RunShellCommand(nil, "cp", "-r", "./terraform", tempDir)
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
	targetArg := fmt.Sprintf("-target=module.%s", nodeKey)
	err = shell.RunShellCommand(&shellOptions, "terraform", "destroy", "-force", targetArg)
	if err != nil {
		return err
	}

	// Remove node from tf config
	err = parsedConfig.Delete("module", nodeKey)
	if err != nil {
		return err
	}

	// Save main.tf.json to manta
	jsonBytes = []byte(parsedConfig.StringIndent("", "\t"))
	objInput := storage.PutObjectInput{
		ObjectPath:   tfJSONMantaPath,
		ContentType:  "application/json",
		ObjectReader: bytes.NewReader(jsonBytes),
	}
	err = tritonStorageClient.Objects().Put(context.Background(), &objInput)
	if err != nil {
		return err
	}

	return nil
}