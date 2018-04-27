package destroy

import (
	"testing"

	"github.com/joyent/triton-kubernetes/backend/mocks"
	"github.com/spf13/viper"
)

func TestDeleteManagerNoClusterManager(t *testing.T) {

	localBackend := &mocks.Backend{}
	localBackend.On("States").Return([]string{}, nil)

	expected := "No cluster managers, please create a cluster manager before creating a kubernetes cluster."

	err := DeleteManager(localBackend)
	if expected != err.Error() {
		t.Errorf("Wrong output, expected %s, received %s", expected, err.Error())
	}
}

func TestDeleteManagerMissingClusterManager(t *testing.T) {
	viper.Reset()
	viper.Set("non-interactive", true)

	localBackend := &mocks.Backend{}
	localBackend.On("States").Return([]string{"dev-manager", "beta-manager"}, nil)

	expected := "cluster_manager must be specified"

	err := DeleteManager(localBackend)
	if expected != err.Error() {
		t.Errorf("Wrong output, expected %s, received %s", expected, err.Error())
	}
}

func TestDeleteManagerNotExist(t *testing.T) {
	viper.Reset()
	viper.Set("non-interactive", true)
	viper.Set("cluster_manager", "prod-cluster")

	localBackend := &mocks.Backend{}
	localBackend.On("States").Return([]string{"dev-manager", "beta-manager"}, nil)

	expected := "Selected cluster manager 'prod-cluster' does not exist."

	err := DeleteManager(localBackend)
	if expected != err.Error() {
		t.Errorf("Wrong output, expected %s, received %s", expected, err.Error())
	}
}
