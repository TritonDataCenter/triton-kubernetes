package create

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/joyent/triton-kubernetes/backend"
	"github.com/joyent/triton-kubernetes/shell"
	homedir "github.com/mitchellh/go-homedir"

	triton "github.com/joyent/triton-go"
	"github.com/joyent/triton-go/authentication"
	"github.com/joyent/triton-go/compute"
	"github.com/joyent/triton-go/network"
	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

const (
	tritonRancherTerraformModulePath = "terraform/modules/triton-rancher"
)

// This struct represents the definition of a Terraform .tf file.
// Marshalled into json this struct can be passed directly to Terraform.
type tritonManagerTerraformConfig struct {
	Source string `json:"source"`

	Name            string `json:"name"`
	HA              bool   `json:"ha"`
	MasterNodeCount int    `json:"gcm_node_count"`

	TritonAccount string `json:"triton_account"`
	TritonKeyPath string `json:"triton_key_path"`
	TritonKeyID   string `json:"triton_key_id"`
	TritonURL     string `json:"triton_url,omitempty"`

	TritonNetworkNames         []string `json:"triton_network_names,omitempty"`
	TritonImageName            string   `json:"triton_image_name,omitempty"`
	TritonImageVersion         string   `json:"triton_image_version,omitempty"`
	TritonSSHUser              string   `json:"triton_ssh_user,omitempty"`
	MasterTritonMachinePackage string   `json:"master_triton_machine_package,omitempty"`

	TritonMySQLImageName        string `json:"triton_mysql_image_name,omitempty"`
	TritonMySQLImageVersion     string `json:"triton_mysql_image_version,omitempty"`
	MySQLDBTritonMachinePackage string `json:"mysqldb_triton_machine_package,omitempty"`

	RancherServerImage      string `json:"rancher_server_image,omitempty"`
	RancherAgentImage       string `json:"rancher_agent_image,omitempty"`
	RancherRegistry         string `json:"rancher_registry,omitempty"`
	RancherRegistryUsername string `json:"rancher_registry_username,omitempty"`
	RancherRegistryPassword string `json:"rancher_registry_password,omitempty"`
}

func NewTritonManager(remoteBackend backend.Backend) error {
	cfg := tritonManagerTerraformConfig{}

	baseSource := defaultSourceURL
	if viper.IsSet("source_url") {
		baseSource = viper.GetString("source_url")
	}

	baseSourceRef := defaultSourceRef
	if viper.IsSet("source_ref") {
		baseSourceRef = viper.GetString("source_ref")
	}

	// Module Source location e.g. github.com/joyent/triton-kubernetes//terraform/modules/triton-rancher?ref=master
	cfg.Source = fmt.Sprintf("%s//%s?ref=%s", baseSource, tritonRancherTerraformModulePath, baseSourceRef)

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
		return errors.New("Invalid Cluster Manager Name")
	}

	// Validate that a cluster manager with the same name doesn't already exist.
	existingClusterManagers, err := remoteBackend.States()
	if err != nil {
		return err
	}

	found := false
	for _, clusterManagerName := range existingClusterManagers {
		if cfg.Name == clusterManagerName {
			found = true
			break
		}
	}
	if found {
		return fmt.Errorf("A Cluster Manager with the name '%s' already exists.", cfg.Name)
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

	if !cfg.HA {
		cfg.MasterNodeCount = 1
	} else {
		// HA enabled, ask user how many master nodes
		if viper.IsSet("gcm_node_count") {
			cfg.MasterNodeCount = viper.GetInt("gcm_node_count")
		} else {
			prompt := promptui.Prompt{
				Label: "How many master nodes",
				Validate: func(input string) error {
					masterNodeCount, err := strconv.Atoi(input)
					if err != nil {
						return err
					}

					if masterNodeCount < 2 {
						return errors.New("Minimum nodes for HA is 2")
					}

					return nil
				},
				Default: "2",
			}

			result, err := prompt.Run()
			if err != nil {
				return err
			}

			masterNodeCount, err := strconv.Atoi(result)
			if err != nil {
				return err
			}

			cfg.MasterNodeCount = masterNodeCount
		}
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
	rawTritonKeyPath := ""
	if viper.IsSet("triton_key_path") {
		rawTritonKeyPath = viper.GetString("triton_key_path")
	} else {
		prompt := promptui.Prompt{
			Label: "Triton Key Path",
			Validate: func(input string) error {
				expandedPath, err := homedir.Expand(input)
				if err != nil {
					return err
				}

				_, err = os.Stat(expandedPath)
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
		rawTritonKeyPath = result
	}

	expandedTritonKeyPath, err := homedir.Expand(rawTritonKeyPath)
	if err != nil {
		return err
	}
	cfg.TritonKeyPath = expandedTritonKeyPath

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
		MantaURL:    "https://us-east.manta.joyent.com",
		AccountName: cfg.TritonAccount,
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
		networkPrompt := promptui.Select{
			Label: "Triton Networks to attach",
			Items: networks,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}?",
				Active:   fmt.Sprintf("%s {{ .Name | underline }}", promptui.IconSelect),
				Inactive: "  {{.Name}}",
				Selected: fmt.Sprintf(`{{ "%s" | green }} {{ "Triton Networks:" | bold}} {{ .Name }}`, promptui.IconGood),
			},
		}

		continueOptions := []struct {
			Name  string
			Value bool
		}{
			{"Yes", true},
			{"No", false},
		}

		continuePrompt := promptui.Select{
			Label: "Attach another",
			Items: continueOptions,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}?",
				Active:   fmt.Sprintf("%s {{ .Name | underline }}", promptui.IconSelect),
				Inactive: "  {{.Name}}",
				Selected: "  Attach another? {{.Name}}",
			},
		}

		networksChosen := []string{}
		shouldPrompt := true
		for shouldPrompt {
			// Network Prompt
			i, _, err := networkPrompt.Run()
			if err != nil {
				return err
			}
			networksChosen = append(networksChosen, networks[i].Name)

			// Remove the chosen network from the list of choices
			networkChoices := networkPrompt.Items.([]*network.Network)
			remainingChoices := append(networkChoices[:i], networkChoices[i+1:]...)

			if len(remainingChoices) == 0 {
				shouldPrompt = false
			} else {
				networkPrompt.Items = remainingChoices
				// Continue Prompt
				i, _, err = continuePrompt.Run()
				if err != nil {
					return err
				}
				shouldPrompt = continueOptions[i].Value
			}
		}

		cfg.TritonNetworkNames = networksChosen
	}

	// Triton Image
	if viper.IsSet("triton_image_name") && viper.IsSet("triton_image_version") {
		cfg.TritonImageName = viper.GetString("triton_image_name")
		cfg.TritonImageVersion = viper.GetString("triton_image_version")
	} else {
		listImageInput := compute.ListImagesInput{
			Name: "ubuntu-certified-16.04",
		}
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

	// Triton MySQL Image
	if cfg.HA && viper.IsSet("triton_mysql_image_name") && viper.IsSet("triton_mysql_image_version") {
		cfg.TritonMySQLImageName = viper.GetString("triton_mysql_image_name")
		cfg.TritonMySQLImageVersion = viper.GetString("triton_mysql_image_version")
	} else if cfg.HA {
		listImageInput := compute.ListImagesInput{
			Name: "ubuntu-certified-16.04",
		}
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
			Label: "Triton MySQL Image to use",
			Items: images,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}?",
				Active:   fmt.Sprintf(`%s {{ .Name | underline }}{{ "@" | underline }}{{ .Version | underline }}`, promptui.IconSelect),
				Inactive: `  {{ .Name }}@{{ .Version }}`,
				Selected: fmt.Sprintf(`{{ "%s" | green }} {{ "Triton MySQL Image:" | bold}} {{ .Name }}{{ "@" }}{{ .Version }}`, promptui.IconGood),
			},
			Searcher: searcher,
		}

		i, _, err := prompt.Run()
		if err != nil {
			return err
		}

		cfg.TritonMySQLImageName = images[i].Name
		cfg.TritonMySQLImageVersion = images[i].Version
	}

	// MySQL DB Triton Machine Package
	if cfg.HA && viper.IsSet("mysqldb_triton_machine_package") {
		cfg.MySQLDBTritonMachinePackage = viper.GetString("mysqldb_triton_machine_package")
	} else if cfg.HA {
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
			Label: "MySQL DB Triton Machine Package to use for Rancher Master",
			Items: kvmPackages,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}?",
				Active:   fmt.Sprintf(`%s {{ .Name | underline }}`, promptui.IconSelect),
				Inactive: `  {{ .Name }}`,
				Selected: fmt.Sprintf(`{{ "%s" | green }} {{ "MySQL DB Triton Machine Package:" | bold}} {{ .Name }}`, promptui.IconGood),
			},
			Searcher: searcher,
		}

		i, _, err := prompt.Run()
		if err != nil {
			return err
		}

		cfg.MySQLDBTritonMachinePackage = kvmPackages[i].Name
	}

	state, err := remoteBackend.State(cfg.Name)
	if err != nil {
		return err
	}

	state.Add("module.cluster-manager", &cfg)
	state.Add(remoteBackend.StateTerraformConfig(cfg.Name))

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
