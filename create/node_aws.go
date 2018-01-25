package create

import (
	"errors"
	"fmt"
	"strings"

	"github.com/joyent/triton-kubernetes/backend"
	"github.com/joyent/triton-kubernetes/state"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

const (
	awsNodeKeyFormat                            = "module.node_aws_%s"
	awsRancherKubernetesHostTerraformModulePath = "terraform/modules/aws-rancher-k8s-host"
)

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
}

// Adds new AWS nodes to the given cluster and manager.
// Returns:
// - a slice of the hostnames added
// - the new state
// - error or nil
func newAWSNode(selectedClusterManager, selectedCluster string, remoteBackend backend.Backend, state state.State) ([]string, error) {
	baseConfig, err := getBaseNodeTerraformConfig(awsRancherKubernetesHostTerraformModulePath, selectedCluster, state)
	if err != nil {
		return []string{}, err
	}

	cfg := awsNodeTerraformConfig{
		baseNodeTerraformConfig: baseConfig,

		// Grab variables from cluster config
		AWSAccessKey: state.Get(fmt.Sprintf("module.%s.aws_access_key", selectedCluster)),
		AWSSecretKey: state.Get(fmt.Sprintf("module.%s.aws_secret_key", selectedCluster)),
		AWSRegion:    state.Get(fmt.Sprintf("module.%s.aws_region", selectedCluster)),

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

	// Get existing node names
	nodes, err := state.Nodes(selectedCluster)
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
		err = state.Add(fmt.Sprintf(awsNodeKeyFormat, newHostname), cfgCopy)
		if err != nil {
			return []string{}, err
		}
	}

	return newHostnames, nil
}
