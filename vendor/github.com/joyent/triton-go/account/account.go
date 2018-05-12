package account

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/joyent/triton-go/client"
)

type Account struct {
	ID               string    `json:"id"`
	Login            string    `json:"login"`
	Email            string    `json:"email"`
	CompanyName      string    `json:"companyName"`
	FirstName        string    `json:"firstName"`
	LastName         string    `json:"lastName"`
	Address          string    `json:"address"`
	PostalCode       string    `json:"postalCode"`
	City             string    `json:"city"`
	State            string    `json:"state"`
	Country          string    `json:"country"`
	Phone            string    `json:"phone"`
	Created          time.Time `json:"created"`
	Updated          time.Time `json:"updated"`
	TritonCNSEnabled bool      `json:"triton_cns_enabled"`
}

type GetInput struct{}

func (c AccountClient) Get(ctx context.Context, input *GetInput) (*Account, error) {
	path := fmt.Sprintf("/%s", c.Client.AccountName)
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   path,
	}
	respReader, err := c.Client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errwrap.Wrapf("Error executing GetAccount request: {{err}}", err)
	}

	var result *Account
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding GetAccount response: {{err}}", err)
	}

	return result, nil
}

type UpdateInput struct {
	Email            string `json:"email,omitempty"`
	CompanyName      string `json:"companyName,omitempty"`
	FirstName        string `json:"firstName,omitempty"`
	LastName         string `json:"lastName,omitempty"`
	Address          string `json:"address,omitempty"`
	PostalCode       string `json:"postalCode,omitempty"`
	City             string `json:"city,omitempty"`
	State            string `json:"state,omitempty"`
	Country          string `json:"country,omitempty"`
	Phone            string `json:"phone,omitempty"`
	TritonCNSEnabled bool   `json:"triton_cns_enabled,omitempty"`
}

// UpdateAccount updates your account details with the given parameters.
// TODO(jen20) Work out a safe way to test this
func (c AccountClient) Update(ctx context.Context, input *UpdateInput) (*Account, error) {
	path := fmt.Sprintf("/%s", c.Client.AccountName)
	reqInputs := client.RequestInput{
		Method: http.MethodPost,
		Path:   path,
		Body:   input,
	}
	respReader, err := c.Client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errwrap.Wrapf("Error executing UpdateAccount request: {{err}}", err)
	}

	var result *Account
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding UpdateAccount response: {{err}}", err)
	}

	return result, nil
}
