package destroy

import (
	"testing"

	"github.com/mesoform/triton-kubernetes/backend/mocks"
	"github.com/mesoform/triton-kubernetes/state"
	"github.com/spf13/viper"
)

var mockNodeHost = []byte(`{
								"module":{
												"cluster_triton_dev-cluster":{"name":"dev_cluster"}
								}
}`)

func TestDeleteNodeNoClusterManager(t *testing.T) {

	localBackend := &mocks.Backend{}
	localBackend.On("States").Return([]string{}, nil)

	expected := "No cluster managers, please create a cluster manager before creating a kubernetes node."

	err := DeleteNode(localBackend)
	if expected != err.Error() {
		t.Errorf("Wrong output, expected %s, received %s", expected, err.Error())
	}
}

func TestDeleteNodeMissingClusterManager(t *testing.T) {
	viper.Reset()
	viper.Set("non-interactive", true)

	localBackend := &mocks.Backend{}
	localBackend.On("States").Return([]string{"dev-manager", "beta-manager"}, nil)

	expected := "cluster_manager must be specified"

	err := DeleteNode(localBackend)
	if expected != err.Error() {
		t.Errorf("Wrong output, expected %s, received %s", expected, err.Error())
	}
}

func TestDeleteNodeManagerNotExist(t *testing.T) {
	viper.Reset()
	viper.Set("non-interactive", true)
	viper.Set("cluster_manager", "prod-manager")

	localBackend := &mocks.Backend{}
	localBackend.On("States").Return([]string{"dev-manager", "beta-manager"}, nil)

	expected := "Selected cluster manager 'prod-manager' does not exist."

	err := DeleteNode(localBackend)
	if expected != err.Error() {
		t.Errorf("Wrong output, expected %s, received %s", expected, err.Error())
	}
}
func TestDeleteNodeMustSpecifyClusterName(t *testing.T) {
	viper.Reset()
	viper.Set("non-interactive", true)
	viper.Set("cluster_manager", "dev-manager")

	stateObj, _ := state.New("NodeState", mockClusters)

	backend := &mocks.Backend{}
	backend.On("States").Return([]string{"dev-manager", "beta-manager"}, nil)
	backend.On("State", "dev-manager").Return(stateObj, nil)

	expected := "cluster_name must be specified"

	err := DeleteNode(backend)
	if expected != err.Error() {
		t.Errorf("Wrong output, expected %s, received %s", expected, err.Error())
	}
}

func TestDeleteNodeClusterNotExist(t *testing.T) {
	viper.Reset()
	viper.Set("non-interactive", true)
	viper.Set("cluster_manager", "dev-manager")
	viper.Set("cluster_name", "cluster_alpha")

	stateObj, _ := state.New("NodeState", mockClusters)

	backend := &mocks.Backend{}
	backend.On("States").Return([]string{"dev-manager", "beta-manager"}, nil)
	backend.On("State", "dev-manager").Return(stateObj, nil)

	expected := "A cluster named 'cluster_alpha', does not exist."

	err := DeleteNode(backend)
	if expected != err.Error() {
		t.Errorf("Wrong output, expected %s, received %s", expected, err.Error())
	}
}

func TestDeleteNodeHostNameMustSpecified(t *testing.T) {
	viper.Reset()
	viper.Set("non-interactive", true)
	viper.Set("cluster_manager", "dev-manager")
	viper.Set("cluster_name", "dev_cluster")

	stateObj, _ := state.New("NodeState", mockNodeHost)

	backend := &mocks.Backend{}
	backend.On("States").Return([]string{"dev-manager", "beta-manager"}, nil)
	backend.On("State", "dev-manager").Return(stateObj, nil)

	expected := "hostname must be specified"

	err := DeleteNode(backend)
	if expected != err.Error() {
		t.Errorf("Wrong output, expected %s, received %s", expected, err.Error())
	}
}

func TestDeleteNodeNotExist(t *testing.T) {
	viper.Reset()
	viper.Set("non-interactive", true)
	viper.Set("cluster_manager", "dev-manager")
	viper.Set("cluster_name", "dev_cluster")
	viper.Set("hostname", "dev_node_host")

	stateObj, _ := state.New("NodeState", mockNodeHost)

	backend := &mocks.Backend{}
	backend.On("States").Return([]string{"dev-manager", "beta-manager"}, nil)
	backend.On("State", "dev-manager").Return(stateObj, nil)

	expected := "A node named 'dev_node_host', does not exist."

	err := DeleteNode(backend)
	if expected != err.Error() {
		t.Errorf("Wrong output, expected %s, received %s", expected, err.Error())
	}
}
