package identity

import (
	triton "github.com/joyent/triton-go"
	"github.com/joyent/triton-go/client"
)

type IdentityClient struct {
	Client *client.Client
}

func newIdentityClient(client *client.Client) *IdentityClient {
	return &IdentityClient{
		Client: client,
	}
}

// NewClient returns a new client for working with Identity endpoints and
// resources within CloudAPI
func NewClient(config *triton.ClientConfig) (*IdentityClient, error) {
	// TODO: Utilize config interface within the function itself
	client, err := client.New(config.TritonURL, config.MantaURL, config.AccountName, config.Signers...)
	if err != nil {
		return nil, err
	}
	return newIdentityClient(client), nil
}

// Roles returns a Roles client used for accessing functions pertaining to
// Role functionality in the Triton API.
func (c *IdentityClient) Roles() *RolesClient {
	return &RolesClient{c.Client}
}

// Users returns a Users client used for accessing functions pertaining to
// User functionality in the Triton API.
func (c *IdentityClient) Users() *UsersClient {
	return &UsersClient{c.Client}
}
