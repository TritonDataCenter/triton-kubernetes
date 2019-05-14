package create

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/mesoform/triton-kubernetes/backend"
	"github.com/mesoform/triton-kubernetes/state"

	"github.com/manifoldco/promptui"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
)

const (
	gcpRancherKubernetesTerraformModulePath = "terraform/modules/gcp-rancher-k8s"
)

// This struct represents the definition of a Terraform .tf file.
// Marshalled into json this struct can be passed directly to Terraform.
type gcpClusterTerraformConfig struct {
	baseClusterTerraformConfig

	GCPPathToCredentials string `json:"gcp_path_to_credentials"`
	GCPProjectID         string `json:"gcp_project_id"`
	GCPComputeRegion     string `json:"gcp_compute_region"`
}

// Returns the name of the cluster that was created and the new state.
func newGCPCluster(remoteBackend backend.Backend, currentState state.State) (string, error) {
	nonInteractiveMode := viper.GetBool("non-interactive")
	baseConfig, err := getBaseClusterTerraformConfig(gcpRancherKubernetesTerraformModulePath)
	if err != nil {
		return "", err
	}

	cfg := gcpClusterTerraformConfig{
		baseClusterTerraformConfig: baseConfig,
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

	jwtCfg, err := google.JWTConfigFromJSON(gcpCredentials, "https://www.googleapis.com/auth/compute.readonly")
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
		cfg.GCPComputeRegion = viper.GetString("gcp_compute_region")

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

	// Add new cluster to terraform config
	err = currentState.AddCluster("gcp", cfg.Name, &cfg)
	if err != nil {
		return "", err
	}

	return cfg.Name, nil
}
