package backend

import (
	"github.com/mesoform/triton-kubernetes/state"
)

type Backend interface {
	// State returns the current state.
	//
	// If the named state doesn't exist it will be created.
	State(name string) (state.State, error)

	// DeleteState removes the named state if it exists.
	//
	// DeleteState does not prevent deleting a state that is in use.
	DeleteState(name string) error

	// PersistState persist the given state.
	PersistState(state state.State) error

	// States returns a list of configured named states.
	States() ([]string, error)

	// StateTerraformConfig returns the path and object that
	// represents a terraform backend configuration
	StateTerraformConfig(name string) (string, interface{})
}
