package util

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/joyent/triton-kubernetes/backend"
	"github.com/joyent/triton-kubernetes/backend/local"
	"github.com/joyent/triton-kubernetes/backend/manta"

	"github.com/manifoldco/promptui"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

func PromptForBackend() (backend.Backend, error) {
	// Ask user what backend to use
	selectedBackendProvider := ""
	if viper.IsSet("backend_provider") {
		selectedBackendProvider = viper.GetString("backend_provider")
	} else {
		prompt := promptui.Select{
			Label: "Backend to persist data",
			Items: []string{"Local", "Manta"},
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}?",
				Active:   fmt.Sprintf(`%s {{ . | underline }}`, promptui.IconSelect),
				Inactive: `  {{ . }}`,
				Selected: fmt.Sprintf(`{{ "%s" | green }} {{ "Backend Provider:" | bold}} {{ . }}`, promptui.IconGood),
			},
		}

		_, value, err := prompt.Run()
		if err != nil {
			return nil, err
		}

		selectedBackendProvider = strings.ToLower(value)
	}

	switch selectedBackendProvider {
	case "local":
		return local.New()
	case "manta":
		// Triton Account
		tritonAccount := ""
		if viper.IsSet("triton_account") {
			tritonAccount = viper.GetString("triton_account")
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
				return nil, err
			}
			tritonAccount = result
		}

		// Triton Key Path
		rawTritonKeyPath := ""
		if viper.IsSet("triton_key_path") {
			rawTritonKeyPath = viper.GetString("triton_key_path")
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
				return nil, err
			}
			rawTritonKeyPath = result
		}

		expandedTritonKeyPath, err := homedir.Expand(rawTritonKeyPath)
		if err != nil {
			return nil, err
		}
		tritonKeyPath := expandedTritonKeyPath

		// Triton Key ID
		tritonKeyID := ""
		if viper.IsSet("triton_key_id") {
			tritonKeyID = viper.GetString("triton_key_id")
		} else {
			keyID, err := GetPublicKeyFingerprintFromPrivateKey(tritonKeyPath)
			if err != nil {
				return nil, err
			}
			tritonKeyID = keyID
		}

		// Triton URL
		tritonURL := ""
		if viper.IsSet("triton_url") {
			tritonURL = viper.GetString("triton_url")
		} else {
			prompt := promptui.Prompt{
				Label:   "Triton URL",
				Default: "https://us-east-1.api.joyent.com",
			}

			result, err := prompt.Run()
			if err != nil {
				return nil, err
			}
			tritonURL = result
		}

		// Manta URL
		mantaURL := ""
		if viper.IsSet("manta_url") {
			mantaURL = viper.GetString("manta_url")
		} else {
			prompt := promptui.Prompt{
				Label:   "Manta URL",
				Default: "https://us-east.manta.joyent.com",
			}

			result, err := prompt.Run()
			if err != nil {
				return nil, err
			}
			mantaURL = result
		}

		return manta.New(tritonAccount, tritonKeyPath, tritonKeyID, tritonURL, mantaURL)
	}

	return nil, fmt.Errorf("Unsupported backend provider '%s'", selectedBackendProvider)
}
