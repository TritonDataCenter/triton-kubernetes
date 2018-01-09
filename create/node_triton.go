package create

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Jeffail/gabs"
	triton "github.com/joyent/triton-go"
	"github.com/joyent/triton-go/authentication"
	"github.com/joyent/triton-go/compute"
	"github.com/joyent/triton-go/network"
	"github.com/joyent/triton-go/storage"
	"github.com/joyent/triton-kubernetes/shell"
	"github.com/joyent/triton-kubernetes/util"
	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

type tritonNodeTerraformConfig struct {
	Source string `json:"source"`

	Hostname string `json:"hostname"`

	RancherAPIURL        string                  `json:"rancher_api_url"`
	RancherAccessKey     string                  `json:"rancher_access_key"`
	RancherSecretKey     string                  `json:"rancher_secret_key"`
	RancherEnvironmentID string                  `json:"rancher_environment_id"`
	RancherHostLabels    rancherHostLabelsConfig `json:"rancher_host_labels"`

	TritonAccount string `json:"triton_account"`
	TritonKeyPath string `json:"triton_key_path"`
	TritonKeyID   string `json:"triton_key_id"`
	TritonURL     string `json:"triton_url,omitempty"`

	TritonNetworkNames   []string `json:"triton_network_names,omitempty"`
	TritonImageName      string   `json:"triton_image_name,omitempty"`
	TritonImageVersion   string   `json:"triton_image_version,omitempty"`
	TritonSSHUser        string   `json:"triton_ssh_user,omitempty"`
	TritonMachinePackage string   `json:"triton_machine_package,omitempty"`

	RancherRegistry         string `json:"rancher_registry,omitempty"`
	RancherRegistryUsername string `json:"rancher_registry_username,omitempty"`
	RancherRegistryPassword string `json:"rancher_registry_password,omitempty"`

	KubernetesRegistry         string `json:"k8s_registry,omitempty"`
	KubernetesRegistryUsername string `json:"k8s_registry_username,omitempty"`
	KubernetesRegistryPassword string `json:"k8s_registry_password,omitempty"`
}

type rancherHostLabelsConfig struct {
	Orchestration bool `json:"orchestration,omitempty"`
	Etcd          bool `json:"etcd,omitempty"`
	Compute       bool `json:"compute,omitempty"`
}

func NewTritonNode() error {
	var hostLabel string
	var clusterManager string
	var clusterKey string

	cfg := tritonNodeTerraformConfig{}

	// TODO: Move this to const or make configurable
	cfg.Source = "./terraform/modules/triton-rancher-k8s-host"

	// hostname
	if viper.IsSet("hostname") {
		cfg.Hostname = viper.GetString("hostname")
	} else {
		prompt := promptui.Prompt{
			Label: "Hostname",
		}

		result, err := prompt.Run()
		if err != nil {
			return err
		}
		cfg.Hostname = result
	}

	if cfg.Hostname == "" {
		return errors.New("Invalid Hostname")
	}

	// Rancher Host Label
	hostLabelOptions := []string{
		"compute",
		"etcd",
		"orchestration",
	}
	if viper.IsSet("rancher_host_label") {
		hostLabel = viper.GetString("rancher_host_label")
	} else {
		prompt := promptui.Select{
			Label: "Which type of node?",
			Items: hostLabelOptions,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}?",
				Active:   fmt.Sprintf("%s {{ . | underline }}", promptui.IconSelect),
				Inactive: "  {{ . }}",
				Selected: fmt.Sprintf(`{{ "%s" | green }} {{ "Highly Available:" | bold}} {{ . }}`, promptui.IconGood),
			},
		}

		i, _, err := prompt.Run()
		if err != nil {
			return err
		}

		hostLabel = hostLabelOptions[i]
	}

	if hostLabel == "compute" {
		cfg.RancherHostLabels.Compute = true
	} else if hostLabel == "etcd" {
		cfg.RancherHostLabels.Etcd = true
	} else if hostLabel == "orchestration" {
		cfg.RancherHostLabels.Orchestration = true
	} else {
		return errors.New("Invalid rancher host label")
	}

	// Rancher API URL
	cfg.RancherAPIURL = "http://${element(module.cluster-manager.masters, 0)}:8080"

	// Triton account
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

	// Cluster Manager
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

	// Cluster Name
	if viper.IsSet("cluster_name") {
		clusterName := viper.GetString("cluster_name")
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

	// Rancher Environment ID
	cfg.RancherEnvironmentID = fmt.Sprintf("${module.%s.environment_id}", clusterKey)

	tritonComputeClient, err := compute.NewClient(config)
	if err != nil {
		return err
	}

	tritonNetworkClient, err := network.NewClient(config)
	if err != nil {
		return err
	}

	// Triton Network Names
	if viper.IsSet("triton_network_names") {
		cfg.TritonNetworkNames = viper.GetStringSlice("triton_network_names")
	} else {
		networks, err := tritonNetworkClient.List(context.Background(), nil)
		if err != nil {
			return err
		}

		prompt := promptui.Select{
			Label: "Triton Networks to attach",
			Items: networks,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}?",
				Active:   fmt.Sprintf("%s {{ .Name | underline }}", promptui.IconSelect),
				Inactive: "  {{.Name}}",
				Selected: fmt.Sprintf(`{{ "%s" | green }} {{ "Triton Networks:" | bold}} {{ .Name }}`, promptui.IconGood),
			},
		}

		i, _, err := prompt.Run()
		if err != nil {
			return err
		}

		cfg.TritonNetworkNames = []string{networks[i].Name}
	}

	// Triton Image Name and Triton Image Version
	if viper.IsSet("triton_image_name") && viper.IsSet("triton_image_version") {
		cfg.TritonImageName = viper.GetString("triton_image_name")
		cfg.TritonImageVersion = viper.GetString("triton_image_version")
	} else {
		listImageInput := compute.ListImagesInput{}
		images, err := tritonComputeClient.Images().List(context.Background(), &listImageInput)
		if err != nil {
			return err
		}

		searcher := func(input string, index int) bool {
			image := images[index]
			name := strings.Replace(strings.ToLower(image.Name), " ", "", -1)
			input = strings.Replace(strings.ToLower(input), " ", "", -1)

			return strings.Contains(name, input)
		}

		prompt := promptui.Select{
			Label: "Triton Image to use",
			Items: images,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}?",
				Active:   fmt.Sprintf(`%s {{ .Name | underline }}{{ "@" | underline }}{{ .Version | underline }}`, promptui.IconSelect),
				Inactive: `  {{ .Name }}@{{ .Version }}`,
				Selected: fmt.Sprintf(`{{ "%s" | green }} {{ "Triton Image:" | bold}} {{ .Name }}{{ "@" }}{{ .Version }}`, promptui.IconGood),
			},
			Searcher: searcher,
		}

		i, _, err := prompt.Run()
		if err != nil {
			return err
		}

		cfg.TritonImageName = images[i].Name
		cfg.TritonImageVersion = images[i].Version
	}

	// Triton SSH User
	if viper.IsSet("triton_ssh_user") {
		cfg.TritonSSHUser = viper.GetString("triton_ssh_user")
	} else {
		prompt := promptui.Prompt{
			Label:   "Triton SSH User",
			Default: "root",
		}

		result, err := prompt.Run()
		if err != nil {
			return err
		}
		cfg.TritonSSHUser = result
	}

	// Triton Machine Package
	if viper.IsSet("triton_machine_package") {
		cfg.TritonMachinePackage = viper.GetString("triton_machine_package")
	} else {
		listPackageInput := compute.ListPackagesInput{}
		packages, err := tritonComputeClient.Packages().List(context.Background(), &listPackageInput)
		if err != nil {
			return err
		}

		// Filter to only kvm packages
		kvmPackages := []*compute.Package{}
		for _, pkg := range packages {
			if strings.Contains(pkg.Name, "kvm") {
				kvmPackages = append(kvmPackages, pkg)
			}
		}

		searcher := func(input string, index int) bool {
			pkg := kvmPackages[index]
			name := strings.Replace(strings.ToLower(pkg.Name), " ", "", -1)
			input = strings.Replace(strings.ToLower(input), " ", "", -1)

			return strings.Contains(name, input)
		}

		prompt := promptui.Select{
			Label: "Triton Machine Package to use for node",
			Items: kvmPackages,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}?",
				Active:   fmt.Sprintf(`%s {{ .Name | underline }}`, promptui.IconSelect),
				Inactive: `  {{ .Name }}`,
				Selected: fmt.Sprintf(`{{ "%s" | green }} {{ "Triton Machine Package:" | bold}} {{ .Name }}`, promptui.IconGood),
			},
			Searcher: searcher,
		}

		i, _, err := prompt.Run()
		if err != nil {
			return err
		}

		cfg.TritonMachinePackage = kvmPackages[i].Name
	}

	// Rancher Registry
	if viper.IsSet("rancher_registry") {
		cfg.RancherRegistry = viper.GetString("rancher_registry")
	} else {
		prompt := promptui.Prompt{
			Label: "Rancher Registry",
		}

		result, err := prompt.Run()
		if err != nil {
			return err
		}
		cfg.RancherRegistry = result
	}

	// Ask for rancher registry username/password only if rancher registry is given
	if cfg.RancherRegistry != "" {
		// Rancher Registry Username
		if viper.IsSet("rancher_registry_username") {
			cfg.RancherRegistryUsername = viper.GetString("rancher_registry_username")
		} else {
			prompt := promptui.Prompt{
				Label: "Rancher Registry Username",
			}

			result, err := prompt.Run()
			if err != nil {
				return err
			}
			cfg.RancherRegistryUsername = result
		}

		// Rancher Registry Password
		if viper.IsSet("rancher_registry_password") {
			cfg.RancherRegistryPassword = viper.GetString("rancher_registry_password")
		} else {
			prompt := promptui.Prompt{
				Label: "Rancher Registry Password",
			}

			result, err := prompt.Run()
			if err != nil {
				return err
			}
			cfg.RancherRegistryPassword = result
		}
	}

	// Add node configuration to tf config
	nodeKey := fmt.Sprintf("node_%s", cfg.Hostname)
	parsedConfig.SetP(&cfg, fmt.Sprintf("module.%s", nodeKey))

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
