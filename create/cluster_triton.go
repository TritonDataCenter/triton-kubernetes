package create

import (
	"errors"
	"os"

	"github.com/mesoform/triton-kubernetes/backend"
	"github.com/mesoform/triton-kubernetes/state"
	"github.com/mesoform/triton-kubernetes/util"
	"github.com/manifoldco/promptui"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

const (
	tritonRancherKubernetesTerraformModulePath = "terraform/modules/triton-rancher-k8s"
)

// This struct represents the definition of a Terraform .tf file.
// Marshalled into json this struct can be passed directly to Terraform.
type tritonClusterTerraformConfig struct {
	baseClusterTerraformConfig

	TritonAccount string `json:"triton_account"`
	TritonKeyPath string `json:"triton_key_path"`
	TritonKeyID   string `json:"triton_key_id"`
	TritonURL     string `json:"triton_url,omitempty"`
}

// Returns the name of the cluster that was created and the new state.
func newTritonCluster(remoteBackend backend.Backend, currentState state.State) (string, error) {
	nonInteractiveMode := viper.GetBool("non-interactive")
	baseConfig, err := getBaseClusterTerraformConfig(tritonRancherKubernetesTerraformModulePath)
	if err != nil {
		return "", err
	}

	cfg := tritonClusterTerraformConfig{
		baseClusterTerraformConfig: baseConfig,
	}

	// Triton Account
	if viper.IsSet("triton_account") {
		cfg.TritonAccount = viper.GetString("triton_account")
	} else if nonInteractiveMode {
		return "", errors.New("triton_account must be specified")
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
			return "", err
		}
		cfg.TritonAccount = result
	}

	// Triton Key Path
	rawTritonKeyPath := ""
	if viper.IsSet("triton_key_path") {
		rawTritonKeyPath = viper.GetString("triton_key_path")
	} else if nonInteractiveMode {
		return "", errors.New("triton_key_path must be specified")
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
			return "", err
		}
		rawTritonKeyPath = result
	}

	expandedTritonKeyPath, err := homedir.Expand(rawTritonKeyPath)
	if err != nil {
		return "", err
	}
	cfg.TritonKeyPath = expandedTritonKeyPath

	// Triton Key ID
	if viper.IsSet("triton_key_id") {
		cfg.TritonKeyID = viper.GetString("triton_key_id")
	} else {
		keyID, err := util.GetPublicKeyFingerprintFromPrivateKey(cfg.TritonKeyPath)
		if err != nil {
			return "", err
		}
		cfg.TritonKeyID = keyID
	}

	// Triton URL
	if viper.IsSet("triton_url") {
		cfg.TritonURL = viper.GetString("triton_url")
	} else if nonInteractiveMode {
		return "", errors.New("triton_url must be specified")
	} else {
		prompt := promptui.Prompt{
			Label:   "Triton URL",
			Default: "https://us-east-1.api.joyent.com",
		}

		result, err := prompt.Run()
		if err != nil {
			return "", err
		}
		cfg.TritonURL = result
	}

	// Add new cluster to terraform config
	err = currentState.AddCluster("triton", cfg.Name, &cfg)
	if err != nil {
		return "", err
	}

	return cfg.Name, nil
}
