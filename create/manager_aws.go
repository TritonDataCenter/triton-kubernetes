package create

import (
	"errors"
	"fmt"
	"net"
	"os"
	"sort"
	"strings"

	"github.com/mesoform/triton-kubernetes/state"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/manifoldco/promptui"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

const (
	awsRancherTerraformModulePath = "terraform/modules/aws-rancher"
)

// This struct represents the definition of a Terraform .tf file.
// Marshalled into json this struct can be passed directly to Terraform.
type awsManagerTerraformConfig struct {
	baseManagerTerraformConfig

	AWSAccessKey string `json:"aws_access_key"`
	AWSSecretKey string `json:"aws_secret_key"`

	AWSRegion         string `json:"aws_region"`
	AWSVPCCIDR        string `json:"aws_vpc_cidr"`
	AWSSubnetCIDR     string `json:"aws_subnet_cidr"`
	AWSPublicKeyPath  string `json:"aws_public_key_path"`
	AWSPrivateKeyPath string `json:"aws_private_key_path"`
	AWSKeyName        string `json:"aws_key_name"`
	AWSSSHUser        string `json:"aws_ssh_user"`

	AWSAMIID        string `json:"aws_ami_id"`
	AWSInstanceType string `json:"aws_instance_type"`
}

func newAWSManager(currentState state.State, name string) error {
	nonInteractiveMode := viper.GetBool("non-interactive")

	baseConfig, err := getBaseManagerTerraformConfig(awsRancherTerraformModulePath, name)
	if err != nil {
		return err
	}

	cfg := awsManagerTerraformConfig{
		baseManagerTerraformConfig: baseConfig,
	}

	// AWS Access Key
	if viper.IsSet("aws_access_key") {
		cfg.AWSAccessKey = viper.GetString("aws_access_key")
	} else if nonInteractiveMode {
		return errors.New("aws_access_key must be specified")
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
	} else if nonInteractiveMode {
		return errors.New("aws_secret_key must be specified")
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
	} else if nonInteractiveMode {
		return errors.New("aws_region must be specified")
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

	// Reinit ec2 client with selected region
	awsConfig = aws.NewConfig().WithCredentials(creds).WithRegion(cfg.AWSRegion)
	sess, err = session.NewSession(awsConfig)
	if err != nil {
		return err
	}
	ec2Client = ec2.New(sess)

	// AWS Key
	// If either aws_key_name or aws_public_key_path is set use it
	// Otherwise ask the user if they'd like to upload a key or use an existing key
	if viper.IsSet("aws_key_name") {
		cfg.AWSKeyName = viper.GetString("aws_key_name")
		if viper.IsSet("aws_public_key_path") {
			expandedAWSPublicKeyPath, err := homedir.Expand(viper.GetString("aws_public_key_path"))
			if err != nil {
				return err
			}
			cfg.AWSPublicKeyPath = expandedAWSPublicKeyPath
		}
	} else if nonInteractiveMode {
		return errors.New("aws_key_name must be specified")
	} else {
		// List all available aws keys
		input := ec2.DescribeKeyPairsInput{}
		rawKeyPairs, err := ec2Client.DescribeKeyPairs(&input)
		if err != nil {
			return err
		}

		keyPairs := []string{}
		for _, key := range rawKeyPairs.KeyPairs {
			keyPairs = append(keyPairs, *key.KeyName)
		}

		// If there are no key pairs ask the user to create a new public key
		createNewKeyPair := false
		if len(keyPairs) == 0 {
			createNewKeyPair = true

			prompt := promptui.Prompt{
				Label:   "Name for new aws public key",
				Default: "triton-kubernetes_public_key",
			}

			value, err := prompt.Run()
			if err != nil {
				return err
			}

			cfg.AWSKeyName = value

		} else {
			prompt := promptui.SelectWithAdd{
				Label:    "AWS Key to use",
				Items:    keyPairs,
				AddLabel: "Upload new key",
			}

			i, value, err := prompt.Run()
			if err != nil {
				return err
			}

			// i == -1 when user selects "Upload new key"
			if i == -1 {
				createNewKeyPair = true
			}

			cfg.AWSKeyName = value
		}

		if createNewKeyPair {
			// User chose to create new key, ask for aws_public_key_path
			prompt := promptui.Prompt{
				Label: "AWS Public Key Path",
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
				Default: "~/.ssh/id_rsa.pub",
			}

			result, err := prompt.Run()
			if err != nil {
				return err
			}

			expandedKeyPath, err := homedir.Expand(result)
			if err != nil {
				return err
			}

			cfg.AWSPublicKeyPath = expandedKeyPath
		}
	}

	rawAWSPrivateKeyPath := ""
	if viper.IsSet("aws_private_key_path") {
		rawAWSPrivateKeyPath = viper.GetString("aws_private_key_path")
	} else if nonInteractiveMode {
		return errors.New("aws_private_key_path must be specified")
	} else {
		prompt := promptui.Prompt{
			Label: "AWS Private Key Path",
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
			return err
		}
		rawAWSPrivateKeyPath = result
	}

	expandedAWSPrivateKeyPath, err := homedir.Expand(rawAWSPrivateKeyPath)
	if err != nil {
		return err
	}
	cfg.AWSPrivateKeyPath = expandedAWSPrivateKeyPath

	if viper.IsSet("aws_ssh_user") {
		cfg.AWSSSHUser = viper.GetString("aws_ssh_user")
	} else if nonInteractiveMode {
		return errors.New("aws_ssh_user must be specified")
	} else {
		prompt := promptui.Prompt{
			Label:   "AWS SSH User",
			Default: "ubuntu",
		}

		result, err := prompt.Run()
		if err != nil {
			return err
		}
		cfg.AWSSSHUser = result
	}

	// AWS VPC CIDR
	if viper.IsSet("aws_vpc_cidr") {
		cfg.AWSVPCCIDR = viper.GetString("aws_vpc_cidr")
	} else if nonInteractiveMode {
		return errors.New("aws_vpc_cidr must be specified")
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
				if prefixLength < 16 {
					return fmt.Errorf("Prefix length must be 16 or greater. Found %d.", prefixLength)
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
	} else if nonInteractiveMode {
		return errors.New("aws_subnet_cidr must be specified")
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

	// AWS AMI ID
	if viper.IsSet("aws_ami_id") {
		cfg.AWSAMIID = viper.GetString("aws_ami_id")

		// TODO: Verify aws_ami_id
	} else if nonInteractiveMode {
		return errors.New("aws_ami_id must be specified")
	} else {
		// TODO: Ask the user for a search term
		describeImagesInput := ec2.DescribeImagesInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("name"),
					Values: []*string{aws.String("*hvm-ssd/ubuntu-xenial-16.04-amd64-server*")},
				},
			},
		}
		describeImagesResponse, err := ec2Client.DescribeImages(&describeImagesInput)
		if err != nil {
			return err
		}

		type ami struct {
			Name         string
			ID           string
			CreationDate string
		}
		amis := []ami{}
		for _, image := range describeImagesResponse.Images {
			amis = append(amis, ami{
				Name:         *image.Name,
				ID:           *image.ImageId,
				CreationDate: *image.CreationDate,
			})
		}

		// Sort images by creation date in reverse chronological order
		sort.SliceStable(amis, func(i, j int) bool {
			return amis[i].CreationDate > amis[j].CreationDate
		})

		searcher := func(input string, index int) bool {
			ami := amis[index]
			name := strings.Replace(strings.ToLower(ami.Name), " ", "", -1)
			input = strings.Replace(strings.ToLower(input), " ", "", -1)

			return strings.Contains(name, input)
		}

		prompt := promptui.Select{
			Label: "AWS AMI to use",
			Items: amis,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ .Name }}?",
				Active:   fmt.Sprintf(`%s {{ .Name | underline }}`, promptui.IconSelect),
				Inactive: `  {{ .Name }}`,
				Selected: fmt.Sprintf(`{{ "%s" | green }} {{ "AWS AMI:" | bold}} {{ .Name }}`, promptui.IconGood),
			},
			Searcher: searcher,
		}

		i, _, err := prompt.Run()
		if err != nil {
			return err
		}

		cfg.AWSAMIID = amis[i].ID
	}

	// AWS Instance Type
	if viper.IsSet("aws_instance_type") {
		cfg.AWSInstanceType = viper.GetString("aws_instance_type")
	} else if nonInteractiveMode {
		return errors.New("aws_instance_type must be specified")
	} else {
		// AWS doesn't have an API to get a list of available instance types
		// Ask the user to free form input it
		prompt := promptui.Prompt{
			Label: "AWS Instance Type",
			Validate: func(input string) error {
				if len(input) == 0 {
					return errors.New("Invalid Instance Type")
				}
				return nil
			},
			Default: "t2.micro",
		}

		result, err := prompt.Run()
		if err != nil {
			return err
		}
		cfg.AWSInstanceType = result
	}

	currentState.SetManager(&cfg)

	return nil
}
