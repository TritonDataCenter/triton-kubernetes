package create

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/joyent/triton-kubernetes/backend"
	"github.com/joyent/triton-kubernetes/shell"
	"github.com/joyent/triton-kubernetes/state"
	"github.com/joyent/triton-kubernetes/util"
	"github.com/manifoldco/promptui"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

const (
	clusterBackupKeyFormat           = "module.cluster-backup_%s"
	clusterBackupTerraformModulePath = "terraform/modules/k8s-backup"
)

type baseClusterBackupTerraformConfig struct {
	Source string `json:"source"`

	RancherAccessKey string `json:"rancher_access_key,omitempty"`
	RancherSecretKey string `json:"rancher_secret_key,omitempty"`
	KubernetesHostIP string `json:"kubernetes_host_ip,omitempty"`
	ClusterID        string `json:"cluster_id,omitempty"`
	ClusterName      string `json:"cluster_name,omitempty"`
	AdminName        string `json:"admin_name,omitempty"`
	TritonSSHUser    string `json:"triton_ssh_user,omitempty"`
	TritonSSHHost    string `json:"triton_ssh_host,omitempty"`
	TritonKeyPath    string `json:"triton_key_path,omitempty"`
}

func NewClusterBackup(remoteBackend backend.Backend) error {
	nonInteractiveMode := viper.GetBool("non-interactive")
	clusterManagers, err := remoteBackend.States()
	if err != nil {
		return err
	}

	if len(clusterManagers) == 0 {
		return fmt.Errorf("No cluster managers, please create a cluster manager and cluster before creating a cluster backup.")
	}

	selectedClusterManager := ""
	if viper.IsSet("cluster_manager") {
		selectedClusterManager = viper.GetString("cluster_manager")
	} else if nonInteractiveMode {
		return errors.New("cluster_manager must be specified")
	} else {
		prompt := promptui.Select{
			Label: "Cluster Manager",
			Items: clusterManagers,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}?",
				Active:   fmt.Sprintf(`%s {{ . | underline }}`, promptui.IconSelect),
				Inactive: `  {{ . }}`,
				Selected: fmt.Sprintf(`{{ "%s" | green }} {{ "Cluster Manager:" | bold}} {{ . }}`, promptui.IconGood),
			},
		}

		_, value, err := prompt.Run()
		if err != nil {
			return err
		}

		selectedClusterManager = value
	}

	// Verify selected cluster manager exists
	found := false
	for _, clusterManager := range clusterManagers {
		if selectedClusterManager == clusterManager {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("Selected cluster manager '%s' does not exist.", selectedClusterManager)
	}

	currentState, err := remoteBackend.State(selectedClusterManager)
	if err != nil {
		return err
	}

	// Get existing clusters
	clusters, err := currentState.Clusters()
	if err != nil {
		return err
	}

	selectedClusterKey := ""
	if viper.IsSet("cluster_name") {
		clusterName := viper.GetString("cluster_name")
		clusterKey, ok := clusters[clusterName]
		if !ok {
			return fmt.Errorf("A cluster named '%s', does not exist.", clusterName)
		}

		selectedClusterKey = clusterKey
	} else if nonInteractiveMode {
		return errors.New("cluster_name must be specified")
	} else {
		clusterNames := make([]string, 0, len(clusters))
		for name := range clusters {
			clusterNames = append(clusterNames, name)
		}
		sort.Strings(clusterNames)
		prompt := promptui.Select{
			Label: "Cluster to deploy node to",
			Items: clusterNames,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}?",
				Active:   fmt.Sprintf("%s {{ . | underline }}", promptui.IconSelect),
				Inactive: " {{ . }}",
				Selected: fmt.Sprintf(`{{ "%s" | green }} {{ "Cluster:" | bold}} {{ . }}`, promptui.IconGood),
			},
		}

		_, value, err := prompt.Run()
		if err != nil {
			return err
		}
		selectedClusterKey = clusters[value]
	}

	err = validateRequiredNodeTypes(currentState, selectedClusterKey)
	if err != nil {
		return err
	}

	// Return error if cluster already has backup
	existingBackup, err := currentState.ClusterBackup(selectedClusterKey)
	if err != nil {
		return err
	}
	if existingBackup != "" {
		return errors.New("Backup already exists for this cluster.")
	}

	// TODO Prompt for backup storage type
	// Backup Storage Type
	selectedStorageType := ""
	if viper.IsSet("cluster_backup_storage_type") {
		selectedStorageType = viper.GetString("cluster_backup_storage_type")
	} else if nonInteractiveMode {
		return errors.New("cluster_backup_storage_type must be specified")
	} else {
		prompt := promptui.Select{
			Label: "Create Cluster Backup in which storage",
			Items: []string{"S3"},
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}?",
				Active:   fmt.Sprintf(`%s {{ . | underline }}`, promptui.IconSelect),
				Inactive: `  {{ . }}`,
				Selected: fmt.Sprintf(`{{ "%s" | green }} {{ "Storage:" | bold}} {{ . }}`, promptui.IconGood),
			},
		}

		_, value, err := prompt.Run()
		if err != nil {
			return err
		}

		selectedStorageType = strings.ToLower(value)
	}

	switch selectedStorageType {
	case "s3":
		_, err = newS3ClusterBackup(selectedClusterKey, currentState)
	default:
		return fmt.Errorf("Unsupported cluster backup storage '%s', cannot create backup", selectedStorageType)
	}

	if !nonInteractiveMode {
		label := "Proceed with the cluster backup creation"
		selected := "Proceed"
		confirmed, err := util.PromptForConfirmation(label, selected)
		if err != nil {
			return err
		}
		if !confirmed {
			fmt.Println("Cluster backup creation canceled.")
			return nil
		}
	}

	// Run terraform apply with state
	err = shell.RunTerraformApplyWithState(currentState)
	if err != nil {
		return err
	}

	// After terraform succeeds, commit state
	err = remoteBackend.PersistState(currentState)
	if err != nil {
		return err
	}

	return nil
}

func getBaseClusterBackupTerraformConfig(terraformModulePath, selectedCluster string) (baseClusterBackupTerraformConfig, error) {
	nonInteractiveMode := viper.GetBool("non-interactive")
	cfg := baseClusterBackupTerraformConfig{
		RancherAccessKey: "${module.cluster-manager.rancher_access_key}",
		RancherSecretKey: "${module.cluster-manager.rancher_secret_key}",
		KubernetesHostIP: "${element(module.cluster-manager.masters, 0)}",
		ClusterID:        fmt.Sprintf("${module.%s.rancher_environment_id}", selectedCluster),
		ClusterName:      fmt.Sprintf("${module.%s.name}", selectedCluster),
		AdminName:        "${module.cluster-manager.rancher_admin_username}",
		TritonSSHHost:    `${module.cluster-manager.ssh_bastion_ip == "" ? element(module.cluster-manager.masters, 0) : module.cluster-manager.ssh_bastion_ip}`,
	}

	baseSource := defaultSourceURL
	if viper.IsSet("source_url") {
		baseSource = viper.GetString("source_url")
	}

	baseSourceRef := defaultSourceRef
	if viper.IsSet("source_ref") {
		baseSourceRef = viper.GetString("source_ref")
	}

	cfg.Source = fmt.Sprintf("%s//%s?ref=%s", baseSource, terraformModulePath, baseSourceRef)

	// Triton SSH User
	if viper.IsSet("triton_ssh_user") {
		cfg.TritonSSHUser = viper.GetString("triton_ssh_user")
	} else if nonInteractiveMode {
		return baseClusterBackupTerraformConfig{}, errors.New("triton_ssh_user must be specified")
	} else {
		prompt := promptui.Prompt{
			Label:   "Triton SSH User",
			Default: "ubuntu",
		}

		result, err := prompt.Run()
		if err != nil {
			return baseClusterBackupTerraformConfig{}, err
		}
		cfg.TritonSSHUser = result
	}

	// Triton Key Path
	rawTritonKeyPath := ""
	if viper.IsSet("triton_key_path") {
		rawTritonKeyPath = viper.GetString("triton_key_path")
	} else if nonInteractiveMode {
		return baseClusterBackupTerraformConfig{}, errors.New("triton_key_path must be specified")
	} else {
		prompt := promptui.Prompt{
			Label: "Triton Key Path",
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
			return baseClusterBackupTerraformConfig{}, err
		}
		rawTritonKeyPath = result
	}

	expandedTritonKeyPath, err := homedir.Expand(rawTritonKeyPath)
	if err != nil {
		return baseClusterBackupTerraformConfig{}, err
	}
	cfg.TritonKeyPath = expandedTritonKeyPath

	return cfg, nil

}

func validateRequiredNodeTypes(currentState state.State, selectedClusterKey string) error {
	nodes, err := currentState.Nodes(selectedClusterKey)
	if err != nil {
		return err
	}

	hasOrch := false
	hasEtcd := false
	hasCompute := false
	for _, nodeKey := range nodes {
		orch := currentState.Get(fmt.Sprintf("module.%s.rancher_host_labels.orchestration", nodeKey))
		etcd := currentState.Get(fmt.Sprintf("module.%s.rancher_host_labels.etcd", nodeKey))
		compute := currentState.Get(fmt.Sprintf("module.%s.rancher_host_labels.compute", nodeKey))
		if orch == "true" {
			hasOrch = true
		}
		if etcd == "true" {
			hasEtcd = true
		}
		if compute == "true" {
			hasCompute = true
		}
	}
	if !(hasOrch && hasEtcd && hasCompute) {
		typesFound := ""
		if hasOrch {
			typesFound = "orchestration"
		}
		if hasEtcd {
			if len(typesFound) > 0 {
				typesFound += " and "
			}
			typesFound += "etcd"
		}
		if hasCompute {
			if len(typesFound) > 0 {
				typesFound += " and "
			}
			typesFound += "compute"
		}
		return fmt.Errorf("Cluster must have at least 1 node of each etcd, orchestration, and compute. Found %s", typesFound)
	}
	return nil
}
