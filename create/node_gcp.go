package create

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"

	"github.com/joyent/triton-kubernetes/backend"
	"github.com/joyent/triton-kubernetes/state"

	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
)

const (
	gcpRancherKubernetesHostTerraformModulePath = "terraform/modules/gcp-rancher-k8s-host"
)

type gcpNodeTerraformConfig struct {
	baseNodeTerraformConfig

	GCPPathToCredentials string `json:"gcp_path_to_credentials"`
	GCPProjectID         string `json:"gcp_project_id"`
	GCPComputeRegion     string `json:"gcp_compute_region"`

	GCPComputeNetworkName     string `json:"gcp_compute_network_name"`
	GCPComputeFirewallHostTag string `json:"gcp_compute_firewall_host_tag"`

	GCPMachineType  string `json:"gcp_machine_type"`
	GCPInstanceZone string `json:"gcp_instance_zone"`
	GCPImage        string `json:"gcp_image"`

	GCPDiskType      string `json:"gcp_disk_type"`
	GCPDiskSize      string `json:"gcp_disk_size"`
	GCPDiskMountPath string `json:"gcp_disk_mount_path"`
}

// Adds new GCP nodes to the given cluster and manager.
// Returns:
// - a slice of the hostnames added
// - the new state
// - error or nil
func newGCPNode(selectedClusterManager, selectedCluster string, remoteBackend backend.Backend, currentState state.State) ([]string, error) {
	nonInteractiveMode := viper.GetBool("non-interactive")
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
		GCPComputeNetworkName:     fmt.Sprintf("${module.%s.gcp_compute_network_name}", selectedCluster),
		GCPComputeFirewallHostTag: fmt.Sprintf("${module.%s.gcp_compute_firewall_host_tag}", selectedCluster),
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

	} else if nonInteractiveMode {
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

	} else if nonInteractiveMode {
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
			return []string{}, fmt.Errorf("Selected GCP Image '%s' does not exist.", cfg.GCPImage)
		}

	} else if nonInteractiveMode {
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

	// Get list of GCP permanent disk types
	// diskTypesResponse, err := service.DiskTypes.List(cfg.GCPProjectID, cfg.GCPInstanceZone).Do()
	// if err != nil {
	// 	return nil, err
	// }
	// permanentDiskTypes := []*compute.DiskType{}
	// for _, diskType := range diskTypesResponse.Items {
	// 	if strings.HasPrefix(diskType.Name, "pd-") {
	// 		permanentDiskTypes = append(permanentDiskTypes, diskType)
	// 	}
	// }

	// // GCP Disk
	// diskTypeIsSet := viper.IsSet("gcp_disk_type")
	// diskSizeIsSet := viper.IsSet("gcp_disk_size")
	// diskMountPathIsSet := viper.IsSet("gcp_disk_mount_path")
	// if nonInteractiveMode {
	// 	// If disk type is defined, assume the user's intent is to add
	// 	// a disk to the host and throw an error if neither size nor mount path is set.
	// 	if diskTypeIsSet && !(diskSizeIsSet || diskMountPathIsSet) {
	// 		return nil, errors.New("If gcp_disk_type is set, gcp_disk_size and gcp_disk_mount must also be set.")
	// 	} else if diskTypeIsSet {
	// 		cfg.GCPDiskType = viper.GetString("gcp_disk_type")
	// 		if !isValidDiskType(permanentDiskTypes, cfg.GCPDiskType) {
	// 			return nil, fmt.Errorf("gcp_disk_type must be valid. Found '%s'.", cfg.GCPDiskType)
	// 		}
	// 		cfg.GCPDiskSize = viper.GetString("gcp_disk_size")
	// 		cfg.GCPDiskMountPath = viper.GetString("gcp_disk_mount_path")
	// 	}
	// } else {
	// 	shouldCreateDisk, err := util.PromptForConfirmation("Create a disk for this node", "Disk created")
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	if shouldCreateDisk {
	// 		// GCP Disk Type
	// 		if diskTypeIsSet {
	// 			cfg.GCPDiskType = viper.GetString("gcp_disk_type")
	// 			if !isValidDiskType(permanentDiskTypes, cfg.GCPDiskType) {
	// 				return nil, fmt.Errorf("gcp_disk_type must be valid. Found '%s'.", cfg.GCPDiskType)
	// 			}
	// 		} else {
	// 			prompt := promptui.Select{
	// 				Label: "GCP Disk Type",
	// 				Items: permanentDiskTypes,
	// 				Templates: &promptui.SelectTemplates{
	// 					Label:    "{{ . }}?",
	// 					Active:   fmt.Sprintf("%s {{ .Name | underline }}", promptui.IconSelect),
	// 					Inactive: "  {{.Name}}",
	// 					Selected: "  GCP Disk Type? {{.Name}}",
	// 				},
	// 			}

	// 			i, _, err := prompt.Run()
	// 			if err != nil {
	// 				return nil, err
	// 			}

	// 			cfg.GCPDiskType = permanentDiskTypes[i].Name
	// 		}

	// 		// GCP Disk Size
	// 		if diskSizeIsSet {
	// 			cfg.GCPDiskSize = viper.GetString("gcp_disk_size")
	// 		} else {
	// 			prompt := promptui.Prompt{
	// 				Label: "GCP Disk Size in GB",
	// 				Validate: func(input string) error {
	// 					num, err := strconv.ParseInt(input, 10, 64)
	// 					if err != nil {
	// 						return errors.New("Invalid number")
	// 					}
	// 					if num <= 0 {
	// 						return errors.New("Number must be greater than 0")
	// 					}
	// 					return nil
	// 				},
	// 			}
	// 			result, err := prompt.Run()
	// 			if err != nil {
	// 				return nil, err
	// 			}
	// 			cfg.GCPDiskSize = result
	// 		}

	// 		// GCP Disk Mount path
	// 		if diskMountPathIsSet {
	// 			cfg.GCPDiskMountPath = viper.GetString("gcp_disk_mount_path")
	// 		} else {
	// 			prompt := promptui.Prompt{
	// 				Label: "GCP Disk Mount Path",
	// 			}

	// 			result, err := prompt.Run()
	// 			if err != nil {
	// 				return nil, err
	// 			}
	// 			cfg.GCPDiskMountPath = result
	// 		}
	// 	}
	// }

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

func isValidDiskType(validDiskTypes []*compute.DiskType, diskTypeName string) bool {
	for _, diskType := range validDiskTypes {
		if diskType.Name == diskTypeName {
			return true
		}
	}
	return false
}
