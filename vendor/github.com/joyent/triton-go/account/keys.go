package account

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/errwrap"
	"github.com/joyent/triton-go/client"
)

type KeysClient struct {
	client *client.Client
}

// Key represents a public key
type Key struct {
	// Name of the key
	Name string `json:"name"`

	// Key fingerprint
	Fingerprint string `json:"fingerprint"`

	// OpenSSH-formatted public key
	Key string `json:"key"`
}

type ListKeysInput struct{}

// ListKeys lists all public keys we have on record for the specified
// account.
func (c *KeysClient) List(ctx context.Context, _ *ListKeysInput) ([]*Key, error) {
	path := fmt.Sprintf("/%s/keys", c.client.AccountName)
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   path,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errwrap.Wrapf("Error executing ListKeys request: {{err}}", err)
	}

	var result []*Key
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding ListKeys response: {{err}}", err)
	}

	return result, nil
}

type GetKeyInput struct {
	KeyName string
}

func (c *KeysClient) Get(ctx context.Context, input *GetKeyInput) (*Key, error) {
	path := fmt.Sprintf("/%s/keys/%s", c.client.AccountName, input.KeyName)
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   path,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errwrap.Wrapf("Error executing GetKey request: {{err}}", err)
	}

	var result *Key
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding GetKey response: {{err}}", err)
	}

	return result, nil
}

type DeleteKeyInput struct {
	KeyName string
}

func (c *KeysClient) Delete(ctx context.Context, input *DeleteKeyInput) error {
	path := fmt.Sprintf("/%s/keys/%s", c.client.AccountName, input.KeyName)
	reqInputs := client.RequestInput{
		Method: http.MethodDelete,
		Path:   path,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return errwrap.Wrapf("Error executing DeleteKey request: {{err}}", err)
	}

	return nil
}

// CreateKeyInput represents the option that can be specified
// when creating a new key.
type CreateKeyInput struct {
	// Name of the key. Optional.
	Name string `json:"name,omitempty"`

	// OpenSSH-formatted public key.
	Key string `json:"key"`
}

// CreateKey uploads a new OpenSSH key to Triton for use in HTTP signing and SSH.
func (c *KeysClient) Create(ctx context.Context, input *CreateKeyInput) (*Key, error) {
	path := fmt.Sprintf("/%s/keys", c.client.AccountName)
	reqInputs := client.RequestInput{
		Method: http.MethodPost,
		Path:   path,
		Body:   input,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errwrap.Wrapf("Error executing CreateKey request: {{err}}", err)
	}

	var result *Key
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding CreateKey response: {{err}}", err)
	}

	return result, nil
}
