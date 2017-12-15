package create

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	triton "github.com/joyent/triton-go"
	"github.com/joyent/triton-go/authentication"
	"github.com/joyent/triton-go/compute"
	"github.com/joyent/triton-go/network"
	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

// This struct represents the definition of a Terraform .tf file.
// Marshalled into json this struct can be passed directly to Terraform.
type tritonManagerTerraformConfig struct {
	Module struct {
		Name struct {
			Source string `json:"source"`

			Name string `json:"name"`
			HA   bool   `json:"ha"`

			TritonAccount string `json:"triton_account"`
			TritonKeyPath string `json:"triton_key_path"`
			TritonKeyID   string `json:"triton_key_id"`
			TritonURL     string `json:"triton_url,omitempty"`

			TritonNetworkNames   []string `json:"triton_network_names,omitempty"`
			TritonImageName      string   `json:"triton_image_name,omitempty"`
			TritonImageVersion   string   `json:"triton_image_version,omitempty"`
			TritonSSHUser        string   `json:"triton_ssh_user,omitempty"`
			TritonMachinePackage string   `json:"triton_machine_package,omitempty"`

			RancherServerImage      string `json:"rancher_server_image,omitempty"`
			RancherAgentImage       string `json:"rancher_agent_image,omitempty"`
			RancherRegistry         string `json:"rancher_registry,omitempty"`
			RancherRegistryUsername string `json:"rancher_registry_username,omitempty"`
			RancherRegistryPassword string `json:"rancher_registry_password,omitempty"`
		} `json:"name"`
	} `json:"module"`
}

func NewTritonManager() error {
	cfg := tritonManagerTerraformConfig{}

	// TODO: Move this to const or make configurable
	cfg.Module.Name.Source = "github.com/joyent/triton-kubernetes//terraform/modules/triton-rancher"

	// Validate configuration provided, if no configuration provided ask user for input.

	// Name
	if viper.IsSet("name") {
		cfg.Module.Name.Name = viper.GetString("name")
	} else {
		prompt := promptui.Prompt{
			Label: "Cluster Manager Name",
		}

		result, err := prompt.Run()
		if err != nil {
			return err
		}
		cfg.Module.Name.Name = result
	}

	if cfg.Module.Name.Name == "" {
		return errors.New("Invalid Name")
	}

	// HA
	if viper.IsSet("ha") {
		cfg.Module.Name.HA = viper.GetBool("ha")
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

		cfg.Module.Name.HA = options[i].Value
	}

	// Triton Account
	if viper.IsSet("triton_account") {
		cfg.Module.Name.TritonAccount = viper.GetString("triton_account")
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
		cfg.Module.Name.TritonAccount = result
	}

	// Triton Key Path
	if viper.IsSet("triton_key_path") {
		cfg.Module.Name.TritonKeyPath = viper.GetString("triton_key_path")
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
		cfg.Module.Name.TritonKeyPath = result
	}

	// Triton Key ID
	if viper.IsSet("triton_key_id") {
		cfg.Module.Name.TritonKeyID = viper.GetString("triton_key_id")
	} else {
		// TODO: Exec ssh-keygen -E md5 -lf PATH_TO_FILE
	}

	// Triton URL
	if viper.IsSet("triton_url") {
		cfg.Module.Name.TritonURL = viper.GetString("triton_url")
	} else {
		prompt := promptui.Prompt{
			Label:   "Triton URL",
			Default: "https://us-east-1.api.joyent.com",
		}

		result, err := prompt.Run()
		if err != nil {
			return err
		}
		cfg.Module.Name.TritonURL = result
	}

	// We now have enough information to init a triton client
	fingerprint := "68:9f:9a:c4:76:3a:f4:62:77:47:3e:47:d4:34:4a:b7"
	keyMaterial, err := ioutil.ReadFile(cfg.Module.Name.TritonKeyPath)
	if err != nil {
		return err
	}

	sshKeySigner, err := authentication.NewPrivateKeySigner(fingerprint, keyMaterial, cfg.Module.Name.TritonAccount)
	if err != nil {
		return err
	}

	config := &triton.ClientConfig{
		TritonURL:   cfg.Module.Name.TritonURL,
		AccountName: cfg.Module.Name.TritonAccount,
		Signers:     []authentication.Signer{sshKeySigner},
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
		cfg.Module.Name.TritonNetworkNames = viper.GetStringSlice("triton_network_names")
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

		cfg.Module.Name.TritonNetworkNames = []string{networks[i].Name}
	}

	if viper.IsSet("triton_image_name") && viper.IsSet("triton_image_version") {
		cfg.Module.Name.TritonImageName = viper.GetString("triton_image_name")
		cfg.Module.Name.TritonImageVersion = viper.GetString("triton_image_version")
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

		cfg.Module.Name.TritonImageName = images[i].Name
		cfg.Module.Name.TritonImageVersion = images[i].Version
	}

	if viper.IsSet("triton_ssh_user") {
		cfg.Module.Name.TritonSSHUser = viper.GetString("triton_ssh_user")
	} else {
		prompt := promptui.Prompt{
			Label:   "Triton SSH User",
			Default: "root",
		}

		result, err := prompt.Run()
		if err != nil {
			return err
		}
		cfg.Module.Name.TritonSSHUser = result
	}

	if viper.IsSet("triton_machine_package") {
		cfg.Module.Name.TritonMachinePackage = viper.GetString("triton_machine_package")
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
			Label: "Triton Machine Package to use",
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

		cfg.Module.Name.TritonMachinePackage = kvmPackages[i].Name
	}

	jsonBytes, err := json.MarshalIndent(&cfg, "", "\t")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonBytes))

	return nil
}
