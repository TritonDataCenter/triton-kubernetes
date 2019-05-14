package create

import (
	"errors"

	"github.com/mesoform/triton-kubernetes/backend"
	"github.com/mesoform/triton-kubernetes/state"
	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

const (
	vSphereRancherKubernetesTerraformModulePath = "terraform/modules/vsphere-rancher-k8s"
)

// This struct represents the definition of a Terraform .tf file.
// Marshalled into json this struct can be passed directly to Terraform.
type vSphereClusterTerraformConfig struct {
	baseClusterTerraformConfig

	VSphereUser     string `json:"vsphere_user"`
	VSpherePassword string `json:"vsphere_password"`
	VSphereServer   string `json:"vsphere_server"`

	VSphereDatacenterName   string `json:"vsphere_datacenter_name"`
	VSphereDatastoreName    string `json:"vsphere_datastore_name"`
	VSphereResourcePoolName string `json:"vsphere_resource_pool_name"`
	VSphereNetworkName      string `json:"vsphere_network_name"`
}

// Returns the name of the cluster that was created and the new state.
func newVSphereCluster(remoteBackend backend.Backend, currentState state.State) (string, error) {
	nonInteractiveMode := viper.GetBool("non-interactive")
	baseConfig, err := getBaseClusterTerraformConfig(vSphereRancherKubernetesTerraformModulePath)
	if err != nil {
		return "", err
	}

	cfg := vSphereClusterTerraformConfig{
		baseClusterTerraformConfig: baseConfig,
	}

	// vSphere User
	if viper.IsSet("vsphere_user") {
		cfg.VSphereUser = viper.GetString("vsphere_user")
	} else if nonInteractiveMode {
		return "", errors.New("vsphere_user must be specified.")
	} else {
		prompt := promptui.Prompt{
			Label: "vSphere User",
			Validate: func(input string) error {
				if len(input) == 0 {
					return errors.New("Invalid vSphere user")
				}
				return nil
			},
		}

		result, err := prompt.Run()
		if err != nil {
			return "", err
		}
		cfg.VSphereUser = result
	}

	// vSphere Password
	if viper.IsSet("vsphere_password") {
		cfg.VSpherePassword = viper.GetString("vsphere_password")
	} else if nonInteractiveMode {
		return "", errors.New("vsphere_password must be specified.")
	} else {
		prompt := promptui.Prompt{
			Label: "vSphere Password",
			Mask:  '*',
		}

		result, err := prompt.Run()
		if err != nil {
			return "", err
		}
		cfg.VSpherePassword = result
	}

	// vSphere Server
	if viper.IsSet("vsphere_server") {
		cfg.VSphereServer = viper.GetString("vsphere_server")
	} else if nonInteractiveMode {
		return "", errors.New("vsphere_server must be specified.")
	} else {
		prompt := promptui.Prompt{
			Label: "vSphere Server",
			Validate: func(input string) error {
				if len(input) == 0 {
					return errors.New("Invalid vSphere server")
				}
				return nil
			},
		}

		result, err := prompt.Run()
		if err != nil {
			return "", err
		}
		cfg.VSphereServer = result
	}

	// vSphere Datacenter Name
	// TODO Fetch datacenters
	if viper.IsSet("vsphere_datacenter_name") {
		cfg.VSphereDatacenterName = viper.GetString("vsphere_datacenter_name")
	} else if nonInteractiveMode {
		return "", errors.New("vsphere_datacenter_name must be specified.")
	} else {
		prompt := promptui.Prompt{
			Label: "vSphere Datacenter Name",
			Validate: func(input string) error {
				if len(input) == 0 {
					return errors.New("Invalid vSphere datacenter name")
				}
				return nil
			},
		}

		result, err := prompt.Run()
		if err != nil {
			return "", err
		}
		cfg.VSphereDatacenterName = result
	}

	// vSphere Datastore Name
	// TODO Fetch datastores
	if viper.IsSet("vsphere_datastore_name") {
		cfg.VSphereDatastoreName = viper.GetString("vsphere_datastore_name")
	} else if nonInteractiveMode {
		return "", errors.New("vsphere_datastore_name must be specified.")
	} else {
		prompt := promptui.Prompt{
			Label: "vSphere Datastore Name",
			Validate: func(input string) error {
				if len(input) == 0 {
					return errors.New("Invalid vSphere datastore name")
				}
				return nil
			},
		}

		result, err := prompt.Run()
		if err != nil {
			return "", err
		}
		cfg.VSphereDatastoreName = result
	}

	// vSphere Resource Pool Name
	// TODO Fetch clusters from vsphere
	if viper.IsSet("vsphere_resource_pool_name") {
		cfg.VSphereResourcePoolName = viper.GetString("vsphere_resource_pool_name")
	} else if nonInteractiveMode {
		return "", errors.New("vsphere_resource_pool_name must be specified.")
	} else {
		prompt := promptui.Prompt{
			Label: "vSphere Resource Pool Name",
			Validate: func(input string) error {
				if len(input) == 0 {
					return errors.New("Invalid vSphere resource pool name")
				}
				return nil
			},
		}

		result, err := prompt.Run()
		if err != nil {
			return "", err
		}
		cfg.VSphereResourcePoolName = result
	}

	// vSphere Network Name
	// TODO Fetch Networks from vsphere
	if viper.IsSet("vsphere_network_name") {
		cfg.VSphereNetworkName = viper.GetString("vsphere_network_name")
	} else if nonInteractiveMode {
		return "", errors.New("vsphere_network_name must be specified.")
	} else {
		prompt := promptui.Prompt{
			Label: "vSphere Network Name",
			Validate: func(input string) error {
				if len(input) == 0 {
					return errors.New("Invalid vSphere network name")
				}
				return nil
			},
		}

		result, err := prompt.Run()
		if err != nil {
			return "", err
		}
		cfg.VSphereNetworkName = result
	}

	// Add new cluster to terraform config
	err = currentState.AddCluster("vsphere", cfg.Name, &cfg)
	if err != nil {
		return "", err
	}

	return cfg.Name, nil
}
