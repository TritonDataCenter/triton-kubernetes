package remote

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"

	triton "github.com/joyent/triton-go"
	"github.com/joyent/triton-go/authentication"
	"github.com/joyent/triton-go/storage"
)

const (
	rootDirectory                    = "/stor/triton-kubernetes"
	clusterManagerRootPathFormat     = rootDirectory + "/%s"
	clusterManagerTFConfigPathFormat = rootDirectory + "/%s/main.tf.json"
	clusterManagerTFStatePathFormat  = rootDirectory + "/%s/terraform.tfstate"
)

// Stores terraform json configuration files for all cluster managers in Manta
// Each cluster manager has a separate directory under triton-kubernetes with a main.tf.json file
// and a terraform.tfstate file.
// triton-kubernetes manages the main.tf.json file and terraform manages the terraform.tfstate file
// Directory Path: /stor/triton-kubernetes/${CLUSTER_MANAGER_NAME}/main.tf.json
// TODO: Lock terraform json configuration similar to how terraform locks tfstate file.
type RemoteClusterManagerStateManta struct {
	tritonStorageClient *storage.StorageClient
}

// List all cluster managers
func (manta *RemoteClusterManagerStateManta) List() ([]string, error) {
	input := storage.ListDirectoryInput{
		DirectoryName: rootDirectory,
		Limit:         100,
	}

	result, err := manta.tritonStorageClient.Dir().List(context.Background(), &input)
	if err != nil {
		return nil, err
	}

	if result.ResultSetSize == 0 {
		return []string{}, nil
	}

	clusterManagers := []string{}
	for _, clusterManager := range result.Entries {
		clusterManagers = append(clusterManagers, clusterManager.Name)
	}

	return clusterManagers, nil
}

func (manta *RemoteClusterManagerStateManta) GetTerraformConfig(clusterManagerName string) ([]byte, error) {
	tfConfigPath := fmt.Sprintf(clusterManagerTFConfigPathFormat, clusterManagerName)

	getObjectInput := &storage.GetObjectInput{
		ObjectPath: tfConfigPath,
	}
	output, err := manta.tritonStorageClient.Objects().Get(context.Background(), getObjectInput)
	if err != nil {
		return nil, err
	}

	currentConfigBytes, err := ioutil.ReadAll(output.ObjectReader)
	if err != nil {
		return nil, err
	}

	return currentConfigBytes, nil
}

func (manta *RemoteClusterManagerStateManta) CommitTerraformConfig(clusterManagerName string, clusterManagerConfigJSON []byte) error {
	tfConfigPath := fmt.Sprintf(clusterManagerTFConfigPathFormat, clusterManagerName)

	objInput := storage.PutObjectInput{
		ObjectPath:   tfConfigPath,
		ContentType:  "application/json",
		ObjectReader: bytes.NewReader(clusterManagerConfigJSON),
	}
	err := manta.tritonStorageClient.Objects().Put(context.Background(), &objInput)
	if err != nil {
		return err
	}

	return nil
}

// Deletes the cluster manager state folder
// Note: For some reason, attempting to force delete the entire directory only deletes the json file.
// As a workaround, this function explicitly deletes the json file and tfstate file before deleting the directory.
func (manta *RemoteClusterManagerStateManta) Delete(clusterManagerName string) error {
	objClient := manta.tritonStorageClient.Objects()

	// Deleting the main.tf.json file
	tfConfigPath := fmt.Sprintf(clusterManagerTFConfigPathFormat, clusterManagerName)
	deleteObjInput := &storage.DeleteObjectInput{
		ObjectPath: tfConfigPath,
	}
	err := objClient.Delete(context.Background(), deleteObjInput)
	if err != nil {
		return err
	}

	// Deleting the terraform.tfstate file
	tfStatePath := fmt.Sprintf(clusterManagerTFStatePathFormat, clusterManagerName)
	deleteObjInput = &storage.DeleteObjectInput{
		ObjectPath: tfStatePath,
	}
	err = objClient.Delete(context.Background(), deleteObjInput)
	if err != nil {
		return err
	}

	// Deleting the cluster manager directory
	clusterManagerRootPath := fmt.Sprintf(clusterManagerRootPathFormat, clusterManagerName)
	deleteDirInput := &storage.DeleteDirectoryInput{
		DirectoryName: clusterManagerRootPath,
	}
	err = manta.tritonStorageClient.Dir().Delete(context.Background(), deleteDirInput)
	if err != nil {
		return err
	}

	return nil
}

func NewRemoteClusterManagerStateManta(tritonAccount, tritonKeyPath, tritonKeyID, tritonURL, mantaURL string) (RemoteClusterManagerStateManta, error) {
	keyMaterial, err := ioutil.ReadFile(tritonKeyPath)
	if err != nil {
		return RemoteClusterManagerStateManta{}, err
	}

	sshKeySigner, err := authentication.NewPrivateKeySigner(tritonKeyID, keyMaterial, tritonAccount)
	if err != nil {
		return RemoteClusterManagerStateManta{}, err
	}

	// Create manta client
	config := &triton.ClientConfig{
		TritonURL:   tritonURL,
		MantaURL:    mantaURL,
		AccountName: tritonAccount,
		Signers:     []authentication.Signer{sshKeySigner},
	}
	tritonStorageClient, err := storage.NewClient(config)
	if err != nil {
		return RemoteClusterManagerStateManta{}, err
	}

	return RemoteClusterManagerStateManta{tritonStorageClient}, nil
}
