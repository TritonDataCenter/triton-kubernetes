package state

import (
	"testing"
)

// Get test
func TestGet(t *testing.T) {
	stateObj, err := New("GetState", []byte(`{"config":{"triton":{"key":"55fd4s","url":"https://api.storage.com"}}}`))
	if err != nil {
		t.Error(err)
	}

	dumyKey := stateObj.Get("config.triton.key")
	if dumyKey != "55fd4s" {
		t.Errorf("value in state object, got: %s, want: %s.", dumyKey, "55fd4s")
	}
}

// Add test
func TestAdd(t *testing.T) {
	stateObj, err := New("AddState", []byte(`{}`))
	if err != nil {
		t.Error(err)
	}

	err1 := stateObj.Add("config.triton.dns", "aws-provider")
	if err1 != nil {
		t.Error(err1)
	}

	notEmptyPath := stateObj.Get("config.triton.dns")
	if notEmptyPath != "aws-provider" {
		t.Errorf("value in state object, got: %s, want: %s", notEmptyPath, "aws-provider")
	}
}

// Delete test
func TestDelete(t *testing.T) {
	stateObj, err := New("DelState", []byte(`{"config":{"triton":{"key":"55fd4s","url":"https://api.storage.com"}}}`))
	if err != nil {
		t.Error(err)
	}

	err1 := stateObj.Delete("config.triton.key")
	if err1 != nil {
		t.Error(err1)
	}

	deletedKey := stateObj.Get("config.triton.key")
	if deletedKey != "" {
		t.Errorf("value in state object, got: %s, want: %s.", deletedKey, "")
	}
}

// GetClusters test
func TestGetClusters(t *testing.T) {
	stateObj, err := New("ClusterState", []byte(`{
		"module":{
						"cluster_1":{"name":"dev_cluster"},
						"cluster_2":{"name":"beta_cluster"},
						"cluster_3":{"name":"prod_cluster"}
					}
				}`))
	if err != nil {
		t.Error(err)
	}

	clusterMap, err := stateObj.Clusters()
	if err != nil {
		t.Error(err)
	}

	if len(clusterMap) != 3 {
		t.Errorf("wrong length of map: %v", len(clusterMap))
	}
}


func TestGetNodes(t *testing.T) {

	clusterStateObj, err := New("ClusterState", []byte(`{"config":{"triton":{"key":"cluster_triton_dev-cluster","url":"https://api.storage.com"}}}`))


	clusterKeyString := clusterStateObj.Get("config.triton.key")

	stateObj, err := New("NodeState", []byte(`{
    "module":{
      "node_triton_dev-cluster_1":{"hostname":"dev-worker1"},
      "node_triton_dev-cluster_2":{"hostname":"dev-etcd1"},
      "node_triton_dev-cluster_3":{"hostname":"dev-control1"},
      "node_aws_dev-cluster_1":{"hostname":"dev-control2"},
	  "node_aws_dev-cluster_2":{"hostname":"dev-control3"}
    }
    }`))

	if err != nil {
		t.Error(err)
	}

	nodeMap, err := stateObj.Nodes(clusterKeyString)

	if err != nil {
		t.Error(err)
	}

	if len(nodeMap) != 3 {
		t.Errorf("wrong length of map: %v", len(nodeMap))
	}

}
