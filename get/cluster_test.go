package get

import (
	"testing"
	"github.com/joyent/triton-kubernetes/backend/mocks"
	"github.com/joyent/triton-kubernetes/state"
	"github.com/spf13/viper"
)

func TestGetClusterNoClusterManager(t *testing.T) {
	localBackend := &mocks.Backend{}

	localBackend.On("States").Return([]string{}, nil)

	err := GetCluster(localBackend)

	expected := "No cluster managers."

	if expected != err.Error() {
		t.Error(err)
	}
}

func TestMissingClusterManagerNonInteractiveMode(t *testing.T){
	viper.Set("non-interactive", true)

	localBackend := &mocks.Backend{}

	localBackend.On("States").Return([]string{"dev-manager", "test-manager"}, nil)

	err := GetCluster(localBackend)

	expected := "cluster_manager must be specified"

	if expected != err.Error() {
		t.Error(err)
	}

}

func TestUnidentifiedClusterManagerNonInteractiveMode(t *testing.T){
	viper.Set("non-interactive", true)
	viper.Set("cluster_manager", "xyz")

	localBackend := &mocks.Backend{}

	localBackend.On("States").Return([]string{"dev-manager", "test-manager"}, nil)

	err := GetCluster(localBackend)

	expected := "Selected cluster manager 'xyz' does not exist."

	if expected != err.Error() {
		t.Error(err)
	}

}

func TestNoCluster(t *testing.T){
	viper.Set("non-interactive", true)
	viper.Set("cluster_manager", "dev-manager")

	stateObj, _ := state.New("ClusterState", []byte(`{
		"module":{}
				}`))

	clusterManagerBackend := &mocks.Backend{}

	clusterManagerBackend.On("States").Return([]string{"dev-manager", "test-manager"}, nil)

	clusterManagerBackend.On("State", "dev-manager").Return(stateObj, nil)

	err:= GetCluster(clusterManagerBackend)

	expected := "No clusters."

	if expected != err.Error() {
		t.Error(err)
	}

}

func TestNoClusterNameNonInterative(t *testing.T){

	stateObj, _ := state.New("ClusterState", []byte(`{
		"module":{
					"cluster_1":{"name":"dev_cluster"},
					"cluster_2":{"name":"beta_cluster"},
					"cluster_3":{"name":"prod_cluster"}
				}
				}`))

	clusterManagerBackend := &mocks.Backend{}

	clusterManagerBackend.On("States").Return([]string{"dev-manager", "test-manager"}, nil)

	clusterManagerBackend.On("State", "dev-manager").Return(stateObj, nil)

	err:= GetCluster(clusterManagerBackend)

	expected := "cluster_name must be specified"

	if expected != err.Error() {
		t.Error(err)
	}

}

func TestUnidentifiedCluster(t *testing.T){
	viper.Set("non-interactive", true)
	viper.Set("cluster_manager", "dev-manager")
	viper.Set("cluster_name", "cluster_xyz")

	stateObj, _ := state.New("ClusterState", []byte(`{
		"module":{
					"cluster_1":{"name":"dev_cluster"},
					"cluster_2":{"name":"beta_cluster"},
					"cluster_3":{"name":"prod_cluster"}
				}
				}`))

	clusterManagerBackend := &mocks.Backend{}

	clusterManagerBackend.On("States").Return([]string{"dev-manager", "test-manager"}, nil)

	clusterManagerBackend.On("State", "dev-manager").Return(stateObj, nil)

	err:= GetCluster(clusterManagerBackend)

	expected := "A cluster named 'cluster_xyz', does not exist."

	if expected != err.Error() {
		t.Error(err)
	}

}


