package create

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/joyent/triton-kubernetes/backend"
	"github.com/joyent/triton-kubernetes/state"
	"github.com/joyent/triton-kubernetes/util"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

const (
	awsRancherKubernetesHostTerraformModulePath = "terraform/modules/aws-rancher-k8s-host"
)

// List of valid EBS Volume types
var ebsVolumeTypes = []struct {
	Key         string
	Name        string
	DefaultSize string
}{
	{"standard", "Magnetic", "100"},
	{"gp2", "General Purpose SSD", "100"},
	{"io1", "Provisioned IOPS SSD", "100"},
	{"st1", "Throughput Optimised HDD", "500"},
	{"sc1", "Cold HDD", "500"},
}

type awsNodeTerraformConfig struct {
	baseNodeTerraformConfig

	AWSAccessKey string `json:"aws_access_key"`
	AWSSecretKey string `json:"aws_secret_key"`

	AWSRegion          string `json:"aws_region"`
	AWSSubnetID        string `json:"aws_subnet_id"`
	AWSSecurityGroupID string `json:"aws_security_group_id"`
	AWSKeyName         string `json:"aws_key_name"`

	AWSAMIID        string `json:"aws_ami_id"`
	AWSInstanceType string `json:"aws_instance_type"`

	EBSVolumeDeviceName string `json:"ebs_volume_device_name,omitempty"`
	EBSVolumeMountPath  string `json:"ebs_volume_mount_path,omitempty"`
	EBSVolumeType       string `json:"ebs_volume_type,omitempty"`
	EBSVolumeIOPS       string `json:"ebs_volume_iops,omitempty"`
	EBSVolumeSize       string `json:"ebs_volume_size,omitempty"`
}

// Adds new AWS nodes to the given cluster and manager.
// Returns:
// - a slice of the hostnames added
// - the new state
// - error or nil
func newAWSNode(selectedClusterManager, selectedCluster string, remoteBackend backend.Backend, currentState state.State) ([]string, error) {
	nonInteractiveMode := viper.GetBool("non-interactive")
	baseConfig, err := getBaseNodeTerraformConfig(awsRancherKubernetesHostTerraformModulePath, selectedCluster, currentState)
	if err != nil {
		return []string{}, err
	}

	cfg := awsNodeTerraformConfig{
		baseNodeTerraformConfig: baseConfig,

		// Grab variables from cluster config
		AWSAccessKey: currentState.Get(fmt.Sprintf("module.%s.aws_access_key", selectedCluster)),
		AWSSecretKey: currentState.Get(fmt.Sprintf("module.%s.aws_secret_key", selectedCluster)),
		AWSRegion:    currentState.Get(fmt.Sprintf("module.%s.aws_region", selectedCluster)),

		// Reference terraform output variables from cluster module
		AWSSubnetID:        fmt.Sprintf("${module.%s.aws_subnet_id}", selectedCluster),
		AWSSecurityGroupID: fmt.Sprintf("${module.%s.aws_security_group_id}", selectedCluster),
		AWSKeyName:         fmt.Sprintf("${module.%s.aws_key_name}", selectedCluster),
	}

	creds := credentials.NewStaticCredentials(cfg.AWSAccessKey, cfg.AWSSecretKey, "")

	awsConfig := aws.NewConfig().WithCredentials(creds).WithRegion(cfg.AWSRegion)
	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return []string{}, err
	}
	ec2Client := ec2.New(sess)

	// AWS AMI ID
	if viper.IsSet("aws_ami_id") {
		cfg.AWSAMIID = viper.GetString("aws_ami_id")

		// TODO: Verify aws_ami_id
	} else if nonInteractiveMode {
		return []string{}, errors.New("aws_ami_id must be specified")
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
			return []string{}, err
		}

		type ami struct {
			Name string
			ID   string
		}
		amis := []ami{}
		for _, image := range describeImagesResponse.Images {
			amis = append(amis, ami{
				Name: *image.Name,
				ID:   *image.ImageId,
			})
		}

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
			return []string{}, err
		}

		cfg.AWSAMIID = amis[i].ID
	}

	// AWS Instance Type
	if viper.IsSet("aws_instance_type") {
		cfg.AWSInstanceType = viper.GetString("aws_instance_type")
	} else if nonInteractiveMode {
		return []string{}, errors.New("aws_instance_type must be specified")
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
			return []string{}, err
		}
		cfg.AWSInstanceType = result
	}

	// EBS Volume
	deviceNameIsSet := viper.IsSet("ebs_volume_device_name")
	mountPathIsSet := viper.IsSet("ebs_volume_mount_path")
	volumeSizeIsSet := viper.IsSet("ebs_volume_size")
	volumeTypeIsSet := viper.IsSet("ebs_volume_type")
	if nonInteractiveMode && deviceNameIsSet {
		cfg.EBSVolumeDeviceName = viper.GetString("ebs_volume_device_name")
		// Volume Type
		if volumeTypeIsSet {
			cfg.EBSVolumeType = viper.GetString("ebs_volume_type")
		}

		// Validating Volume Type
		typeIsValid := false
		for _, volumeType := range ebsVolumeTypes {
			if volumeType.Key == cfg.EBSVolumeType {
				typeIsValid = true
				break
			}
		}
		if !typeIsValid {
			return nil, fmt.Errorf("ebs_volume_type must be a valid volume type. Found '%s'.", cfg.EBSVolumeType)
		}

		if volumeSizeIsSet {
			cfg.EBSVolumeSize = viper.GetString("ebs_volume_size")
		} else {
			// If volume size is not defined, use the default value
			for _, volumeType := range ebsVolumeTypes {
				if volumeType.Key == cfg.EBSVolumeType {
					cfg.EBSVolumeSize = volumeType.DefaultSize
					break
				}
			}
		}

	} else {
		shouldCreateVolume, err := util.PromptForConfirmation("Create a volume for this node", "Volume Created")
		if err != nil {
			return nil, err
		}

		if shouldCreateVolume {
			// EBS device name
			if deviceNameIsSet {
				cfg.EBSVolumeDeviceName = viper.GetString("ebs_volume_device_name")
			} else {
				prompt := promptui.Prompt{
					Label: "EBS Volume Device Name",
					Validate: func(input string) error {
						r, err := regexp.Compile("^/dev/sd[f-p]$")
						if err != nil {
							return err
						}
						if r.FindString(input) == "" {
							return errors.New("Device name must follow the format: /dev/sd[f-p] (e.g. /dev/sdf, /dev/sdp")
						}
						return nil
					},
					Default: "/dev/sdf",
				}

				result, err := prompt.Run()
				if err != nil {
					return nil, err
				}
				cfg.EBSVolumeDeviceName = result
			}

			// Mount Path
			if mountPathIsSet {
				cfg.EBSVolumeMountPath = viper.GetString("ebs_volume_mount_path")
			} else {
				prompt := promptui.Prompt{
					Label: "EBS Volume Mount Path",
				}

				result, err := prompt.Run()
				if err != nil {
					return nil, err
				}
				cfg.EBSVolumeMountPath = result
			}

			if volumeTypeIsSet {
				cfg.EBSVolumeType = viper.GetString("ebs_volume_type")
			} else {
				prompt := promptui.Select{
					Label: "EBS Volume Type",
					Items: ebsVolumeTypes,
					Templates: &promptui.SelectTemplates{
						Label:    "{{ . }}?",
						Active:   fmt.Sprintf("%s {{ .Name | underline }}", promptui.IconSelect),
						Inactive: "  {{.Name}}",
						Selected: "  EBS Volume Type? {{.Name}}",
					},
				}

				i, _, err := prompt.Run()
				if err != nil {
					return nil, err
				}
				cfg.EBSVolumeType = ebsVolumeTypes[i].Key
			}

			// EBS Volume Size
			if volumeSizeIsSet {
				cfg.EBSVolumeSize = viper.GetString("ebs_volume_size")
			} else {
				defaultSize := ""
				for _, volumeType := range ebsVolumeTypes {
					if cfg.EBSVolumeType == volumeType.Key {
						defaultSize = volumeType.DefaultSize
						break
					}
				}
				prompt := promptui.Prompt{
					Label: "EBS Volume Size in GiB",
					Validate: func(input string) error {
						num, err := strconv.ParseInt(input, 10, 64)
						if err != nil {
							return errors.New("Invalid number")
						}
						if num <= 0 {
							return errors.New("Number must be greater than 0")
						}
						return nil
					},
					Default: defaultSize,
				}
				result, err := prompt.Run()
				if err != nil {
					return nil, err
				}
				cfg.EBSVolumeSize = result
			}
		}
	}

	// Get existing node names
	nodes, err := currentState.Nodes(selectedCluster)
	if err != nil {
		return []string{}, err
	}
	existingNames := []string{}
	for nodeName := range nodes {
		existingNames = append(existingNames, nodeName)
	}

	// Determine what the hostnames should be for the new node(s)
	newHostnames := getNewHostnames(existingNames, cfg.Hostname, cfg.NodeCount)

	// Add new node to terraform config with the new hostnames
	for _, newHostname := range newHostnames {
		cfgCopy := cfg
		cfgCopy.Hostname = newHostname
		err = currentState.AddNode(selectedCluster, newHostname, cfgCopy)
		if err != nil {
			return []string{}, err
		}
	}

	return newHostnames, nil
}
