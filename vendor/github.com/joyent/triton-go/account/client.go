package account

import (
	triton "github.com/joyent/triton-go"
	"github.com/joyent/triton-go/client"
)

type AccountClient struct {
	Client *client.Client
}

func newAccountClient(client *client.Client) *AccountClient {
	return &AccountClient{
		Client: client,
	}
}

// NewClient returns a new client for working with Account endpoints and
// resources within CloudAPI
func NewClient(config *triton.ClientConfig) (*AccountClient, error) {
	// TODO: Utilize config interface within the function itself
	client, err := client.New(config.TritonURL, config.MantaURL, config.AccountName, config.Signers...)
	if err != nil {
		return nil, err
	}
	return newAccountClient(client), nil
}

// Config returns a c used for accessing functions pertaining
// to Config functionality in the Triton API.
func (c *AccountClient) Config() *ConfigClient {
	return &ConfigClient{c.Client}
}

// Keys returns a Compute client used for accessing functions pertaining to SSH
// key functionality in the Triton API.
func (c *AccountClient) Keys() *KeysClient {
	return &KeysClient{c.Client}
}
