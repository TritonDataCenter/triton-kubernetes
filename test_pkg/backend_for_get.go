package test_pkg

import "github.com/joyent/triton-kubernetes/state"

type MockEmptyBackend struct {
}

func (backend *MockEmptyBackend) State(name string) (state.State, error) {
	return state.New(name, []byte("{}"))
}

func (backend *MockEmptyBackend) DeleteState(name string) error {
	return nil
}

func (backend *MockEmptyBackend) PersistState(state state.State) error {
	return nil
}

func (backend *MockEmptyBackend) States() ([]string, error) {
	return []string{}, nil
}

func (backend *MockEmptyBackend) StateTerraformConfig(name string) (string, interface{}) {
	return "", nil
}

type MockBackend struct {
}

func (backend *MockBackend) State(name string) (state.State, error) {
	return state.New(name, []byte("{}"))
}

func (backend *MockBackend) DeleteState(name string) error {
	return nil
}

func (backend *MockBackend) PersistState(state state.State) error {
	return nil
}

func (backend *MockBackend) States() ([]string, error) {
	return []string{"dev-manager", "beta-manager"}, nil
}

func (backend *MockBackend) StateTerraformConfig(name string) (string, interface{}) {
	return "", nil
}
