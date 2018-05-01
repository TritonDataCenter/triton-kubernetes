package create

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"github.com/joyent/triton-kubernetes/state"
	"github.com/joyent/triton-kubernetes/util"

	triton "github.com/joyent/triton-go"
	"github.com/joyent/triton-go/authentication"
	"github.com/joyent/triton-go/compute"
	"github.com/joyent/triton-go/network"
	"github.com/manifoldco/promptui"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

const (
	tritonRancherTerraformModulePath = "terraform/modules/triton-rancher"
)

// This struct represents the definition of a Terraform .tf file.
// Marshalled into json this struct can be passed directly to Terraform.
type tritonManagerTerraformConfig struct {
	baseManagerTerraformConfig

	TritonAccount string `json:"triton_account"`
	TritonKeyPath string `json:"triton_key_path"`
	TritonKeyID   string `json:"triton_key_id"`
	TritonURL     string `json:"triton_url,omitempty"`

	TritonNetworkNames         []string `json:"triton_network_names,omitempty"`
	TritonImageName            string   `json:"triton_image_name,omitempty"`
	TritonImageVersion         string   `json:"triton_image_version,omitempty"`
	TritonSSHUser              string   `json:"triton_ssh_user,omitempty"`
	MasterTritonMachinePackage string   `json:"master_triton_machine_package,omitempty"`
}

func newTritonManager(currentState state.State, name string) error {
	nonInteractiveMode := viper.GetBool("non-interactive")

	baseConfig, err := getBaseManagerTerraformConfig(tritonRancherTerraformModulePath, name)
	if err != nil {
		return err
	}

	cfg := tritonManagerTerraformConfig{
		baseManagerTerraformConfig: baseConfig,
	}

	// Triton Account
	if viper.IsSet("triton_account") {
		cfg.TritonAccount = viper.GetString("triton_account")
	} else if nonInteractiveMode {
		return errors.New("triton_account must be specified")
	} else {
		prompt := promptui.Prompt{
			Label: "Triton Account Name",
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
	} else if nonInteractiveMode {
		return errors.New("triton_key_path must be specified")
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
		keyID, err := util.GetPublicKeyFingerprintFromPrivateKey(cfg.TritonKeyPath)
		if err != nil {
			return err
		}
		cfg.TritonKeyID = keyID
	}

	// Triton URL
	if viper.IsSet("triton_url") {
		cfg.TritonURL = viper.GetString("triton_url")
	} else if nonInteractiveMode {
		return errors.New("triton_url must be specified")
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

	validNetworksMap := map[string]struct{}{}
	validNetworksSlice := []string{}
	for _, validNetwork := range networks {
		validNetworksMap[validNetwork.Name] = struct{}{}
		validNetworksSlice = append(validNetworksSlice, validNetwork.Name)
	}

	// Triton Network Names
	if viper.IsSet("triton_network_names") {
		cfg.TritonNetworkNames = viper.GetStringSlice("triton_network_names")

		// Verify triton network names
		for _, network := range cfg.TritonNetworkNames {
			if _, ok := validNetworksMap[network]; !ok {
				return fmt.Errorf("Invalid Triton Network '%s', must be one of the following: %s", network, strings.Join(validNetworksSlice, ", "))
			}
		}
	} else if nonInteractiveMode {
		return errors.New("triton_network_names must be specified")
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

	// Get existing images
	listImageInput := compute.ListImagesInput{}
	images, err := tritonComputeClient.Images().List(context.Background(), &listImageInput)
	if err != nil {
		return err
	}

	// Sort images by publish date in reverse chronological order
	sort.SliceStable(images, func(i, j int) bool {
		return images[i].PublishedAt.After(images[j].PublishedAt)
	})

	// Triton Image
	if viper.IsSet("triton_image_name") && viper.IsSet("triton_image_version") {
		cfg.TritonImageName = viper.GetString("triton_image_name")
		cfg.TritonImageVersion = viper.GetString("triton_image_version")
		// Verify triton image name and version
		found := false
		for _, image := range images {
			if image.Name == cfg.TritonImageName && image.Version == cfg.TritonImageVersion {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("Invalid Triton Image Name and Version '%s@%s'", cfg.TritonImageName, cfg.TritonImageVersion)
		}
	} else if nonInteractiveMode {
		return errors.New("Both triton_image_name and triton_image_version must be specified")
	} else {
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
	} else if nonInteractiveMode {
		return errors.New("triton_ssh_user must be specified")
	} else {
		prompt := promptui.Prompt{
			Label:   "Triton SSH User",
			Default: "ubuntu",
		}

		result, err := prompt.Run()
		if err != nil {
			return err
		}
		cfg.TritonSSHUser = result
	}

	// Get list of packages
	listPackageInput := compute.ListPackagesInput{}
	packages, err := tritonComputeClient.Packages().List(context.Background(), &listPackageInput)
	if err != nil {
		return err
	}

	// Sort packages by amount of memory in increasing order
	sort.SliceStable(packages, func(i, j int) bool {
		return packages[i].Memory < packages[j].Memory
	})

	if viper.IsSet("master_triton_machine_package") {
		cfg.MasterTritonMachinePackage = viper.GetString("master_triton_machine_package")
		// Verify master triton machine package
		found := false
		for _, pkg := range packages {
			if cfg.MasterTritonMachinePackage == pkg.Name {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("Invalid Master Triton Machine Package '%s'", cfg.MasterTritonMachinePackage)
		}
	} else if nonInteractiveMode {
		return errors.New("master_triton_machine_package must be specified")
	} else {
		searcher := func(input string, index int) bool {
			pkg := packages[index]
			name := strings.Replace(strings.ToLower(pkg.Name), " ", "", -1)
			input = strings.Replace(strings.ToLower(input), " ", "", -1)

			return strings.Contains(name, input)
		}

		prompt := promptui.Select{
			Label: "Triton Machine Package to use for Rancher Master",
			Items: packages,
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

		cfg.MasterTritonMachinePackage = packages[i].Name
	}

	currentState.SetManager(&cfg)

	return nil
}
