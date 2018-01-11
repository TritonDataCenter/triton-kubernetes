package util

import (
	"errors"
	"os"

	"github.com/joyent/triton-kubernetes/shell"

	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

func GetTritonAccountVariables() (tritonAccount, tritonKeyPath, tritonKeyID, tritonURL, mantaURL string, funcErr error) {
	// Triton Account
	if viper.IsSet("triton_account") {
		tritonAccount = viper.GetString("triton_account")
	} else {
		prompt := promptui.Prompt{
			Label: "Triton Account Name (usually your email)",
			Validate: func(input string) error {
				if len(input) == 0 {
					return errors.New("Invalid Triton Account")
				}
				return nil
			},
		}

		result, err := prompt.Run()
		if err != nil {
			funcErr = err
			return
		}
		tritonAccount = result
	}

	// Triton Key Path
	if viper.IsSet("triton_key_path") {
		tritonKeyPath = viper.GetString("triton_key_path")
	} else {
		prompt := promptui.Prompt{
			Label: "Triton Key Path",
			Validate: func(input string) error {
				_, err := os.Stat(input)
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
			funcErr = err
			return
		}
		tritonKeyPath = result
	}

	// Triton Key ID
	if viper.IsSet("triton_key_id") {
		tritonKeyID = viper.GetString("triton_key_id")
	} else {
		keyID, err := shell.GetPublicKeyFingerprintFromPrivateKey(tritonKeyPath)
		if err != nil {
			funcErr = err
			return
		}
		tritonKeyID = keyID
	}

	// Triton URL
	if viper.IsSet("triton_url") {
		tritonURL = viper.GetString("triton_url")
	} else {
		prompt := promptui.Prompt{
			Label:   "Triton URL",
			Default: "https://us-east-1.api.joyent.com",
		}

		result, err := prompt.Run()
		if err != nil {
			funcErr = err
			return
		}
		tritonURL = result
	}

	// Manta URL
	if viper.IsSet("manta_url") {
		mantaURL = viper.GetString("manta_url")
	} else {
		prompt := promptui.Prompt{
			Label:   "Manta URL",
			Default: "https://us-east.manta.joyent.com",
		}

		result, err := prompt.Run()
		if err != nil {
			funcErr = err
			return
		}
		mantaURL = result
	}

	return
}
