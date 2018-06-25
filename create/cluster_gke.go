package create

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/joyent/triton-kubernetes/backend"
	"github.com/joyent/triton-kubernetes/state"

	"github.com/manifoldco/promptui"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
	container "google.golang.org/api/container/v1beta1"
)

const (
	gkeRancherKubernetesTerraformModulePath = "terraform/modules/gke-rancher-k8s"
)

// This struct represents the definition of a Terraform .tf file.
// Marshalled into json this struct can be passed directly to Terraform.
type gkeClusterTerraformConfig struct {
	Source string `json:"source"`

	Name string `json:"name"`

	RancherAPIURL    string `json:"rancher_api_url"`
	RancherAccessKey string `json:"rancher_access_key"`
	RancherSecretKey string `json:"rancher_secret_key"`

	GCPPathToCredentials string   `json:"gcp_path_to_credentials"`
	GCPProjectID         string   `json:"gcp_project_id"`
	GCPComputeRegion     string   `json:"gcp_compute_region"`
	GCPZone              string   `json:"gcp_zone"`
	GCPAdditionalZones   []string `json:"gcp_additional_zones"`
	GCPMachineType       string   `json:"gcp_machine_type"`
	KubernetesVersion    string   `json:"k8s_version"`
	NodeCount            int      `json:"node_count"`
	Password             string   `json:"password"`
}

// Returns the name of the cluster that was created and the new state.
func newGKECluster(remoteBackend backend.Backend, currentState state.State) (string, error) {
	nonInteractiveMode := viper.GetBool("non-interactive")
	cfg := gkeClusterTerraformConfig{
		RancherAPIURL:    "${module.cluster-manager.rancher_url}",
		RancherAccessKey: "${module.cluster-manager.rancher_access_key}",
		RancherSecretKey: "${module.cluster-manager.rancher_secret_key}",
	}

	baseSource := defaultSourceURL
	if viper.IsSet("source_url") {
		baseSource = viper.GetString("source_url")
	}

	baseSourceRef := defaultSourceRef
	if viper.IsSet("source_ref") {
		baseSourceRef = viper.GetString("source_ref")
	}

	// Module Source location e.g. github.com/joyent/triton-kubernetes//terraform/modules/azure-rancher-k8s?ref=master
	cfg.Source = fmt.Sprintf("%s//%s?ref=%s", baseSource, gkeRancherKubernetesTerraformModulePath, baseSourceRef)

	// Name
	clusterNameRegexp := regexp.MustCompile("^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$")
	if viper.IsSet("name") {
		cfg.Name = viper.GetString("name")
	} else if nonInteractiveMode {
		return "", errors.New("name must be specified")
	} else {
		prompt := promptui.Prompt{
			Label: "Cluster Name",
			Validate: func(input string) error {
				if !clusterNameRegexp.MatchString(input) {
					return errors.New("A DNS-1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character")
				}

				return nil
			},
		}

		result, err := prompt.Run()
		if err != nil {
			return "", err
		}
		cfg.Name = result
	}

	if cfg.Name == "" || !clusterNameRegexp.MatchString(cfg.Name) {
		return "", errors.New("Invalid Cluster Name")
	}

	// GCP path_to_credentials
	rawGCPPathToCredentials := ""
	if viper.IsSet("gcp_path_to_credentials") {
		rawGCPPathToCredentials = viper.GetString("gcp_path_to_credentials")
	} else if nonInteractiveMode {
		return "", errors.New("gcp_path_to_credentials must be specified")
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
			return "", err
		}
		rawGCPPathToCredentials = result
	}

	expandedGCPPathToCredentials, err := homedir.Expand(rawGCPPathToCredentials)
	if err != nil {
		return "", err
	}
	cfg.GCPPathToCredentials = expandedGCPPathToCredentials

	gcpCredentials, err := ioutil.ReadFile(cfg.GCPPathToCredentials)
	if err != nil {
		return "", err
	}

	jwtCfg, err := google.JWTConfigFromJSON(gcpCredentials, "https://www.googleapis.com/auth/compute.readonly", "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return "", err
	}

	// jwt.Config does not expose the project ID, so re-unmarshal to get it.
	var pid struct {
		ProjectID string `json:"project_id"`
	}
	if err := json.Unmarshal(gcpCredentials, &pid); err != nil {
		return "", err
	}
	cfg.GCPProjectID = pid.ProjectID

	service, err := compute.New(jwtCfg.Client(context.Background()))
	if err != nil {
		return "", err
	}

	regions, err := service.Regions.List(cfg.GCPProjectID).Do()
	if err != nil {
		return "", err
	}

	// GCP Compute Region
	if viper.IsSet("gcp_compute_region") {
		cfg.GCPComputeRegion = viper.GetString("gcp_region")

		found := false
		for _, region := range regions.Items {
			if region.Name == cfg.GCPComputeRegion {
				found = true
				break
			}
		}
		if !found {
			return "", fmt.Errorf("Selected GCP Compute Region '%s' does not exist.", cfg.GCPComputeRegion)
		}

	} else if nonInteractiveMode {
		return "", errors.New("gcp_compute_region must be specified")
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
			return "", err
		}

		cfg.GCPComputeRegion = regions.Items[i].Name
	}

	zones, err := service.Zones.List(cfg.GCPProjectID).Filter(fmt.Sprintf("region eq https://www.googleapis.com/compute/v1/projects/%s/regions/%s", cfg.GCPProjectID, cfg.GCPComputeRegion)).Do()
	if err != nil {
		return "", err
	}

	// GCP Zone
	if viper.IsSet("gcp_zone") {
		cfg.GCPZone = viper.GetString("gcp_zone")

		found := false
		for _, zone := range zones.Items {
			if zone.Name == cfg.GCPZone {
				found = true
				break
			}
		}
		if !found {
			return "", fmt.Errorf("Selected GCP Zone '%s' does not exist.", cfg.GCPZone)
		}

	} else if nonInteractiveMode {
		return "", errors.New("gcp_zone must be specified")
	} else {
		searcher := func(input string, index int) bool {
			zone := zones.Items[index]
			name := strings.Replace(strings.ToLower(zone.Name), " ", "", -1)
			input = strings.Replace(strings.ToLower(input), " ", "", -1)

			return strings.Contains(name, input)
		}

		prompt := promptui.Select{
			Label: "GCP Zone",
			Items: zones.Items,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ .Name }}?",
				Active:   fmt.Sprintf(`%s {{ .Name | underline }}`, promptui.IconSelect),
				Inactive: `  {{ .Name }}`,
				Selected: fmt.Sprintf(`{{ "%s" | green }} {{ "GCP Zone:" | bold}} {{ .Name }}`, promptui.IconGood),
			},
			Searcher: searcher,
		}

		i, _, err := prompt.Run()
		if err != nil {
			return "", err
		}

		cfg.GCPZone = zones.Items[i].Name
	}

	// GCP Additional Zones
	if viper.IsSet("gcp_additional_zones") {
		cfg.GCPAdditionalZones = viper.GetStringSlice("gcp_additional_zones")

		for _, givenZone := range cfg.GCPAdditionalZones {
			found := false
			for _, zone := range zones.Items {
				if zone.Name == givenZone {
					found = true
					break
				}
			}
			if !found {
				return "", fmt.Errorf("One of the GCP Additional Zones '%s' does not exist.", givenZone)
			}
		}

	} else if nonInteractiveMode {
		return "", errors.New("gcp_additional_zones must be specified")
	} else {
		unselectedZones := zones.Items
		for i, zone := range unselectedZones {
			if zone.Name == cfg.GCPZone {
				// remove primary zone that's already selected
				unselectedZones = append(unselectedZones[:i], unselectedZones[i+1:]...)
				break
			}
		}

		searcher := func(input string, index int) bool {
			zone := unselectedZones[index]
			name := strings.Replace(strings.ToLower(zone.Name), " ", "", -1)
			input = strings.Replace(strings.ToLower(input), " ", "", -1)

			return strings.Contains(name, input)
		}

		prompt := promptui.Select{
			Label: "GCP Additional Zones",
			Items: unselectedZones,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ .Name }}?",
				Active:   fmt.Sprintf(`%s {{ .Name | underline }}`, promptui.IconSelect),
				Inactive: `  {{ .Name }}`,
				Selected: fmt.Sprintf(`{{ "%s" | green }} {{ "GCP Additional Zones:" | bold}} {{ .Name }}`, promptui.IconGood),
			},
			Searcher: searcher,
		}

		continueOptions := []struct {
			Name  string
			Value bool
		}{
			{"Yes", true},
			{"No", false},
		}

		continuePrompt := promptui.Select{
			Label: "Add another",
			Items: continueOptions,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}?",
				Active:   fmt.Sprintf("%s {{ .Name | underline }}", promptui.IconSelect),
				Inactive: "  {{.Name}}",
				Selected: "  Add another? {{.Name}}",
			},
		}

		zonesChosen := []string{}
		shouldPrompt := true
		for shouldPrompt {
			i, _, err := prompt.Run()
			if err != nil {
				return "", err
			}
			zonesChosen = append(zonesChosen, zones.Items[i].Name)

			// Remove the chosen zone from the list of choices
			unselectedZones = append(unselectedZones[:i], unselectedZones[i+1:]...)

			if len(unselectedZones) == 0 {
				shouldPrompt = false
			} else {
				prompt.Items = unselectedZones

				// Continue Prompt
				i, _, err = continuePrompt.Run()
				if err != nil {
					return "", err
				}
				shouldPrompt = continueOptions[i].Value
			}
		}

		cfg.GCPAdditionalZones = zonesChosen
	}

	machineTypes, err := service.MachineTypes.List(cfg.GCPProjectID, cfg.GCPZone).Do()
	if err != nil {
		return "", err
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
			return "", fmt.Errorf("Selected GCP Machine Type '%s' does not exist.", cfg.GCPMachineType)
		}

	} else if nonInteractiveMode {
		return "", errors.New("gcp_machine_type must be specified")
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
			return "", err
		}

		cfg.GCPMachineType = machineTypes.Items[i].Name
	}

	containerService, err := container.New(jwtCfg.Client(context.Background()))
	if err != nil {
		return "", err
	}

	projectsZonesService := container.NewProjectsZonesService(containerService)
	serverConfig, err := projectsZonesService.GetServerconfig(cfg.GCPProjectID, cfg.GCPZone).Do()
	if err != nil {
		return "", err
	}

	// Kubernetes Version
	if viper.IsSet("k8s_version") {
		cfg.KubernetesVersion = viper.GetString("k8s_version")
	} else if nonInteractiveMode {
		return "", errors.New("k8s_version must be specified")
	} else {

		prompt := promptui.Select{
			Label: "Kubernetes Version",
			Items: serverConfig.ValidMasterVersions,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}?",
				Active:   fmt.Sprintf(`%s {{ . | underline }}`, promptui.IconSelect),
				Inactive: `  {{ . }}`,
				Selected: fmt.Sprintf(`{{ "%s" | green }} {{ "Kubernetes Version:" | bold}} {{ . }}`, promptui.IconGood),
			},
		}

		i, _, err := prompt.Run()
		if err != nil {
			return "", err
		}

		cfg.KubernetesVersion = serverConfig.ValidMasterVersions[i]
	}

	// Allow user to specify number of nodes to be created.
	var countInput string
	if viper.IsSet("node_count") {
		countInput = viper.GetString("node_count")
	} else {
		prompt := promptui.Prompt{
			Label: "Number of nodes to create",
			Validate: func(input string) error {
				num, err := strconv.ParseInt(input, 10, 64)
				if err != nil {
					return errors.New("Invalid number")
				}
				if num <= 0 {
					return errors.New("Number must be greater than 0")
				}
				return nil
			},
			Default: "3",
		}
		result, err := prompt.Run()
		if err != nil {
			return "", err
		}
		countInput = result
	}

	// Verifying node count
	nodeCount, err := strconv.Atoi(countInput)
	if err != nil {
		return "", fmt.Errorf("node_count must be a valid number. Found '%s'.", countInput)
	}
	if nodeCount <= 0 {
		return "", fmt.Errorf("node_count must be greater than 0. Found '%d'.", nodeCount)
	}

	cfg.NodeCount = nodeCount

	// Password
	if viper.IsSet("password") {
		cfg.Password = viper.GetString("password")
	} else if nonInteractiveMode {
		return "", errors.New("password must be specified")
	} else {
		prompt := promptui.Prompt{
			Label: "Kubernetes Master Password",
			Mask:  '*',
			Validate: func(input string) error {
				if len(input) < 16 {
					return errors.New("password must be at least 16 characters in length")
				}
				return nil
			},
		}

		result, err := prompt.Run()
		if err != nil {
			return "", err
		}
		cfg.Password = result
	}

	// Add new cluster to terraform config
	err = currentState.AddCluster("gke", cfg.Name, &cfg)
	if err != nil {
		return "", err
	}

	return cfg.Name, nil
}
