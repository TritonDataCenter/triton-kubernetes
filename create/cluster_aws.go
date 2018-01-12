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

// This struct represents the definition of a Terraform .tf file.
// Marshalled into json this struct can be passed directly to Terraform.
type awsClusterTerraformConfig struct {
	baseClusterTerraformConfig

	AWSAccessKey     string `json:"aws_access_key"`
	AWSSecretKey     string `json:"aws_secret_key"`
	AWSPublicKeyPath string `json:"aws_public_key_path"`
	AWSKeyName       string `json:"aws_key_name"`

	AWSRegion     string `json:"aws_region"`
	AWSAMIID      string `json:"aws_ami_id"`
	AWSVPCCIDR    string `json:"aws_vpc_cidr"`
	AWSSubnetCIDR string `json:"aws_subnet_cidr"`
}

func newAWSCluster(selectedClusterManager string, remoteClusterManagerState remote.RemoteClusterManagerStateManta) error {
	baseConfig, err := getBaseClusterTerraformConfig(awsRancherKubernetesTerraformModulePath)
	if err != nil {
		return err
	}

	cfg := awsClusterTerraformConfig{
		baseClusterTerraformConfig: baseConfig,
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
