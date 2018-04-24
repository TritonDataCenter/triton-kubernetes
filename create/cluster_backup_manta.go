package create

import (
	"errors"
	"fmt"
	"os"

	"github.com/joyent/triton-kubernetes/state"
	"github.com/joyent/triton-kubernetes/util"
	"github.com/manifoldco/promptui"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

const (
	clusterBackupMantaTerraformModulePath = "terraform/modules/k8s-backup-manta"
)

type mantaClusterBackupTerraformConfig struct {
	baseClusterBackupTerraformConfig

	TritonAccount string `json:"triton_account,omitempty"`
	TritonKeyID   string `json:"triton_key_id,omitempty"`
	MantaSubuser  string `json:"manta_subuser,omitempty"`
}

func newMantaClusterBackup(selectedClusterKey string, currentState state.State) (string, error) {
	nonInteractiveMode := viper.GetBool("non-interactive")
	baseConfig, err := getBaseClusterBackupTerraformConfig(clusterBackupMantaTerraformModulePath, selectedClusterKey)
	if err != nil {
		return "", err
	}

	cfg := mantaClusterBackupTerraformConfig{
		baseClusterBackupTerraformConfig: baseConfig,
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

	// Manta Subuser
	if viper.IsSet("manta_subuser") {
		cfg.MantaSubuser = viper.GetString("manta_subuser")
	} else if nonInteractiveMode {
		return "", errors.New("manta_subuser must be specified")
	} else {
		prompt := promptui.Prompt{
			Label: "Manta Subuser",
			Validate: func(input string) error {
				if len(input) == 0 {
					return errors.New("Invalid Manta Subuser")
				}
				return nil
			},
		}

		result, err := prompt.Run()
		if err != nil {
			return "", err
		}
		cfg.MantaSubuser = result
	}

	clusterBackupKey := fmt.Sprintf(clusterBackupKeyFormat, selectedClusterKey)
	err = currentState.Add(clusterBackupKey, cfg)
	if err != nil {
		return "", err
	}

	return clusterBackupKey, nil
}
