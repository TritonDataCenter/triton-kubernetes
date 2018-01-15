package manta

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/joyent/triton-kubernetes/backend"
	"github.com/joyent/triton-kubernetes/state"

	triton "github.com/joyent/triton-go"
	"github.com/joyent/triton-go/authentication"
	"github.com/joyent/triton-go/storage"
)

const (
	rootDirectory             = "/stor/triton-kubernetes"
	rootPathFormat            = rootDirectory + "/%s"
	terraformConfigPathFormat = rootDirectory + "/%s/main.tf.json"
	terraformStatePathFormat  = rootDirectory + "/%s/terraform.tfstate"

	terraformBackendRootPathFormat = "/triton-kubernetes/%s"
)

// Stores terraform json configuration files for all cluster managers in Manta
// Each cluster manager has a separate directory under triton-kubernetes with a main.tf.json file
// and a terraform.tfstate file.
// triton-kubernetes manages the main.tf.json file and terraform manages the terraform.tfstate file
// Directory Path: /stor/triton-kubernetes/${CLUSTER_MANAGER_NAME}/main.tf.json
// TODO: Lock terraform json configuration similar to how terraform locks tfstate file.
type mantaBackend struct {
	tritonAccount string
	tritonKeyPath string
	tritonKeyID   string
	tritonURL     string
	mantaURL      string

	tritonStorageClient *storage.StorageClient
}

type mantaTerraformBackendConfig struct {
	Account     string `json:"account"`
	KeyMaterial string `json:"key_material"`
	KeyID       string `json:"key_id"`
	Path        string `json:"path"`
}

func New(tritonAccount, tritonKeyPath, tritonKeyID, tritonURL, mantaURL string) (backend.Backend, error) {
	keyMaterial, err := ioutil.ReadFile(tritonKeyPath)
	if err != nil {
		return nil, err
	}

	sshKeySigner, err := authentication.NewPrivateKeySigner(tritonKeyID, keyMaterial, tritonAccount)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	// Create root directory if it doesn't exist
	putDirInput := &storage.PutDirectoryInput{
		DirectoryName: rootDirectory,
	}
	err = tritonStorageClient.Dir().Put(context.Background(), putDirInput)
	if err != nil {
		return nil, err
	}

	return &mantaBackend{
		tritonAccount:       tritonAccount,
		tritonKeyPath:       tritonKeyPath,
		tritonKeyID:         tritonKeyID,
		tritonURL:           tritonURL,
		mantaURL:            mantaURL,
		tritonStorageClient: tritonStorageClient,
	}, nil
}

func (backend *mantaBackend) States() ([]string, error) {
	input := storage.ListDirectoryInput{
		DirectoryName: rootDirectory,
		Limit:         100,
	}

	result, err := backend.tritonStorageClient.Dir().List(context.Background(), &input)
	if err != nil {
		return nil, err
	}

	if result.ResultSetSize == 0 {
		return []string{}, nil
	}

	states := []string{}
	for _, state := range result.Entries {
		states = append(states, state.Name)
	}

	return states, nil
}

func (backend *mantaBackend) State(name string) (state.State, error) {
	terraformConfigPath := fmt.Sprintf(terraformConfigPathFormat, name)

	getObjectInput := &storage.GetObjectInput{
		ObjectPath: terraformConfigPath,
	}
	output, err := backend.tritonStorageClient.Objects().Get(context.Background(), getObjectInput)
	if err != nil {
		// TODO: Find a better way to determine this error
		if strings.Contains(err.Error(), "ResourceNotFound") {
			// Since no state exists, lets create an empty one
			return state.New(name, []byte("{}"))
		}
		return state.State{}, err
	}

	currentConfigBytes, err := ioutil.ReadAll(output.ObjectReader)
	if err != nil {
		return state.State{}, err
	}

	return state.New(name, currentConfigBytes)
}

func (backend *mantaBackend) PersistState(state state.State) error {
	terraformConfigPath := fmt.Sprintf(terraformConfigPathFormat, state.Name)

	objInput := storage.PutObjectInput{
		ObjectPath:   terraformConfigPath,
		ContentType:  "application/json",
		ObjectReader: bytes.NewReader(state.Bytes()),
	}
	err := backend.tritonStorageClient.Objects().Put(context.Background(), &objInput)
	if err != nil {
		return err
	}

	return nil
}

func (backend *mantaBackend) DeleteState(name string) error {
	objClient := backend.tritonStorageClient.Objects()

	// Deleting the main.tf.json file
	terraformConfigPath := fmt.Sprintf(terraformConfigPathFormat, name)
	deleteObjInput := &storage.DeleteObjectInput{
		ObjectPath: terraformConfigPath,
	}
	err := objClient.Delete(context.Background(), deleteObjInput)
	if err != nil {
		return err
	}

	// Deleting the terraform.tfstate file
	terraformStatePath := fmt.Sprintf(terraformStatePathFormat, name)
	deleteObjInput = &storage.DeleteObjectInput{
		ObjectPath: terraformStatePath,
	}
	err = objClient.Delete(context.Background(), deleteObjInput)
	if err != nil {
		return err
	}

	// Deleting the directory
	rootPath := fmt.Sprintf(rootPathFormat, name)
	deleteDirInput := &storage.DeleteDirectoryInput{
		DirectoryName: rootPath,
	}
	err = backend.tritonStorageClient.Dir().Delete(context.Background(), deleteDirInput)
	if err != nil {
		return err
	}

	return nil
}

func (backend *mantaBackend) StateTerraformConfig(name string) (string, interface{}) {
	terraformBackendConfig := mantaTerraformBackendConfig{
		Account:     backend.tritonAccount,
		KeyMaterial: backend.tritonKeyPath,
		KeyID:       backend.tritonKeyID,
		Path:        fmt.Sprintf(terraformBackendRootPathFormat, name),
	}

	return "terraform.backend.manta", terraformBackendConfig
}
