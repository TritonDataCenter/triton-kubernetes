package create

import (
	"github.com/joyent/triton-kubernetes/remote"

	"github.com/Jeffail/gabs"
)

const (
	gcpNodeKeyFormat                            = "node_gcp_%s"
	gcpRancherKubernetesHostTerraformModulePath = "terraform/modules/gcp-rancher-k8s-host"
)

type gcpNodeTerraformConfig struct {
	baseNodeTerraformConfig
}

func newGCPNode(selectedClusterManager, selectedCluster string, remoteClusterManagerState remote.RemoteClusterManagerStateManta, clusterManagerTerraformConfig *gabs.Container) error {
	baseConfig, err := getBaseNodeTerraformConfig(azureRancherKubernetesHostTerraformModulePath, selectedCluster)
	if err != nil {
		return err
	}

	_ = gcpNodeTerraformConfig{
		baseNodeTerraformConfig: baseConfig,
	}

	return nil
}
