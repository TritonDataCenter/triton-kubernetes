package create

import (
	// "bytes"
	// "context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"

	"github.com/joyent/triton-kubernetes/remote"
	"github.com/joyent/triton-kubernetes/shell"

	"github.com/Jeffail/gabs"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

const (
	awsClusterKeyFormat                     = "cluster_aws_%s"
	awsRancherKubernetesTerraformModulePath = "terraform/modules/aws-rancher-k8s"
)

type awsClusterTerraformConfig struct {
	Source string `json:"source"`

	Name string `json:"name"`

	EtcdNodeCount          string `json:"etcd_node_count"`
	OrchestrationNodeCount string `json:"orchestration_node_count"`
	ComputeNodeCount       string `json:"compute_node_count"`

	KubernetesPlaneIsolation string `json:"k8s_plane_isolation"`

	RancherAPIURL    string `json:"rancher_api_url"`
	RancherAccessKey string `json:"rancher_access_key"`
	RancherSecretKey string `json:"rancher_secret_key"`

	AWSAccessKey     string `json:"aws_access_key"`
	AWSSecretKey     string `json:"aws_secret_key"`
	AWSPublicKeyPath string `json:"aws_public_key_path"`
	AWSKeyName       string `json:"aws_key_name"`

	AWSRegion     string `json:"aws_region"`
	AWSAMIID      string `json:"aws_ami_id"`
	AWSVPCCIDR    string `json:"aws_vpc_cidr"`
	AWSSubnetCIDR string `json:"aws_subnet_cidr"`

	RancherRegistry         string `json:"rancher_registry,omitempty"`
	RancherRegistryUsername string `json:"rancher_registry_username,omitempty"`
	RancherRegistryPassword string `json:"rancher_registry_password,omitempty"`

	KubernetesRegistry         string `json:"k8s_registry,omitempty"`
	KubernetesRegistryUsername string `json:"k8s_registry_username,omitempty"`
	KubernetesRegistryPassword string `json:"k8s_registry_password,omitempty"`
}

func newAWSCluster(selectedClusterManager string, remoteClusterManagerState remote.RemoteClusterManagerStateManta, tritonAccount, tritonKeyPath, tritonKeyID, tritonURL, mantaURL string) error {
	cfg := awsClusterTerraformConfig{
		RancherAPIURL: "http://${element(module.cluster-manager.masters, 0)}:8080",

		// Set node counts to 0 since we manage nodes individually in triton-kubernetes cli
		EtcdNodeCount:          "0",
		OrchestrationNodeCount: "0",
		ComputeNodeCount:       "0",
	}

	baseSource := "github.com/joyent/triton-kubernetes"
	if viper.IsSet("source_url") {
		baseSource = viper.GetString("source_url")
	}

	// Module Source location e.g. github.com/joyent/triton-kubernetes//terraform/modules/triton-rancher-k8s
	cfg.Source = fmt.Sprintf("%s//%s", baseSource, awsRancherKubernetesTerraformModulePath)

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
		return errors.New("Invalid Cluster Name")
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

	// Verify selected plane isolation is valid
	if cfg.KubernetesPlaneIsolation != "required" && cfg.KubernetesPlaneIsolation != "none" {
		return fmt.Errorf("Invalid k8s_plane_isolation '%s', must be 'required' or 'none'", cfg.KubernetesPlaneIsolation)
	}

	// AWS Access Key
	if viper.IsSet("aws_access_key") {
		cfg.AWSAccessKey = viper.GetString("aws_access_key")
	} else {
		prompt := promptui.Prompt{
			Label: "AWS Access Key",
			Validate: func(input string) error {
				if len(input) == 0 {
					return errors.New("Invalid access key")
				}
				return nil
			},
		}

		result, err := prompt.Run()
		if err != nil {
			return err
		}
		cfg.AWSAccessKey = result
	}

	// AWS Secret Key
	if viper.IsSet("aws_secret_key") {
		cfg.AWSSecretKey = viper.GetString("aws_secret_key")
	} else {
		prompt := promptui.Prompt{
			Label: "AWS Secret Key",
			Validate: func(input string) error {
				if len(input) == 0 {
					return errors.New("Invalid secret key")
				}
				return nil
			},
		}

		result, err := prompt.Run()
		if err != nil {
			return err
		}
		cfg.AWSSecretKey = result
	}

	// AWS Public Key Path
	if viper.IsSet("aws_public_key_path") {
		cfg.AWSPublicKeyPath = viper.GetString("aws_public_key_path")
	} else {
		prompt := promptui.Prompt{
			Label: "AWS Public Key Path",
			Validate: func(input string) error {
				_, err := os.Stat(input)
				if err != nil {
					if os.IsNotExist(err) {
						return errors.New("File not found")
					}
				}
				return nil
			},
			Default: "~/.ssh/id_rsa.pub",
		}

		result, err := prompt.Run()
		if err != nil {
			return err
		}
		cfg.AWSPublicKeyPath = result
	}

	// AWS Key Name
	if viper.IsSet("aws_key_name") {
		cfg.AWSKeyName = viper.GetString("aws_key_name")
	} else {
		prompt := promptui.Prompt{
			Label:   "AWS Key Name",
			Default: "rancher_public_key",
		}

		result, err := prompt.Run()
		if err != nil {
			return err
		}
		cfg.AWSKeyName = result
	}

	// We now have enough information to init an aws client
	creds := credentials.NewStaticCredentials(cfg.AWSAccessKey, cfg.AWSSecretKey, "")

	// Using us-west-1 region by default. The configuration needs a region set to
	// get all regions available to the aws user.
	awsConfig := aws.NewConfig().WithCredentials(creds).WithRegion(endpoints.UsWest1RegionID)
	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return err
	}
	ec2Client := ec2.New(sess)

	// Get the regions
	regionsResult, err := ec2Client.DescribeRegions(&ec2.DescribeRegionsInput{})
	if err != nil {
		return err
	}
	regions := regionsResult.Regions

	// AWS Region
	if viper.IsSet("aws_region") {
		cfg.AWSRegion = viper.GetString("aws_region")
		// Validate the AWS Region
		found := false
		for _, region := range regions {
			name := region.RegionName
			if name != nil && *name == cfg.AWSRegion {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("Selected AWS Region '%s' does not exist.", cfg.AWSRegion)
		}
	} else {
		// Building an array of strings that will be given to the SelectPrompt.
		// The SelectTemplate has problems displaying struct fields that are string pointers.
		regionNames := []string{}
		for _, region := range regions {
			name := ""
			if region.RegionName != nil {
				name = *region.RegionName
			}
			regionNames = append(regionNames, name)
		}

		searcher := func(input string, index int) bool {
			regionName := regionNames[index]
			name := strings.Replace(strings.ToLower(regionName), " ", "", -1)
			input = strings.Replace(strings.ToLower(input), " ", "", -1)

			return strings.Contains(name, input)
		}

		prompt := promptui.Select{
			Label: "AWS Region",
			Items: regionNames,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}?",
				Active:   fmt.Sprintf("%s {{ . | underline }}", promptui.IconSelect),
				Inactive: " {{.}}",
				Selected: fmt.Sprintf(`{{ "%s" | green }} {{ "AWS Region:" | bold}} {{ . }}`, promptui.IconGood),
			},
			Searcher: searcher,
		}

		i, _, err := prompt.Run()
		if err != nil {
			return err
		}

		cfg.AWSRegion = *regions[i].RegionName
	}

	// AWS VPC CIDR
	if viper.IsSet("aws_vpc_cidr") {
		cfg.AWSVPCCIDR = viper.GetString("aws_vpc_cidr")
	} else {
		prompt := promptui.Prompt{
			Label: "AWS VPC CIDR",
			Validate: func(input string) error {
				_, ipNet, err := net.ParseCIDR(input)
				if err != nil {
					return err
				}
				if ipNet == nil {
					return fmt.Errorf("Invalid CIDR address: %s", input)
				}
				prefixLength, _ := ipNet.Mask.Size()
				if prefixLength > 16 {
					return fmt.Errorf("Prefix length must be 16 or less. Found %d.", prefixLength)
				}
				return nil
			},
			Default: "10.0.0.0/16",
		}

		result, err := prompt.Run()
		if err != nil {
			return err
		}
		cfg.AWSVPCCIDR = result
	}

	// AWS Subnet CIDR
	if viper.IsSet("aws_subnet_cidr") {
		cfg.AWSSubnetCIDR = viper.GetString("aws_subnet_cidr")
	} else {
		// Parsing VPC CIDR to prepare for subnet validation
		_, vpcIPNet, err := net.ParseCIDR(cfg.AWSVPCCIDR)
		if err != nil {
			return err
		}
		vpcPrefix, _ := vpcIPNet.Mask.Size()

		prompt := promptui.Prompt{
			Label: "AWS Subnet CIDR",
			Validate: func(input string) error {
				// Check for valid CIDR format
				ip, ipNet, err := net.ParseCIDR(input)
				if err != nil {
					return err
				}
				if ipNet == nil {
					return fmt.Errorf("Invalid CIDR address: %s", input)
				}

				// Check if VPC contains subnet
				prefix, _ := ipNet.Mask.Size()
				if !vpcIPNet.Contains(ip) || prefix < vpcPrefix {
					return fmt.Errorf("Subnet CIDR '%s' is not within bounds of VPC CIDR '%s'.", input, cfg.AWSVPCCIDR)
				}
				return nil
			},
			Default: "10.0.2.0/24",
		}

		result, err := prompt.Run()
		if err != nil {
			return err
		}
		cfg.AWSSubnetCIDR = result
	}

	// Rancher Registry
	if viper.IsSet("rancher_registry") {
		cfg.RancherRegistry = viper.GetString("rancher_registry")
	} else {
		prompt := promptui.Prompt{
			Label: "Rancher Registry",
		}

		result, err := prompt.Run()
		if err != nil {
			return err
		}
		cfg.RancherRegistry = result
	}

	// Ask for rancher registry username/password only if rancher registry is given
	if cfg.RancherRegistry != "" {
		// Rancher Registry Username
		if viper.IsSet("rancher_registry_username") {
			cfg.RancherRegistryUsername = viper.GetString("rancher_registry_username")
		} else {
			prompt := promptui.Prompt{
				Label: "Rancher Registry Username",
			}

			result, err := prompt.Run()
			if err != nil {
				return err
			}
			cfg.RancherRegistryUsername = result
		}

		// Rancher Registry Password
		if viper.IsSet("rancher_registry_password") {
			cfg.RancherRegistryPassword = viper.GetString("rancher_registry_password")
		} else {
			prompt := promptui.Prompt{
				Label: "Rancher Registry Password",
			}

			result, err := prompt.Run()
			if err != nil {
				return err
			}
			cfg.RancherRegistryPassword = result
		}
	}

	// Kubernetes Registry
	if viper.IsSet("kubernetes_registry") {
		cfg.KubernetesRegistry = viper.GetString("kubernetes_registry")
	} else {
		prompt := promptui.Prompt{
			Label: "Kubernetes Registry",
		}

		result, err := prompt.Run()
		if err != nil {
			return err
		}
		cfg.KubernetesRegistry = result
	}

	// Ask for kubernetes registry username/password only if kubernetes registry is given
	if cfg.KubernetesRegistry != "" {
		// Kubernetes Registry Username
		if viper.IsSet("kubernetes_registry_username") {
			cfg.KubernetesRegistryUsername = viper.GetString("kubernetes_registry_username")
		} else {
			prompt := promptui.Prompt{
				Label: "Kubernetes Registry Username",
			}

			result, err := prompt.Run()
			if err != nil {
				return err
			}
			cfg.KubernetesRegistryUsername = result
		}

		// Kubernetes Registry Password
		if viper.IsSet("kubernetes_registry_password") {
			cfg.KubernetesRegistryPassword = viper.GetString("kubernetes_registry_password")
		} else {
			prompt := promptui.Prompt{
				Label: "Kubernetes Registry Password",
			}

			result, err := prompt.Run()
			if err != nil {
				return err
			}
			cfg.KubernetesRegistryPassword = result
		}
	}

	// Load current cluster manager config
	clusterManagerTerraformConfigBytes, err := remoteClusterManagerState.GetTerraformConfig(selectedClusterManager)
	if err != nil {
		return err
	}

	clusterManagerTerraformConfig, err := gabs.ParseJSON(clusterManagerTerraformConfigBytes)
	if err != nil {
		return err
	}

	// Add new cluster to terraform config
	clusterKey := fmt.Sprintf(awsClusterKeyFormat, cfg.Name)
	clusterManagerTerraformConfig.SetP(&cfg, fmt.Sprintf("module.%s", clusterKey))

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
