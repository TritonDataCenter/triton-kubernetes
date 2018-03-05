package create

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/joyent/triton-kubernetes/backend"
	"github.com/joyent/triton-kubernetes/shell"
	"github.com/joyent/triton-kubernetes/util"
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

	GCMPrivateNetworkName      string   `json:"gcm_private_network_name,omitempty"`
	TritonNetworkNames         []string `json:"triton_network_names,omitempty"`
	TritonImageName            string   `json:"triton_image_name,omitempty"`
	TritonImageVersion         string   `json:"triton_image_version,omitempty"`
	TritonSSHUser              string   `json:"triton_ssh_user,omitempty"`
	MasterTritonMachinePackage string   `json:"master_triton_machine_package,omitempty"`

	TritonMySQLImageName        string `json:"triton_mysql_image_name,omitempty"`
	TritonMySQLImageVersion     string `json:"triton_mysql_image_version,omitempty"`
	MySQLDBTritonMachinePackage string `json:"mysqldb_triton_machine_package,omitempty"`

	RancherAdminUsername    string `json:"rancher_admin_username,omitempty"`
	RancherAdminPassword    string `json:"rancher_admin_password,omitempty"`
	RancherServerImage      string `json:"rancher_server_image,omitempty"`
	RancherAgentImage       string `json:"rancher_agent_image,omitempty"`
	RancherRegistry         string `json:"rancher_registry,omitempty"`
	RancherRegistryUsername string `json:"rancher_registry_username,omitempty"`
	RancherRegistryPassword string `json:"rancher_registry_password,omitempty"`
}

func NewTritonManager(remoteBackend backend.Backend) error {
	nonInteractiveMode := viper.GetBool("non-interactive")
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
	} else if nonInteractiveMode {
		return errors.New("name must be specified")
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
	} else if nonInteractiveMode {
		return errors.New("ha must be specified")
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
		} else if nonInteractiveMode {
			return errors.New("gcm_node_count must be specified")
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

	// Rancher Docker Registry
	if viper.IsSet("private_registry") {
		cfg.RancherRegistry = viper.GetString("private_registry")
	} else if !nonInteractiveMode {
		prompt := promptui.Prompt{
			Label:   "Private Registry",
			Default: "None",
		}

		result, err := prompt.Run()
		if err != nil {
			return err
		}

		if result != "None" {
			cfg.RancherRegistry = result
		}
	}

	// Ask for rancher registry username/password only if rancher registry is given
	if cfg.RancherRegistry != "" {
		// Rancher Registry Username
		if viper.IsSet("private_registry_username") {
			cfg.RancherRegistryUsername = viper.GetString("private_registry_username")
		} else if nonInteractiveMode {
			return errors.New("private_registry_username must be specified")
		} else {
			prompt := promptui.Prompt{
				Label: "Private Registry Username",
			}

			result, err := prompt.Run()
			if err != nil {
				return err
			}
			cfg.RancherRegistryUsername = result
		}

		// Rancher Registry Password
		if viper.IsSet("private_registry_password") {
			cfg.RancherRegistryPassword = viper.GetString("private_registry_password")
		} else if nonInteractiveMode {
			return errors.New("private_registry_password must be specified")
		} else {
			prompt := promptui.Prompt{
				Label: "Private Registry Password",
				Mask:  '*',
			}

			result, err := prompt.Run()
			if err != nil {
				return err
			}
			cfg.RancherRegistryPassword = result
		}
	}

	// Rancher Server Image
	if viper.IsSet("rancher_server_image") {
		cfg.RancherServerImage = viper.GetString("rancher_server_image")
	} else if !nonInteractiveMode {
		prompt := promptui.Prompt{
			Label:   "Rancher Server Image",
			Default: "Default",
		}

		result, err := prompt.Run()
		if err != nil {
			return err
		}

		if result != "Default" {
			cfg.RancherServerImage = result
		}
	}

	// Rancher Agent Image
	if viper.IsSet("rancher_agent_image") {
		cfg.RancherAgentImage = viper.GetString("rancher_agent_image")
	} else if !nonInteractiveMode {
		prompt := promptui.Prompt{
			Label:   "Rancher Agent Image",
			Default: "Default",
		}

		result, err := prompt.Run()
		if err != nil {
			return err
		}

		if result != "Default" {
			cfg.RancherAgentImage = result
		}
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

	// GCM Private Network Name
	if cfg.HA {
		if viper.IsSet("gcm_private_network_name") {
			cfg.GCMPrivateNetworkName = viper.GetString("gcm_private_network_name")

			isValidName := false
			for _, validNetwork := range networks {
				if cfg.GCMPrivateNetworkName == validNetwork.Name {
					isValidName = true
					break
				}
			}
			if !isValidName {
				return fmt.Errorf("Invalid GCM private network name '%s', must be one of the following: %s", cfg.GCMPrivateNetworkName, strings.Join(validNetworksSlice, ", "))
			}
		} else {
			// GCM Private Network Prompt
			gcmNetworkPrompt := promptui.Select{
				Label: "GCM Private Network",
				Items: networks,
				Templates: &promptui.SelectTemplates{
					Label:    "{{ . }}?",
					Active:   fmt.Sprintf("%s {{ .Name | underline }}", promptui.IconSelect),
					Inactive: "  {{.Name}}",
					Selected: fmt.Sprintf(`{{ "%s" | green }} {{ "GCM Private Network:" | bold}} {{ .Name }}`, promptui.IconGood),
				},
			}
			i, _, err := gcmNetworkPrompt.Run()
			if err != nil {
				return err
			}
			cfg.GCMPrivateNetworkName = networks[i].Name
		}
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

	// Triton MySQL Image
	if cfg.HA && viper.IsSet("triton_mysql_image_name") && viper.IsSet("triton_mysql_image_version") {
		cfg.TritonMySQLImageName = viper.GetString("triton_mysql_image_name")
		cfg.TritonMySQLImageVersion = viper.GetString("triton_mysql_image_version")
		// Verify Triton MySQL image name and version
		found := false
		for _, image := range images {
			if image.Name == cfg.TritonMySQLImageName && image.Version == cfg.TritonMySQLImageVersion {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("Invalid Triton MySQL Image Name and Version '%s@%s'", cfg.TritonMySQLImageName, cfg.TritonMySQLImageVersion)
		}
	} else if cfg.HA && nonInteractiveMode {
		return errors.New("Both triton_mysql_image_name and triton_mysql_image_version must be specified when HA is selected")
	} else if cfg.HA {
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
		// Verify MySQL DB triton machine package
		found := false
		for _, pkg := range packages {
			if cfg.MySQLDBTritonMachinePackage == pkg.Name {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("Invalid MySQL DB Triton Machine Package '%s'", cfg.MySQLDBTritonMachinePackage)
		}
	} else if cfg.HA && nonInteractiveMode {
		return errors.New("mysqldb_triton_machine_package must be specified when HA is selected")
	} else if cfg.HA {
		searcher := func(input string, index int) bool {
			pkg := packages[index]
			name := strings.Replace(strings.ToLower(pkg.Name), " ", "", -1)
			input = strings.Replace(strings.ToLower(input), " ", "", -1)

			return strings.Contains(name, input)
		}

		prompt := promptui.Select{
			Label: "MySQL DB Triton Machine Package to use for Rancher Master",
			Items: packages,
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

		cfg.MySQLDBTritonMachinePackage = packages[i].Name
	}

	// Rancher Admin Username
	if viper.IsSet("rancher_admin_username") {
		cfg.RancherAdminUsername = viper.GetString("rancher_admin_username")
	} else if nonInteractiveMode {
		return errors.New("rancher_admin_username must be specified")
	} else {
		prompt := promptui.Prompt{
			Label:   "Rancher Admin Username",
			Default: "admin",
		}

		result, err := prompt.Run()
		if err != nil {
			return err
		}
		cfg.RancherAdminUsername = result
	}

	if cfg.RancherAdminUsername == "" {
		return errors.New("Invalid Rancher Admin username")
	}

	// Rancher Admin Password
	if viper.IsSet("rancher_admin_password") {
		cfg.RancherAdminPassword = viper.GetString("rancher_admin_password")
	} else if nonInteractiveMode {
		return errors.New("rancher_admin_password must be specified")
	} else {
		prompt := promptui.Prompt{
			Label: "Rancher Admin Password",
			Mask:  '*',
		}

		result, err := prompt.Run()
		if err != nil {
			return err
		}
		cfg.RancherAdminPassword = result
	}

	if cfg.RancherAdminPassword == "" {
		return errors.New("Invalid Rancher Admin password")
	}

	state, err := remoteBackend.State(cfg.Name)
	if err != nil {
		return err
	}

	state.Add("module.cluster-manager", &cfg)
	state.Add(remoteBackend.StateTerraformConfig(cfg.Name))

	if !nonInteractiveMode {
		label := "Proceed with the manager creation"
		selected := "Proceed"
		confirmed, err := util.PromptForConfirmation(label, selected)
		if err != nil {
			return err
		}
		if !confirmed {
			fmt.Println("Manager creation canceled.")
			return nil
		}
	}

	err = shell.RunTerraformApplyWithState(state)
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
