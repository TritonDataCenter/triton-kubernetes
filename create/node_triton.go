package create

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"

	"github.com/mesoform/triton-kubernetes/backend"
	"github.com/mesoform/triton-kubernetes/state"

	triton "github.com/joyent/triton-go"
	"github.com/joyent/triton-go/authentication"
	"github.com/joyent/triton-go/compute"
	"github.com/joyent/triton-go/network"
	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

const (
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

// Adds new Triton nodes to the given cluster and manager.
// Returns:
// - a slice of the hostnames added
// - the new state
// - error or nil
func newTritonNode(selectedClusterManager, selectedCluster string, remoteBackend backend.Backend, currentState state.State) ([]string, error) {
	nonInteractiveMode := viper.GetBool("non-interactive")
	baseConfig, err := getBaseNodeTerraformConfig(tritonRancherKubernetesHostTerraformModulePath, selectedCluster, currentState)
	if err != nil {
		return []string{}, err
	}

	baseConfig.NodeCount, err = getNodeCount(baseConfig)
	if err != nil {
		return []string{}, err
	}

	baseConfig.Hostname, err = getNodeHostnamePrefix()
	if err != nil {
		return []string{}, err
	}

	cfg := tritonNodeTerraformConfig{
		baseNodeTerraformConfig: baseConfig,

		// Grab variables from cluster config
		TritonAccount: currentState.Get(fmt.Sprintf("module.%s.triton_account", selectedCluster)),
		TritonKeyPath: currentState.Get(fmt.Sprintf("module.%s.triton_key_path", selectedCluster)),
		TritonKeyID:   currentState.Get(fmt.Sprintf("module.%s.triton_key_id", selectedCluster)),
		TritonURL:     currentState.Get(fmt.Sprintf("module.%s.triton_url", selectedCluster)),
	}

	keyMaterial, err := ioutil.ReadFile(cfg.TritonKeyPath)
	if err != nil {
		return []string{}, err
	}

	privateKeySignerInput := authentication.PrivateKeySignerInput{
		KeyID:              cfg.TritonKeyID,
		PrivateKeyMaterial: keyMaterial,
		AccountName:        cfg.TritonAccount,
	}
	sshKeySigner, err := authentication.NewPrivateKeySigner(privateKeySignerInput)
	if err != nil {
		return []string{}, err
	}

	config := &triton.ClientConfig{
		TritonURL:   cfg.TritonURL,
		AccountName: cfg.TritonAccount,
		Signers:     []authentication.Signer{sshKeySigner},
	}

	tritonNetworkClient, err := network.NewClient(config)
	if err != nil {
		return []string{}, err
	}

	networks, err := tritonNetworkClient.List(context.Background(), nil)
	if err != nil {
		return []string{}, err
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
				return []string{}, fmt.Errorf("Invalid Triton Network '%s', must be one of the following: %s", network, strings.Join(validNetworksSlice, ", "))
			}
		}
	} else if nonInteractiveMode {
		return []string{}, errors.New("triton_network_names must be specified")
	} else {
		networkPrompt := promptui.Select{
			Label: "Triton Networks to attach",
			Items: networks,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}?",
				Active:   fmt.Sprintf("%s {{ .Name | underline }}", promptui.IconSelect),
				Inactive: "  {{.Name}}",
				Selected: fmt.Sprintf(`{{ "%s" | green }} {{ "Triton Network Attached:" | bold}} {{ .Name }}`, promptui.IconGood),
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
				return []string{}, err
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
					return []string{}, err
				}
				shouldPrompt = continueOptions[i].Value
			}
		}

		cfg.TritonNetworkNames = networksChosen
	}

	tritonComputeClient, err := compute.NewClient(config)
	if err != nil {
		return []string{}, err
	}

	// Triton Image Name and Triton Image Version
	if viper.IsSet("triton_image_name") && viper.IsSet("triton_image_version") {
		cfg.TritonImageName = viper.GetString("triton_image_name")
		cfg.TritonImageVersion = viper.GetString("triton_image_version")

		// TODO: Verify Triton Image Name/Version
	} else if nonInteractiveMode {
		return []string{}, errors.New("Both triton_image_name and triton_image_version must be specified")
	} else {
		listImageInput := compute.ListImagesInput{}
		images, err := tritonComputeClient.Images().List(context.Background(), &listImageInput)
		if err != nil {
			return []string{}, err
		}

		// Sort images by publish date in reverse chronological order
		sort.SliceStable(images, func(i, j int) bool {
			return images[i].PublishedAt.After(images[j].PublishedAt)
		})

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
			return []string{}, err
		}

		cfg.TritonImageName = images[i].Name
		cfg.TritonImageVersion = images[i].Version
	}

	// Triton SSH User
	if viper.IsSet("triton_ssh_user") {
		cfg.TritonSSHUser = viper.GetString("triton_ssh_user")
	} else if nonInteractiveMode {
		return []string{}, errors.New("triton_ssh_user must be specified")
	} else {
		prompt := promptui.Prompt{
			Label:   "Triton SSH User",
			Default: "ubuntu",
		}

		result, err := prompt.Run()
		if err != nil {
			return []string{}, err
		}
		cfg.TritonSSHUser = result
	}

	// Triton Machine Package
	if viper.IsSet("triton_machine_package") {
		cfg.TritonMachinePackage = viper.GetString("triton_machine_package")

		// TODO: Verify triton_machine_package
	} else if nonInteractiveMode {
		return []string{}, errors.New("triton_machine_package must be specified")
	} else {
		listPackageInput := compute.ListPackagesInput{}
		packages, err := tritonComputeClient.Packages().List(context.Background(), &listPackageInput)
		if err != nil {
			return []string{}, err
		}

		// Sort packages by memory size in increasing order
		sort.SliceStable(packages, func(i, j int) bool {
			return packages[i].Memory < packages[j].Memory
		})

		searcher := func(input string, index int) bool {
			pkg := packages[index]
			name := strings.Replace(strings.ToLower(pkg.Name), " ", "", -1)
			input = strings.Replace(strings.ToLower(input), " ", "", -1)

			return strings.Contains(name, input)
		}

		prompt := promptui.Select{
			Label: "Triton Machine Package to use for node",
			Items: packages,
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
			return []string{}, err
		}

		cfg.TritonMachinePackage = packages[i].Name
	}

	// Get existing node names
	nodes, err := currentState.Nodes(selectedCluster)
	if err != nil {
		return []string{}, err
	}
	existingNames := []string{}
	for nodeName := range nodes {
		existingNames = append(existingNames, nodeName)
	}

	// Determine what the hostnames should be for the new node(s)
	newHostnames := getNewHostnames(existingNames, cfg.Hostname, cfg.NodeCount)

	// Add new node to terraform config with the new hostnames
	for _, newHostname := range newHostnames {
		cfgCopy := cfg
		cfgCopy.Hostname = newHostname
		err = currentState.AddNode(selectedCluster, newHostname, cfgCopy)
		if err != nil {
			return []string{}, err
		}
	}

	return newHostnames, nil
}
