package create

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/Jeffail/gabs"
	triton "github.com/joyent/triton-go"
	"github.com/joyent/triton-go/authentication"
	"github.com/joyent/triton-go/client"
	"github.com/joyent/triton-go/compute"
	"github.com/joyent/triton-go/network"
	"github.com/joyent/triton-go/storage"
	"github.com/joyent/triton-kubernetes/shell"
	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

// This struct represents the definition of a Terraform .tf file.
// Marshalled into json this struct can be passed directly to Terraform.
type tritonManagerTerraformConfig struct {
	Source string `json:"source"`

	Name string `json:"name"`
	HA   bool   `json:"ha"`

	TritonAccount string `json:"triton_account"`
	TritonKeyPath string `json:"triton_key_path"`
	TritonKeyID   string `json:"triton_key_id"`
	TritonURL     string `json:"triton_url,omitempty"`

	TritonNetworkNames         []string `json:"triton_network_names,omitempty"`
	TritonImageName            string   `json:"triton_image_name,omitempty"`
	TritonImageVersion         string   `json:"triton_image_version,omitempty"`
	TritonSSHUser              string   `json:"triton_ssh_user,omitempty"`
	MasterTritonMachinePackage string   `json:"master_triton_machine_package,omitempty"`

	RancherServerImage      string `json:"rancher_server_image,omitempty"`
	RancherAgentImage       string `json:"rancher_agent_image,omitempty"`
	RancherRegistry         string `json:"rancher_registry,omitempty"`
	RancherRegistryUsername string `json:"rancher_registry_username,omitempty"`
	RancherRegistryPassword string `json:"rancher_registry_password,omitempty"`
}

type mantaTerraformBackendConfig struct {
	Account     string `json:"account"`
	KeyMaterial string `json:"key_material"`
	KeyID       string `json:"key_id"`
	Path        string `json:"path"`
}

func NewTritonManager() error {
	cfg := tritonManagerTerraformConfig{}

	// TODO: Move this to const or make configurable
	cfg.Source = "github.com/joyent/triton-kubernetes//terraform/modules/triton-rancher"

	// Name
	if viper.IsSet("name") {
		cfg.Name = viper.GetString("name")
	} else {
		prompt := promptui.Prompt{
			Label: "Cluster Manager Name",
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

	// HA
	if viper.IsSet("ha") {
		cfg.HA = viper.GetBool("ha")
	} else {
		options := []struct {
			Name  string
			Value bool
		}{
			{
				"Yes",
				true,
			},
			{
				"No",
				false,
			},
		}

		prompt := promptui.Select{
			Label: "Make Cluster Manager Highly Available?",
			Items: options,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}?",
				Active:   fmt.Sprintf("%s {{ .Name | underline }}", promptui.IconSelect),
				Inactive: "  {{.Name}}",
				Selected: fmt.Sprintf(`{{ "%s" | green }} {{ "Highly Available:" | bold}} {{ .Name }}`, promptui.IconGood),
			},
		}

		i, _, err := prompt.Run()
		if err != nil {
			return err
		}

		cfg.HA = options[i].Value
	}

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

	config := &triton.ClientConfig{
		TritonURL:   cfg.TritonURL,
		MantaURL:    "https://us-east.manta.joyent.com", // TODO: Make this configurable
		AccountName: cfg.TritonAccount,
		Signers:     []authentication.Signer{sshKeySigner},
	}

	terraformBackendConfig := mantaTerraformBackendConfig{
		Account:     cfg.TritonAccount,
		KeyMaterial: cfg.TritonKeyPath,
		KeyID:       cfg.TritonKeyID,
		Path:        fmt.Sprintf("/triton-kubernetes/%s/", cfg.Name),
	}

	// Validate that a manta folder with the same cluster name doesn't exist.
	// We leverage manta to store both terraform state and the terraform json configuration.
	tritonStorageClient, err := storage.NewClient(config)
	if err != nil {
		return err
	}

	input := storage.ListDirectoryInput{
		DirectoryName: fmt.Sprintf("/stor/%s", terraformBackendConfig.Path),
		Limit:         100,
	}

	_, err = tritonStorageClient.Dir().List(context.Background(), &input)
	if err != nil && !client.IsResourceNotFoundError(err) {
		return err
	}

	// Verify that the folder path doesn't already exist
	if err == nil || !client.IsResourceNotFoundError(err) {
		fmt.Printf("A cluster manager that's named \"%s\" already exists\n", cfg.Name)
		return nil
	}

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

	if viper.IsSet("master_triton_machine_package") {
		cfg.MasterTritonMachinePackage = viper.GetString("master_triton_machine_package")
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
			Label: "Triton Machine Package to use for Rancher Master",
			Items: kvmPackages,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}?",
				Active:   fmt.Sprintf(`%s {{ .Name | underline }}`, promptui.IconSelect),
				Inactive: `  {{ .Name }}`,
				Selected: fmt.Sprintf(`{{ "%s" | green }} {{ "Rancher Master Triton Machine Package:" | bold}} {{ .Name }}`, promptui.IconGood),
			},
			Searcher: searcher,
		}

		i, _, err := prompt.Run()
		if err != nil {
			return err
		}

		cfg.MasterTritonMachinePackage = kvmPackages[i].Name
	}

	// Create a temporary directory
	tempDir, err := ioutil.TempDir("", "triton-kubernetes-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	configObject := gabs.New()
	configObject.SetP(cfg, "module.cluster-manager")
	configObject.SetP(terraformBackendConfig, "terraform.backend.manta")

	jsonBytes := []byte(configObject.StringIndent("", "\t"))

	jsonPath := fmt.Sprintf("%s/%s", tempDir, "main.tf.json")
	err = ioutil.WriteFile(jsonPath, jsonBytes, 0644)
	if err != nil {
		return err
	}

	// Create manta path so terraform can store state properly
	clusterManagerPath := fmt.Sprintf("/stor/%s", terraformBackendConfig.Path)
	dirInput := storage.PutDirectoryInput{
		DirectoryName: clusterManagerPath,
	}
	err = tritonStorageClient.Dir().Put(context.Background(), &dirInput)
	if err != nil {
		return err
	}

	// Save main.tf.json to manta
	objPath := fmt.Sprintf("%s/%s", clusterManagerPath, "main.tf.json")
	objInput := storage.PutObjectInput{
		ObjectPath:   objPath,
		ContentType:  "application/json",
		ObjectReader: bytes.NewReader(jsonBytes),
	}
	err = tritonStorageClient.Objects().Put(context.Background(), &objInput)
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

	return nil
}
