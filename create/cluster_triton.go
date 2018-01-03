package create

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
	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

type tritonClusterTerraformConfig struct {
	Source string `json:"source"`

	Name string `json:"name"`

	EtcdNodeCount          string `json:"etcd_node_count"`
	OrchestrationNodeCount string `json:"orchestration_node_count"`
	ComputeNodeCount       string `json:"compute_node_count"`

	KubernetesPlaneIsolation string `json:"k8s_plane_isolation"`

	RancherAPIURL    string `json:"rancher_api_url"`
	RancherAccessKey string `json:"rancher_access_key"`
	RancherSecretKey string `json:"rancher_secret_key"`

	TritonAccount string `json:"triton_account"`
	TritonKeyPath string `json:"triton_key_path"`
	TritonKeyID   string `json:"triton_key_id"`
	TritonURL     string `json:"triton_url,omitempty"`

	RancherRegistry         string `json:"rancher_registry,omitempty"`
	RancherRegistryUsername string `json:"rancher_registry_username,omitempty"`
	RancherRegistryPassword string `json:"rancher_registry_password,omitempty"`

	KubernetesRegistry         string `json:"k8s_registry,omitempty"`
	KubernetesRegistryUsername string `json:"k8s_registry_username,omitempty"`
	KubernetesRegistryPassword string `json:"k8s_registry_password,omitempty"`
}

func NewTritonCluster() error {
	cfg := tritonClusterTerraformConfig{}

	// TODO: Move this to const or make configurable
	// cfg.Source = "github.com/joyent/triton-kubernetes//terraform/modules/triton-rancher-k8s"
	cfg.Source = "./terraform/modules/triton-rancher-k8s"

	// Set node counts to 0 since we manage nodes individually in triton-kubernetes cli
	cfg.EtcdNodeCount = "0"
	cfg.OrchestrationNodeCount = "0"
	cfg.ComputeNodeCount = "0"

	// Name
	if viper.IsSet("name") {
		cfg.Name = viper.GetString("name")
	} else {
		prompt := promptui.Prompt{
			Label: "Cluster Name",
		}

		result, err := prompt.Run()
		if err != nil {
			return err
		}
		cfg.Name = result
	}

	if cfg.Name == "" {
		return errors.New("Invalid Name")
	}

	// Kubernetes Plane Isolation
	if viper.IsSet("k8s_plane_isolation") {
		cfg.KubernetesPlaneIsolation = viper.GetString("k8s_plane_isolation")
	} else {
		prompt := promptui.Select{
			Label: "Kubernetes Plane Isolation",
			Items: []string{"required", "none"},
		}

		_, value, err := prompt.Run()
		if err != nil {
			return err
		}

		cfg.KubernetesPlaneIsolation = value
	}

	// Rancher API URL
	cfg.RancherAPIURL = "http://${element(module.cluster-manager.masters, 0)}:8080"

	// Triton Account
	if viper.IsSet("triton_account") {
		cfg.TritonAccount = viper.GetString("triton_account")
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
		cfg.TritonAccount = result
	}

	// Triton Key Path
	if viper.IsSet("triton_key_path") {
		cfg.TritonKeyPath = viper.GetString("triton_key_path")
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
		cfg.TritonKeyPath = result
	}

	// Triton Key ID
	if viper.IsSet("triton_key_id") {
		cfg.TritonKeyID = viper.GetString("triton_key_id")
	} else {
		keyID, err := shell.GetPublicKeyFingerprintFromPrivateKey(cfg.TritonKeyPath)
		if err != nil {
			return err
		}
		cfg.TritonKeyID = keyID
	}

	// Triton URL
	if viper.IsSet("triton_url") {
		cfg.TritonURL = viper.GetString("triton_url")
	} else {
		prompt := promptui.Prompt{
			Label:   "Triton URL",
			Default: "https://us-east-1.api.joyent.com",
		}

		result, err := prompt.Run()
		if err != nil {
			return err
		}
		cfg.TritonURL = result
	}

	// We now have enough information to init a triton client
	keyMaterial, err := ioutil.ReadFile(cfg.TritonKeyPath)
	if err != nil {
		return err
	}

	sshKeySigner, err := authentication.NewPrivateKeySigner(cfg.TritonKeyID, keyMaterial, cfg.TritonAccount)
	if err != nil {
		return err
	}

	// Create manta client
	config := &triton.ClientConfig{
		TritonURL:   cfg.TritonURL,
		MantaURL:    "https://us-east.manta.joyent.com", // TODO: Make this configurable
		AccountName: cfg.TritonAccount,
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

	// Cluster manager
	var targetManager string
	if viper.IsSet("cluster_manager") {
		targetManager = viper.GetString("cluster_manager")
	} else {
		// Prompt for cluster manager
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
		targetManager = result.Entries[i].Name
	}

	// Create a temporary directory
	tempDir, err := ioutil.TempDir("", "triton-kubernetes-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	// Load current tf config from manta
	tfJSONMantaPath := fmt.Sprintf("/stor/%s/%s/%s", "triton-kubernetes", targetManager, "main.tf.json")
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

	// Use combination of cluster prefix and cluster name for the key
	// The cluster prefix will be used to differentiate the clusters from the
	// cluster manager in the tf config file
	clusterKey := fmt.Sprintf("cluster_%s", cfg.Name)
	parsedConfig.SetP(&cfg, fmt.Sprintf("module.%s", clusterKey))

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

	// Run terraform apply
	err = shell.RunShellCommand(&shellOptions, "terraform", "apply", "-auto-approve")
	if err != nil {
		return err
	}

	// After terraform succeeds, save main.tf.json to manta
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
