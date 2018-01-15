package local

import (
	"fmt"

	"github.com/joyent/triton-kubernetes/backend"
	"github.com/joyent/triton-kubernetes/state"
)

const (
	rootDirectory             = "~/.triton-kubernetes"
	rootPathFormat            = rootDirectory + "/%s"
	terraformConfigPathFormat = rootDirectory + "/%s/main.tf.json"
	terraformStatePathFormat  = rootDirectory + "/%s/terraform.tfstate"
)

type localBackend struct {
}

type localTerraformBackendConfig struct {
	Path string `json:"path"`
}

func New() (backend.Backend, error) {
	return localBackend{}, nil
}

func (backend localBackend) State(name string) (state.State, error) {
	return state.State{}, nil
}

func (backend localBackend) DeleteState(name string) error {
	return nil
}

func (backend localBackend) PersistState(state state.State) error {
	return nil
}

func (backend localBackend) States() ([]string, error) {
	return nil, nil
}

func (backend localBackend) StateTerraformConfig(name string) (string, interface{}) {
	return "terraform.backend.local", localTerraformBackendConfig{Path: fmt.Sprintf(terraformStatePathFormat, name)}
}
