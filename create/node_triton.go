package create

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/joyent/triton-kubernetes/backend"
	"github.com/joyent/triton-kubernetes/shell"
	"github.com/joyent/triton-kubernetes/state"

	triton "github.com/joyent/triton-go"
	"github.com/joyent/triton-go/authentication"
	"github.com/joyent/triton-go/compute"
	"github.com/joyent/triton-go/network"
	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

const (
	tritonNodeKeyFormat                            = "module.node_triton_%s"
	tritonRancherKubernetesHostTerraformModulePath = "terraform/modules/triton-rancher-k8s-host"
)

type tritonNodeTerraformConfig struct {
	baseNodeTerraformConfig

	TritonAccount string `json:"triton_account"`
	TritonKeyPath string `json:"triton_key_path"`
	TritonKeyID   string `json:"triton_key_id"`
	TritonURL     string `json:"triton_url,omitempty"`

	TritonNetworkNames   []string `json:"triton_network_names,omitempty"`
	TritonImageName      string   `json:"triton_image_name,omitempty"`
	TritonImageVersion   string   `json:"triton_image_version,omitempty"`
	TritonSSHUser        string   `json:"triton_ssh_user,omitempty"`
	TritonMachinePackage string   `json:"triton_machine_package,omitempty"`
}

func newTritonNode(selectedClusterManager, selectedCluster string, remoteBackend backend.Backend, state state.State) error {
	baseConfig, err := getBaseNodeTerraformConfig(tritonRancherKubernetesHostTerraformModulePath, selectedCluster, state)
	if err != nil {
		return err
	}

	cfg := tritonNodeTerraformConfig{
		baseNodeTerraformConfig: baseConfig,

		// Grab variables from cluster config
		TritonAccount: state.Get(fmt.Sprintf("module.%s.triton_account", selectedCluster)),
		TritonKeyPath: state.Get(fmt.Sprintf("module.%s.triton_key_path", selectedCluster)),
		TritonKeyID:   state.Get(fmt.Sprintf("module.%s.triton_key_id", selectedCluster)),
		TritonURL:     state.Get(fmt.Sprintf("module.%s.triton_url", selectedCluster)),
	}

	keyMaterial, err := ioutil.ReadFile(cfg.TritonKeyPath)
	if err != nil {
		return err
	}

	privateKeySignerInput := authentication.PrivateKeySignerInput{
		KeyID:              cfg.TritonKeyID,
		PrivateKeyMaterial: keyMaterial,
		AccountName:        cfg.TritonAccount,
	}
	sshKeySigner, err := authentication.NewPrivateKeySigner(privateKeySignerInput)
	if err != nil {
		return err
	}

	config := &triton.ClientConfig{
		TritonURL:   cfg.TritonURL,
		AccountName: cfg.TritonAccount,
		Signers:     []authentication.Signer{sshKeySigner},
	}

	tritonNetworkClient, err := network.NewClient(config)
	if err != nil {
		return err
	}

	networks, err := tritonNetworkClient.List(context.Background(), nil)
	if err != nil {
		return err
	}

	// Triton Network Names
	if viper.IsSet("triton_network_names") {
		cfg.TritonNetworkNames = viper.GetStringSlice("triton_network_names")

		// Verify triton network names
		validNetworksMap := map[string]struct{}{}
		validNetworksSlice := []string{}
		for _, validNetwork := range networks {
			validNetworksMap[validNetwork.Name] = struct{}{}
			validNetworksSlice = append(validNetworksSlice, validNetwork.Name)
		}

		for _, network := range cfg.TritonNetworkNames {
			if _, ok := validNetworksMap[network]; !ok {
				return fmt.Errorf("Invalid Triton Network '%s', must be one of the following: %s", network, strings.Join(validNetworksSlice, ", "))
			}
		}
	} else {
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

	tritonComputeClient, err := compute.NewClient(config)
	if err != nil {
		return err
	}

	// Triton Image Name and Triton Image Version
	if viper.IsSet("triton_image_name") && viper.IsSet("triton_image_version") {
		cfg.TritonImageName = viper.GetString("triton_image_name")
		cfg.TritonImageVersion = viper.GetString("triton_image_version")

		// TODO: Verify Triton Image Name/Version
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

		// TODO: Verify triton_machine_package
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

	// Add new node to terraform config
	err = state.Add(fmt.Sprintf(tritonNodeKeyFormat, cfg.Hostname), &cfg)
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
	jsonPath := fmt.Sprintf("%s/%s", tempDir, "main.tf.json")
	err = ioutil.WriteFile(jsonPath, state.Bytes(), 0644)
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

	// After terraform succeeds, commit state
	err = remoteBackend.PersistState(state)
	if err != nil {
		return err
	}

	return nil
}
