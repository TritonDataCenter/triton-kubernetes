package create

import (
	"fmt"

	"github.com/joyent/triton-kubernetes/backend"
	"github.com/joyent/triton-kubernetes/state"
)

const (
	bareMetalClusterKeyFormat                     = "module.cluster_baremetal_%s"
	bareMetalRancherKubernetesTerraformModulePath = "terraform/modules/bare-metal-rancher-k8s"
)

// This struct represents the definition of a Terraform .tf file.
// Marshalled into json this struct can be passed directly to Terraform.
type bareMetalClusterTerraformConfig struct {
	baseClusterTerraformConfig
}

// Returns the name of the cluster that was created and the new state.
func newBareMetalCluster(remoteBackend backend.Backend, currentState state.State) (string, error) {
	baseConfig, err := getBaseClusterTerraformConfig(bareMetalRancherKubernetesTerraformModulePath)
	if err != nil {
		return "", err
	}

	cfg := bareMetalClusterTerraformConfig{
		baseClusterTerraformConfig: baseConfig,
	}

	// Add new cluster to terraform config
	err = currentState.Add(fmt.Sprintf(bareMetalClusterKeyFormat, cfg.Name), &cfg)
	if err != nil {
		return "", err
	}

	return cfg.Name, nil
}
