package create

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

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
	cfg.Source = "github.com/joyent/triton-kubernetes//terraform/modules/triton-rancher-k8s"

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
		// ssh-keygen -E md5 -lf PATH_TO_FILE
		// Sample output:
		// 2048 MD5:68:9f:9a:c4:76:3a:f4:62:77:47:3e:47:d4:34:4a:b7 njalali@Nimas-MacBook-Pro.local (RSA)
		out, err := exec.Command("ssh-keygen", "-E", "md5", "-lf", cfg.TritonKeyPath).Output()
		if err != nil {
			return err
		}

		parts := strings.Split(string(out), " ")
		if len(parts) != 4 {
			return errors.New("Could not get ssh key id")
		}

		cfg.TritonKeyID = strings.TrimPrefix(parts[1], "MD5:")
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

	jsonBytes, err := json.MarshalIndent(&cfg, "", "\t")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonBytes))

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
