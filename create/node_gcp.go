package create

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/joyent/triton-kubernetes/remote"
	"github.com/joyent/triton-kubernetes/shell"

	"github.com/Jeffail/gabs"
	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
)

const (
	gcpNodeKeyFormat                            = "node_gcp_%s"
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

func newGCPNode(selectedClusterManager, selectedCluster string, remoteClusterManagerState remote.RemoteClusterManagerStateManta, clusterManagerTerraformConfig *gabs.Container) error {
	baseConfig, err := getBaseNodeTerraformConfig(gcpRancherKubernetesHostTerraformModulePath, selectedCluster, clusterManagerTerraformConfig)
	if err != nil {
		return err
	}

	cfg := gcpNodeTerraformConfig{
		baseNodeTerraformConfig: baseConfig,

		// Grab variables from cluster config
		GCPPathToCredentials: clusterManagerTerraformConfig.Path(fmt.Sprintf("module.%s.gcp_path_to_credentials", selectedCluster)).Data().(string),
		GCPProjectID:         clusterManagerTerraformConfig.Path(fmt.Sprintf("module.%s.gcp_project_id", selectedCluster)).Data().(string),
		GCPComputeRegion:     clusterManagerTerraformConfig.Path(fmt.Sprintf("module.%s.gcp_compute_region", selectedCluster)).Data().(string),

		// Reference terraform output variables from cluster module
		GCPComputeNetworkName: fmt.Sprintf("${module.%s.gcp_compute_network_name}", selectedCluster),
	}

	gcpCredentials, err := ioutil.ReadFile(cfg.GCPPathToCredentials)
	if err != nil {
		return err
	}

	jwtCfg, err := google.JWTConfigFromJSON(gcpCredentials, "https://www.googleapis.com/auth/compute.readonly")
	if err != nil {
		return err
	}

	service, err := compute.New(jwtCfg.Client(context.Background()))
	if err != nil {
		return err
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

	// Add new node to terraform config
	nodeKey := fmt.Sprintf(gcpNodeKeyFormat, cfg.Hostname)
	clusterManagerTerraformConfig.SetP(&cfg, fmt.Sprintf("module.%s", nodeKey))

	// Create a temporary directory
	tempDir, err := ioutil.TempDir("", "triton-kubernetes-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	// Save the terraform config to the temporary directory
	jsonBytes := []byte(clusterManagerTerraformConfig.StringIndent("", "\t"))
	jsonPath := fmt.Sprintf("%s/%s", tempDir, "main.tf.json")
	err = ioutil.WriteFile(jsonPath, jsonBytes, 0644)
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
	err = remoteClusterManagerState.CommitTerraformConfig(selectedClusterManager, jsonBytes)
	if err != nil {
		return err
	}

	return nil
}
