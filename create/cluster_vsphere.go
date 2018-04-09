package create

import (
	"fmt"

	"github.com/joyent/triton-kubernetes/backend"
	"github.com/joyent/triton-kubernetes/state"
)

const (
	vSphereClusterKeyFormat                     = "module.cluster_vsphere_%s"
	vSphereRancherKubernetesTerraformModulePath = "terraform/modules/vsphere-rancher-k8s"
)

// This struct represents the definition of a Terraform .tf file.
// Marshalled into json this struct can be passed directly to Terraform.
type vSphereClusterTerraformConfig struct {
	baseClusterTerraformConfig
}

// Returns the name of the cluster that was created and the new state.
func newVSphereCluster(remoteBackend backend.Backend, currentState state.State) (string, error) {
	baseConfig, err := getBaseClusterTerraformConfig(vSphereRancherKubernetesTerraformModulePath)
	if err != nil {
		return "", err
	}

	cfg := vSphereClusterTerraformConfig{
		baseClusterTerraformConfig: baseConfig,
	}

	// Add new cluster to terraform config
	err = currentState.Add(fmt.Sprintf(vSphereClusterKeyFormat, cfg.Name), &cfg)
	if err != nil {
		return "", err
	}

	return cfg.Name, nil
}
