package create

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/joyent/triton-kubernetes/remote"
	"github.com/joyent/triton-kubernetes/shell"

	"github.com/Jeffail/gabs"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

const (
	awsNodeKeyFormat                            = "node_aws_%s"
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

func newAWSNode(selectedClusterManager, selectedCluster string, remoteClusterManagerState remote.RemoteClusterManagerStateManta, clusterManagerTerraformConfig *gabs.Container) error {
	baseConfig, err := getBaseNodeTerraformConfig(awsRancherKubernetesHostTerraformModulePath, selectedCluster, clusterManagerTerraformConfig)
	if err != nil {
		return err
	}

	cfg := awsNodeTerraformConfig{
		baseNodeTerraformConfig: baseConfig,

		// Grab variables from cluster config
		AWSAccessKey: clusterManagerTerraformConfig.Path(fmt.Sprintf("module.%s.aws_access_key", selectedCluster)).Data().(string),
		AWSSecretKey: clusterManagerTerraformConfig.Path(fmt.Sprintf("module.%s.aws_secret_key", selectedCluster)).Data().(string),
		AWSRegion:    clusterManagerTerraformConfig.Path(fmt.Sprintf("module.%s.aws_region", selectedCluster)).Data().(string),

		// Reference terraform output variables from cluster module
		AWSSubnetID:        fmt.Sprintf("${module.%s.aws_subnet_id}", selectedCluster),
		AWSSecurityGroupID: fmt.Sprintf("${module.%s.aws_security_group_id}", selectedCluster),
		AWSKeyName:         fmt.Sprintf("${module.%s.aws_key_name}", selectedCluster),
	}

	creds := credentials.NewStaticCredentials(cfg.AWSAccessKey, cfg.AWSSecretKey, "")

	awsConfig := aws.NewConfig().WithCredentials(creds).WithRegion(cfg.AWSRegion)
	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return err
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
			return err
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
			return err
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
			return err
		}
		cfg.AWSInstanceType = result
	}

	// Add new node to terraform config
	nodeKey := fmt.Sprintf(awsNodeKeyFormat, cfg.Hostname)
	clusterManagerTerraformConfig.SetP(&cfg, fmt.Sprintf("module.%s", nodeKey))

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
