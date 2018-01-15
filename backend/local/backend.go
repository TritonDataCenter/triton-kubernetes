package local

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/joyent/triton-kubernetes/backend"
	"github.com/joyent/triton-kubernetes/state"

	homedir "github.com/mitchellh/go-homedir"
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
	// Create root directory
	os.MkdirAll(rootDirectory, os.ModePerm)

	return localBackend{}, nil
}

func (backend localBackend) State(name string) (state.State, error) {
	terraformConfigPath := fmt.Sprintf(terraformConfigPathFormat, name)

	expandedTerraformConfigPath, err := homedir.Expand(terraformConfigPath)
	if err != nil {
		return state.State{}, err
	}

	_, err = os.Stat(expandedTerraformConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return state.New(name, []byte("{}"))
		}

		return state.State{}, err
	}

	content, err := ioutil.ReadFile(expandedTerraformConfigPath)
	if err != nil {
		return state.State{}, err
	}

	return state.New(name, content)
}

func (backend localBackend) DeleteState(name string) error {
	rootPath := fmt.Sprintf(rootPathFormat, name)

	expandedRootPath, err := homedir.Expand(rootPath)
	if err != nil {
		return err
	}

	return os.RemoveAll(expandedRootPath)
}

func (backend localBackend) PersistState(state state.State) error {
	rootPath := fmt.Sprintf(rootPathFormat, state.Name)
	expandedRootPath, err := homedir.Expand(rootPath)
	if err != nil {
		return err
	}

	os.MkdirAll(expandedRootPath, os.ModePerm)

	terraformConfigPath := fmt.Sprintf(terraformConfigPathFormat, state.Name)
	expandedTerraformConfigPath, err := homedir.Expand(terraformConfigPath)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(expandedTerraformConfigPath, state.Bytes(), 0644)
	if err != nil {
		return err
	}

	return nil
}

func (backend localBackend) States() ([]string, error) {
	expandedRootDirectory, err := homedir.Expand(rootDirectory)
	if err != nil {
		return nil, err
	}

	files, err := ioutil.ReadDir(expandedRootDirectory)
	if err != nil {
		return nil, err
	}

	states := []string{}
	for _, f := range files {
		if f.IsDir() {
			states = append(states, f.Name())
		}
	}

	return states, nil
}

func (backend localBackend) StateTerraformConfig(name string) (string, interface{}) {
	terraformStatePath := fmt.Sprintf(terraformStatePathFormat, name)
	expandedTerraformStatePath, _ := homedir.Expand(terraformStatePath)

	terraformBackendConfig := localTerraformBackendConfig{
		Path: expandedTerraformStatePath,
	}

	return "terraform.backend.local", terraformBackendConfig
}
