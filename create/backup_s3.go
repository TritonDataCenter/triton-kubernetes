package create

import (
	"errors"
	"fmt"
	"strings"

	"github.com/joyent/triton-kubernetes/state"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

const (
	s3BackupTerraformModulePath = "terraform/modules/k8s-backup-s3"
)

type s3BackupTerraformConfig struct {
	baseBackupTerraformConfig

	AWSAccessKey string `json:"aws_access_key,omitempty"`
	AWSSecretKey string `json:"aws_secret_key,omitempty"`
	AWSRegion    string `json:"aws_region,omitempty"`
	AWSS3Bucket  string `json:"aws_s3_bucket,omitempty"`
}

func newS3Backup(selectedClusterKey string, currentState state.State) error {
	nonInteractiveMode := viper.GetBool("non-interactive")
	baseConfig, err := getBaseBackupTerraformConfig(s3BackupTerraformModulePath, selectedClusterKey)
	if err != nil {
		return err
	}

	cfg := s3BackupTerraformConfig{
		baseBackupTerraformConfig: baseConfig,
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

	// AWS S3 Bucket
	s3Config := aws.NewConfig().WithCredentials(creds).WithRegion(cfg.AWSRegion)
	s3Sess, err := session.NewSession(s3Config)
	if err != nil {
		return err
	}
	s3Client := s3.New(s3Sess)

	bucketsResult, err := s3Client.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		return err
	}
	existingBuckets := bucketsResult.Buckets

	if viper.IsSet("aws_s3_bucket") {
		bucketInput := viper.GetString("aws_s3_bucket")
		found := false
		for _, bucket := range existingBuckets {
			if bucket.Name != nil && *bucket.Name == bucketInput {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("Could not find bucket '%s'. Backup creation canceled.", bucketInput)
		}
		cfg.AWSS3Bucket = bucketInput
	} else if nonInteractiveMode {
		return errors.New("aws_s3_bucket must be defined.")
	} else {
		// Building an array of strings that will be given to the SelectPrompt.
		// The SelectTemplate has problems displaying struct fields that are string pointers.
		bucketNames := []string{}
		for _, bucket := range existingBuckets {
			name := ""
			if bucket.Name != nil {
				name = *bucket.Name
			}
			bucketNames = append(bucketNames, name)
		}
		// Prompt for bucket
		searcher := func(input string, index int) bool {
			name := strings.Replace(strings.ToLower(bucketNames[index]), " ", "", -1)
			input = strings.Replace(strings.ToLower(input), " ", "", -1)

			return strings.Contains(name, input)
		}
		prompt := promptui.Select{
			Label: "AWS S3 Bucket",
			Items: bucketNames,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}?",
				Active:   fmt.Sprintf("%s {{ . | underline }}", promptui.IconSelect),
				Inactive: " {{.}}",
				Selected: fmt.Sprintf(`{{ "%s" | green }} {{ "AWS S3 Bucket:" | bold}} {{ . }}`, promptui.IconGood),
			},
			Searcher: searcher,
		}

		i, _, err := prompt.Run()
		if err != nil {
			return err
		}
		cfg.AWSS3Bucket = bucketNames[i]
	}

	err = currentState.AddBackup(selectedClusterKey, cfg)
	if err != nil {
		return err
	}

	return nil
}
