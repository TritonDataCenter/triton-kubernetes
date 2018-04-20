package get

import (
	"testing"

	"github.com/joyent/triton-kubernetes/test_pkg"
	"github.com/spf13/viper"
)

func TestNoClusterManager(t *testing.T) {

	tCase := test_pkg.NewT(t)
	localBackend := &test_pkg.MockEmptyBackend{}
	err := GetManager(localBackend)

	expected := "No cluster managers."
	if expected != err.Error() {
		tCase.Fatal("output", expected, err)
	}
}

func TestMissingClusterManager(t *testing.T) {
	viper.Set("non-interactive", true)
	tCase := test_pkg.NewT(t)
	localBackend := &test_pkg.MockBackend{}
	err := GetManager(localBackend)

	expected := "Cluster manager must be specified"
	if expected != err.Error() {
		tCase.Fatal("output", expected, err)
	}
}

func TestClusterManagerNotExists(t *testing.T) {
	viper.Set("cluster_manager", "prod-cluster")
	tCase := test_pkg.NewT(t)
	localBackend := &test_pkg.MockBackend{}
	err := GetManager(localBackend)

	expected := "Selected cluster manager 'prod-cluster' does not exist."
	if expected != err.Error() {
		tCase.Fatal("output", expected, err)
	}
}
