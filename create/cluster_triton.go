package create

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/Jeffail/gabs"
	"github.com/joyent/triton-kubernetes/shell"
	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

type tritonClusterTerraformConfig struct {
	Source string `json:"source"`

	Name string `json:"name"`

	EtcdNodeCount          string `json:"etcd_node_count"`
	OrchestrationNodeCount string `json:"orchestration_node_count"`
	ComputeNodeCount       string `json:"compute_node_count"`

	KubernetesPlaneIsolation string `json:"k8s_plane_isolation"`

	RancherAPIURL    string `json:"rancher_api_url"`
	RancherAccessKey string `json:"rancher_access_key"`
	RancherSecretKey string `json:"rancher_secret_key"`

	TritonAccount string `json:"triton_account"`
	TritonKeyPath string `json:"triton_key_path"`
	TritonKeyID   string `json:"triton_key_id"`
	TritonURL     string `json:"triton_url,omitempty"`

	RancherRegistry         string `json:"rancher_registry,omitempty"`
	RancherRegistryUsername string `json:"rancher_registry_username,omitempty"`
	RancherRegistryPassword string `json:"rancher_registry_password,omitempty"`

	KubernetesRegistry         string `json:"k8s_registry,omitempty"`
	KubernetesRegistryUsername string `json:"k8s_registry_username,omitempty"`
	KubernetesRegistryPassword string `json:"k8s_registry_password,omitempty"`
}

func NewTritonCluster() error {
	cfg := tritonClusterTerraformConfig{}

	// TODO: Move this to const or make configurable
	// cfg.Source = "github.com/joyent/triton-kubernetes//terraform/modules/triton-rancher-k8s"
	cfg.Source = "./terraform/modules/triton-rancher-k8s"

	// Set node counts to 0 since we manage nodes individually in triton-kubernetes cli
	cfg.EtcdNodeCount = "0"
	cfg.OrchestrationNodeCount = "0"
	cfg.ComputeNodeCount = "0"

	// Name
	if viper.IsSet("name") {
		cfg.Name = viper.GetString("name")
	} else {
		prompt := promptui.Prompt{
			Label: "Cluster Name",
		}

		result, err := prompt.Run()
		if err != nil {
			return err
		}
		cfg.Name = result
	}

	if cfg.Name == "" {
		return errors.New("Invalid Name")
	}

	// Kubernetes Plane Isolation
	if viper.IsSet("k8s_plane_isolation") {
		cfg.KubernetesPlaneIsolation = viper.GetString("k8s_plane_isolation")
	} else {
		prompt := promptui.Select{
			Label: "Kubernetes Plane Isolation",
			Items: []string{"required", "none"},
		}

		_, value, err := prompt.Run()
		if err != nil {
			return err
		}

		cfg.KubernetesPlaneIsolation = value
	}

	// Rancher API URL
	cfg.RancherAPIURL = "http://${element(module.cluster-manager.masters, 0)}:8080"

	// Triton Account
	if viper.IsSet("triton_account") {
		cfg.TritonAccount = viper.GetString("triton_account")
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
			return err
		}
		cfg.TritonAccount = result
	}

	// Triton Key Path
	if viper.IsSet("triton_key_path") {
		cfg.TritonKeyPath = viper.GetString("triton_key_path")
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
			return err
		}
		cfg.TritonKeyPath = result
	}

	// Triton Key ID
	if viper.IsSet("triton_key_id") {
		cfg.TritonKeyID = viper.GetString("triton_key_id")
	} else {
		keyID, err := shell.GetPublicKeyFingerprintFromPrivateKey(cfg.TritonKeyPath)
		if err != nil {
			return err
		}
		cfg.TritonKeyID = keyID
	}

	// Triton URL
	if viper.IsSet("triton_url") {
		cfg.TritonURL = viper.GetString("triton_url")
	} else {
		prompt := promptui.Prompt{
			Label:   "Triton URL",
			Default: "https://us-east-1.api.joyent.com",
		}

		result, err := prompt.Run()
		if err != nil {
			return err
		}
		cfg.TritonURL = result
	}

	// Load current tf config
	currentConfigBytes, err := ioutil.ReadFile("main.tf.json")
	if err != nil {
		return err
	}

	parsedConfig, err := gabs.ParseJSON(currentConfigBytes)
	if err != nil {
		return err
	}

	parsedConfig.SetP(&cfg, fmt.Sprintf("module.%s", cfg.Name))

	jsonBytes := []byte(parsedConfig.StringIndent("", "\t"))

	err = ioutil.WriteFile("main.tf.json", jsonBytes, 0644)
	if err != nil {
		return err
	}

	// Run terraform init
	tfInit := exec.Command("terraform", []string{"init", "-force-copy"}...)
	tfInit.Stdin = os.Stdin
	tfInit.Stdout = os.Stdout
	tfInit.Stderr = os.Stderr

	if err := tfInit.Start(); err != nil {
		return err
	}

	err = tfInit.Wait()
	if err != nil {
		return err
	}

	// Run terraform apply
	tfApply := exec.Command("terraform", []string{"apply", "-auto-approve"}...)
	tfApply.Stdin = os.Stdin
	tfApply.Stdout = os.Stdout
	tfApply.Stderr = os.Stderr

	if err := tfApply.Start(); err != nil {
		return err
	}

	err = tfApply.Wait()
	if err != nil {
		return err
	}

	return nil
}
