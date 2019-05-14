package get

import (
	"testing"

	"github.com/mesoform/triton-kubernetes/backend/mocks"
	"github.com/spf13/viper"
)

func TestNoClusterManager(t *testing.T) {

	localBackend := &mocks.Backend{}
	localBackend.On("States").Return([]string{}, nil)

	expected := "No cluster managers."

	err := GetManager(localBackend)
	if expected != err.Error() {
		t.Errorf("Wrong output, expected %s, received %s", expected, err.Error())
	}
}

func TestMissingClusterManager(t *testing.T) {
	viper.Set("non-interactive", true)

	localBackend := &mocks.Backend{}
	localBackend.On("States").Return([]string{"dev-manager", "beta-manager"}, nil)

	expected := "cluster_manager must be specified"

	err := GetManager(localBackend)
	if expected != err.Error() {
		t.Errorf("Wrong output, expected %s, received %s", expected, err.Error())
	}
}

func TestClusterManagerNotExists(t *testing.T) {
	viper.Set("cluster_manager", "prod-cluster")

	localBackend := &mocks.Backend{}
	localBackend.On("States").Return([]string{"dev-manager", "beta-manager"}, nil)

	expected := "Selected cluster manager 'prod-cluster' does not exist."

	err := GetManager(localBackend)
	if expected != err.Error() {
		t.Errorf("Wrong output, expected %s, received %s", expected, err.Error())
	}
}
