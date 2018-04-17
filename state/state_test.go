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
		t.Errorf("value in state object, got: %s, want: %s.", dumyKey, "")
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
