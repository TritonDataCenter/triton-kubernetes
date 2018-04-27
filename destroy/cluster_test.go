package destroy

import (
	"testing"

	"github.com/joyent/triton-kubernetes/backend/mocks"
	"github.com/joyent/triton-kubernetes/state"
	"github.com/spf13/viper"
)

var mockClusters = []byte(`{
								"module":{
												"cluster_1":{"name":"dev_cluster"},
												"cluster_2":{"name":"beta_cluster"},
												"cluster_3":{"name":"prod_cluster"}
								}
}`)

func TestDeleteClusterNoClusterManager(t *testing.T) {

	localBackend := &mocks.Backend{}
	localBackend.On("States").Return([]string{}, nil)

	expected := "No cluster managers."

	err := DeleteCluster(localBackend)
	if expected != err.Error() {
		t.Errorf("Wrong output, expected %s, received %s", expected, err.Error())
	}
}

func TestDeleteClusterMissingClusterManager(t *testing.T) {
	viper.Reset()
	viper.Set("non-interactive", true)

	localBackend := &mocks.Backend{}
	localBackend.On("States").Return([]string{"dev-manager", "beta-manager"}, nil)

	expected := "cluster_manager must be specified"

	err := DeleteCluster(localBackend)
	if expected != err.Error() {
		t.Errorf("Wrong output, expected %s, received %s", expected, err.Error())
	}
}

func TestDeleteClusterManagerNotExists(t *testing.T) {
	viper.Reset()
	viper.Set("non-interactive", true)
	viper.Set("cluster_manager", "prod-manager")

	localBackend := &mocks.Backend{}
	localBackend.On("States").Return([]string{"dev-manager", "beta-manager"}, nil)

	expected := "Selected cluster manager 'prod-manager' does not exist."

	err := DeleteCluster(localBackend)
	if expected != err.Error() {
		t.Errorf("Wrong output, expected %s, received %s", expected, err.Error())
	}
}

func TestDeleteClusterMustSpecifyClusterName(t *testing.T) {
	viper.Reset()
	viper.Set("non-interactive", true)
	viper.Set("cluster_manager", "dev-manager")

	stateObj, _ := state.New("ClusterState", mockClusters)

	backend := &mocks.Backend{}
	backend.On("States").Return([]string{"dev-manager", "beta-manager"}, nil)
	backend.On("State", "dev-manager").Return(stateObj, nil)

	expected := "cluster_name must be specified"

	err := DeleteCluster(backend)
	if expected != err.Error() {
		t.Errorf("Wrong output, expected %s, received %s", expected, err.Error())
	}
}

func TestDeleteClusterNotExists(t *testing.T) {
	viper.Reset()
	viper.Set("non-interactive", true)
	viper.Set("cluster_manager", "dev-manager")
	viper.Set("cluster_name", "cluster_alpha")

	stateObj, _ := state.New("ClusterState", mockClusters)

	backend := &mocks.Backend{}
	backend.On("States").Return([]string{"dev-manager", "beta-manager"}, nil)
	backend.On("State", "dev-manager").Return(stateObj, nil)

	expected := "A cluster named 'cluster_alpha', does not exist."

	err := DeleteCluster(backend)
	if expected != err.Error() {
		t.Errorf("Wrong output, expected %s, received %s", expected, err.Error())
	}
}
