package create

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/joyent/triton-kubernetes/backend"
	"github.com/joyent/triton-kubernetes/state"

	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
)

const (
	gcpNodeKeyFormat                            = "module.node_gcp_%s"
	gcpRancherKubernetesHostTerraformModulePath = "terraform/modules/gcp-rancher-k8s-host"
)

type gcpNodeTerraformConfig struct {
	baseNodeTerraformConfig

	GCPPathToCredentials string `json:"gcp_path_to_credentials"`
	GCPProjectID         string `json:"gcp_project_id"`
	GCPComputeRegion     string `json:"gcp_compute_region"`

	GCPComputeNetworkName string `json:"gcp_compute_network_name"`

	GCPMachineType  string `json:"gcp_machine_type"`
	GCPInstanceZone string `json:"gcp_instance_zone"`
	GCPImage        string `json:"gcp_image"`
}

// Adds new GCP nodes to the given cluster and manager.
// Returns:
// - a slice of the hostnames added
// - the new state
// - error or nil
func newGCPNode(selectedClusterManager, selectedCluster string, remoteBackend backend.Backend, currentState state.State) ([]string, error) {
	silentMode := viper.GetBool("silent")
	baseConfig, err := getBaseNodeTerraformConfig(gcpRancherKubernetesHostTerraformModulePath, selectedCluster, currentState)
	if err != nil {
		return []string{}, err
	}

	cfg := gcpNodeTerraformConfig{
		baseNodeTerraformConfig: baseConfig,

		// Grab variables from cluster config
		GCPPathToCredentials: currentState.Get(fmt.Sprintf("module.%s.gcp_path_to_credentials", selectedCluster)),
		GCPProjectID:         currentState.Get(fmt.Sprintf("module.%s.gcp_project_id", selectedCluster)),
		GCPComputeRegion:     currentState.Get(fmt.Sprintf("module.%s.gcp_compute_region", selectedCluster)),

		// Reference terraform output variables from cluster module
		GCPComputeNetworkName: fmt.Sprintf("${module.%s.gcp_compute_network_name}", selectedCluster),
	}

	gcpCredentials, err := ioutil.ReadFile(cfg.GCPPathToCredentials)
	if err != nil {
		return []string{}, err
	}

	jwtCfg, err := google.JWTConfigFromJSON(gcpCredentials, "https://www.googleapis.com/auth/compute.readonly")
	if err != nil {
		return []string{}, err
	}

	service, err := compute.New(jwtCfg.Client(context.Background()))
	if err != nil {
		return []string{}, err
	}

	zones, err := service.Zones.List(cfg.GCPProjectID).Filter(fmt.Sprintf("region eq https://www.googleapis.com/compute/v1/projects/%s/regions/%s", cfg.GCPProjectID, cfg.GCPComputeRegion)).Do()
	if err != nil {
		return []string{}, err
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
			return []string{}, fmt.Errorf("Selected GCP Instance Zone '%s' does not exist.", cfg.GCPInstanceZone)
		}

	} else if silentMode {
		return []string{}, errors.New("gcp_instance_zone must be specified")
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
			return []string{}, err
		}

		cfg.GCPInstanceZone = zones.Items[i].Name
	}

	machineTypes, err := service.MachineTypes.List(cfg.GCPProjectID, cfg.GCPInstanceZone).Do()
	if err != nil {
		return []string{}, err
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
			return []string{}, fmt.Errorf("Selected GCP Machine Type '%s' does not exist.", cfg.GCPMachineType)
		}

	} else if silentMode {
		return []string{}, errors.New("gcp_machine_type must be specified")
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
			return []string{}, err
		}

		cfg.GCPMachineType = machineTypes.Items[i].Name
	}

	images, err := service.Images.List("ubuntu-os-cloud").Do()
	if err != nil {
		return []string{}, err
	}

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
			return []string{}, fmt.Errorf("Selected GCP Image '%s' does not exist.", cfg.GCPImage)
		}

	} else if silentMode {
		return []string{}, errors.New("gcp_image must be specified")
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
			return []string{}, err
		}

		cfg.GCPImage = images.Items[i].Name
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
		err = currentState.Add(fmt.Sprintf(gcpNodeKeyFormat, newHostname), cfgCopy)
		if err != nil {
			return []string{}, err
		}
	}

	return newHostnames, nil
}
