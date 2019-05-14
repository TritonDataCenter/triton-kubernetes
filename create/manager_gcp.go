package create

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"github.com/mesoform/triton-kubernetes/state"
	"github.com/manifoldco/promptui"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
)

const (
	gcpRancherTerraformModulePath = "terraform/modules/gcp-rancher"
)

// This struct represents the definition of a Terraform .tf file.
// Marshalled into json this struct can be passed directly to Terraform.
type gcpManagerTerraformConfig struct {
	baseManagerTerraformConfig

	GCPPathToCredentials string `json:"gcp_path_to_credentials"`
	GCPProjectID         string `json:"gcp_project_id"`
	GCPComputeRegion     string `json:"gcp_compute_region"`

	GCPMachineType  string `json:"gcp_machine_type"`
	GCPInstanceZone string `json:"gcp_instance_zone"`
	GCPImage        string `json:"gcp_image"`

	GCPPublicKeyPath  string `json:"gcp_public_key_path"`
	GCPPrivateKeyPath string `json:"gcp_private_key_path"`
	GCPSSHUser        string `json:"gcp_ssh_user"`
}

func newGCPManager(currentState state.State, name string) error {
	nonInteractiveMode := viper.GetBool("non-interactive")

	baseConfig, err := getBaseManagerTerraformConfig(gcpRancherTerraformModulePath, name)
	if err != nil {
		return err
	}

	cfg := gcpManagerTerraformConfig{
		baseManagerTerraformConfig: baseConfig,
	}

	// GCP path_to_credentials
	rawGCPPathToCredentials := ""
	if viper.IsSet("gcp_path_to_credentials") {
		rawGCPPathToCredentials = viper.GetString("gcp_path_to_credentials")
	} else if nonInteractiveMode {
		return errors.New("gcp_path_to_credentials must be specified")
	} else {
		prompt := promptui.Prompt{
			Label: "Path to Google Cloud Platform Credentials File",
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
		}

		result, err := prompt.Run()
		if err != nil {
			return err
		}
		rawGCPPathToCredentials = result
	}

	expandedGCPPathToCredentials, err := homedir.Expand(rawGCPPathToCredentials)
	if err != nil {
		return err
	}
	cfg.GCPPathToCredentials = expandedGCPPathToCredentials

	gcpCredentials, err := ioutil.ReadFile(cfg.GCPPathToCredentials)
	if err != nil {
		return err
	}

	jwtCfg, err := google.JWTConfigFromJSON(gcpCredentials, "https://www.googleapis.com/auth/compute.readonly")
	if err != nil {
		return err
	}

	// jwt.Config does not expose the project ID, so re-unmarshal to get it.
	var pid struct {
		ProjectID string `json:"project_id"`
	}
	if err := json.Unmarshal(gcpCredentials, &pid); err != nil {
		return err
	}
	cfg.GCPProjectID = pid.ProjectID

	service, err := compute.New(jwtCfg.Client(context.Background()))
	if err != nil {
		return err
	}

	regions, err := service.Regions.List(cfg.GCPProjectID).Do()
	if err != nil {
		return err
	}

	// GCP Compute Region
	if viper.IsSet("gcp_compute_region") {
		cfg.GCPComputeRegion = viper.GetString("gcp_compute_region")

		found := false
		for _, region := range regions.Items {
			if region.Name == cfg.GCPComputeRegion {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("Selected GCP Compute Region '%s' does not exist.", cfg.GCPComputeRegion)
		}

	} else if nonInteractiveMode {
		return errors.New("gcp_compute_region must be specified")
	} else {
		searcher := func(input string, index int) bool {
			region := regions.Items[index]
			name := strings.Replace(strings.ToLower(region.Name), " ", "", -1)
			input = strings.Replace(strings.ToLower(input), " ", "", -1)

			return strings.Contains(name, input)
		}

		prompt := promptui.Select{
			Label: "GCP Compute Region",
			Items: regions.Items,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ .Name }}?",
				Active:   fmt.Sprintf(`%s {{ .Name | underline }}`, promptui.IconSelect),
				Inactive: `  {{ .Name }}`,
				Selected: fmt.Sprintf(`{{ "%s" | green }} {{ "GCP Compute Region:" | bold}} {{ .Name }}`, promptui.IconGood),
			},
			Searcher: searcher,
		}

		i, _, err := prompt.Run()
		if err != nil {
			return err
		}

		cfg.GCPComputeRegion = regions.Items[i].Name
	}

	zones, err := service.Zones.List(cfg.GCPProjectID).Filter(fmt.Sprintf("region eq https://www.googleapis.com/compute/v1/projects/%s/regions/%s", cfg.GCPProjectID, cfg.GCPComputeRegion)).Do()
	if err != nil {
		return err
	}

	// GCP Instance Zone
	if viper.IsSet("gcp_instance_zone") {
		cfg.GCPInstanceZone = viper.GetString("gcp_instance_zone")

		found := false
		for _, zone := range zones.Items {
			if zone.Name == cfg.GCPInstanceZone {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("Selected GCP Instance Zone '%s' does not exist.", cfg.GCPInstanceZone)
		}

	} else if nonInteractiveMode {
		return errors.New("gcp_instance_zone must be specified")
	} else {
		searcher := func(input string, index int) bool {
			zone := zones.Items[index]
			name := strings.Replace(strings.ToLower(zone.Name), " ", "", -1)
			input = strings.Replace(strings.ToLower(input), " ", "", -1)

			return strings.Contains(name, input)
		}

		prompt := promptui.Select{
			Label: "GCP Instance Zone",
			Items: zones.Items,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ .Name }}?",
				Active:   fmt.Sprintf(`%s {{ .Name | underline }}`, promptui.IconSelect),
				Inactive: `  {{ .Name }}`,
				Selected: fmt.Sprintf(`{{ "%s" | green }} {{ "GCP Instance Zone:" | bold}} {{ .Name }}`, promptui.IconGood),
			},
			Searcher: searcher,
		}

		i, _, err := prompt.Run()
		if err != nil {
			return err
		}

		cfg.GCPInstanceZone = zones.Items[i].Name
	}

	machineTypes, err := service.MachineTypes.List(cfg.GCPProjectID, cfg.GCPInstanceZone).Do()
	if err != nil {
		return err
	}

	// GCP Machine Type
	if viper.IsSet("gcp_machine_type") {
		cfg.GCPMachineType = viper.GetString("gcp_machine_type")

		found := false
		for _, machineType := range machineTypes.Items {
			if machineType.Name == cfg.GCPMachineType {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("Selected GCP Machine Type '%s' does not exist.", cfg.GCPMachineType)
		}

	} else if nonInteractiveMode {
		return errors.New("gcp_machine_type must be specified")
	} else {
		searcher := func(input string, index int) bool {
			machineType := machineTypes.Items[index]
			name := strings.Replace(strings.ToLower(machineType.Name), " ", "", -1)
			input = strings.Replace(strings.ToLower(input), " ", "", -1)

			return strings.Contains(name, input)
		}

		prompt := promptui.Select{
			Label: "GCP Machine Type",
			Items: machineTypes.Items,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ .Name }}?",
				Active:   fmt.Sprintf(`%s {{ .Name | underline }}`, promptui.IconSelect),
				Inactive: `  {{ .Name }}`,
				Selected: fmt.Sprintf(`{{ "%s" | green }} {{ "GCP Machine Type:" | bold}} {{ .Name }}`, promptui.IconGood),
			},
			Searcher: searcher,
		}

		i, _, err := prompt.Run()
		if err != nil {
			return err
		}

		cfg.GCPMachineType = machineTypes.Items[i].Name
	}

	images, err := service.Images.List("ubuntu-os-cloud").Do()
	if err != nil {
		return err
	}

	// Sort images by created timestamp in reverse chronological order
	sort.SliceStable(images.Items, func(i, j int) bool {
		return images.Items[i].CreationTimestamp > images.Items[j].CreationTimestamp
	})

	// GCP Image
	if viper.IsSet("gcp_image") {
		cfg.GCPImage = viper.GetString("gcp_image")

		found := false
		for _, image := range images.Items {
			if image.Name == cfg.GCPImage {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("Selected GCP Image '%s' does not exist.", cfg.GCPImage)
		}

	} else if nonInteractiveMode {
		return errors.New("gcp_image must be specified")
	} else {
		searcher := func(input string, index int) bool {
			image := images.Items[index]
			name := strings.Replace(strings.ToLower(image.Name), " ", "", -1)
			input = strings.Replace(strings.ToLower(input), " ", "", -1)

			return strings.Contains(name, input)
		}

		prompt := promptui.Select{
			Label: "GCP Image",
			Items: images.Items,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ .Name }}?",
				Active:   fmt.Sprintf(`%s {{ .Name | underline }}`, promptui.IconSelect),
				Inactive: `  {{ .Name }}`,
				Selected: fmt.Sprintf(`{{ "%s" | green }} {{ "GCP Image:" | bold}} {{ .Name }}`, promptui.IconGood),
			},
			Searcher: searcher,
		}

		i, _, err := prompt.Run()
		if err != nil {
			return err
		}

		cfg.GCPImage = images.Items[i].Name
	}

	rawGCPPublicKeyPath := ""
	if viper.IsSet("gcp_public_key_path") {
		rawGCPPublicKeyPath = viper.GetString("gcp_public_key_path")
	} else if nonInteractiveMode {
		return errors.New("gcp_public_key_path must be specified")
	} else {
		prompt := promptui.Prompt{
			Label: "GCP Public Key Path",
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
			Default: "~/.ssh/id_rsa.pub",
		}

		result, err := prompt.Run()
		if err != nil {
			return err
		}
		rawGCPPublicKeyPath = result
	}

	expandedGCPPublicKeyPath, err := homedir.Expand(rawGCPPublicKeyPath)
	if err != nil {
		return err
	}
	cfg.GCPPublicKeyPath = expandedGCPPublicKeyPath

	rawGCPPrivateKeyPath := ""
	if viper.IsSet("gcp_private_key_path") {
		rawGCPPrivateKeyPath = viper.GetString("gcp_private_key_path")
	} else if nonInteractiveMode {
		return errors.New("gcp_private_key_path must be specified")
	} else {
		prompt := promptui.Prompt{
			Label: "GCP Private Key Path",
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
		rawGCPPrivateKeyPath = result
	}

	expandedGCPPrivateKeyPath, err := homedir.Expand(rawGCPPrivateKeyPath)
	if err != nil {
		return err
	}
	cfg.GCPPrivateKeyPath = expandedGCPPrivateKeyPath

	if viper.IsSet("gcp_ssh_user") {
		cfg.GCPSSHUser = viper.GetString("gcp_ssh_user")
	} else if nonInteractiveMode {
		return errors.New("gcp_ssh_user must be specified")
	} else {
		prompt := promptui.Prompt{
			Label:   "GCP SSH User",
			Default: "ubuntu",
		}

		result, err := prompt.Run()
		if err != nil {
			return err
		}
		cfg.GCPSSHUser = result
	}

	currentState.SetManager(&cfg)

	return nil
}
