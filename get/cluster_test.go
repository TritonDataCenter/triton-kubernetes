package get

import (
	"testing"
	"github.com/mesoform/triton-kubernetes/backend/mocks"
	"github.com/mesoform/triton-kubernetes/state"
	"github.com/spf13/viper"
)

func TestGetClusterNoClusterManager(t *testing.T) {
	localBackend := &mocks.Backend{}

	localBackend.On("States").Return([]string{}, nil)

	err := GetCluster(localBackend)

	expected := "No cluster managers."

	if expected != err.Error() {
		t.Errorf("Wrong output, expected %s, received %s", expected, err.Error())
	}
}

func TestMissingClusterManagerNonInteractiveMode(t *testing.T){
	viper.Set("non-interactive", true)

	defer viper.Reset()

	localBackend := &mocks.Backend{}

	localBackend.On("States").Return([]string{"dev-manager", "test-manager"}, nil)

	err := GetCluster(localBackend)

	expected := "cluster_manager must be specified"

	if expected != err.Error() {
		t.Errorf("Wrong output, expected %s, received %s", expected, err.Error())
	}

}

func TestUnidentifiedClusterManagerNonInteractiveMode(t *testing.T){
	viper.Set("non-interactive", true)
	viper.Set("cluster_manager", "xyz")

	defer viper.Reset()

	localBackend := &mocks.Backend{}

	localBackend.On("States").Return([]string{"dev-manager", "test-manager"}, nil)

	err := GetCluster(localBackend)

	expected := "Selected cluster manager 'xyz' does not exist."

	if expected != err.Error() {
		t.Errorf("Wrong output, expected %s, received %s", expected, err.Error())
	}

}

func TestNoCluster(t *testing.T){
	viper.Set("non-interactive", true)
	viper.Set("cluster_manager", "dev-manager")

	defer viper.Reset()

	stateObj, _ := state.New("ClusterState", []byte(`{
		"module":{}
				}`))

	clusterManagerBackend := &mocks.Backend{}

	clusterManagerBackend.On("States").Return([]string{"dev-manager", "test-manager"}, nil)

	clusterManagerBackend.On("State", "dev-manager").Return(stateObj, nil)

	err:= GetCluster(clusterManagerBackend)

	expected := "No clusters."

	if expected != err.Error() {
		t.Errorf("Wrong output, expected %s, received %s", expected, err.Error())
	}

}

func TestNoClusterNameNonInterative(t *testing.T){
	viper.Set("non-interactive", true)
	viper.Set("cluster_manager", "dev-manager")

	defer viper.Reset()

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
		t.Errorf("Wrong output, expected %s, received %s", expected, err.Error())
	}

}

func TestUnidentifiedCluster(t *testing.T){
	viper.Set("non-interactive", true)
	viper.Set("cluster_manager", "dev-manager")
	viper.Set("cluster_name", "cluster_xyz")

	defer viper.Reset()

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
		t.Errorf("Wrong output, expected %s, received %s", expected, err.Error())
	}

}


