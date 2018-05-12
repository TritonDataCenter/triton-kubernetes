package identity

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/joyent/triton-go/client"
)

type UsersClient struct {
	Client *client.Client
}

type User struct {
	ID           string    `json:"id"`
	Login        string    `json:"login"`
	EmailAddress string    `json:"email"`
	CompanyName  string    `json:"companyName"`
	FirstName    string    `json:"firstName"`
	LastName     string    `json:"lastName"`
	Address      string    `json:"address"`
	PostalCode   string    `json:"postCode"`
	City         string    `json:"city"`
	State        string    `json:"state"`
	Country      string    `json:"country"`
	Phone        string    `json:"phone"`
	Roles        []string  `json:"roles"`
	DefaultRoles []string  `json:"defaultRoles"`
	CreatedAt    time.Time `json:"created"`
	UpdatedAt    time.Time `json:"updated"`
}

type ListUsersInput struct{}

func (c *UsersClient) List(ctx context.Context, _ *ListUsersInput) ([]*User, error) {
	path := fmt.Sprintf("/%s/users", c.Client.AccountName)
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   path,
	}
	respReader, err := c.Client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errwrap.Wrapf("Error executing List request: {{err}}", err)
	}

	var result []*User
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding List response: {{err}}", err)
	}

	return result, nil
}

type GetUserInput struct {
	UserID string
}

func (c *UsersClient) Get(ctx context.Context, input *GetUserInput) (*User, error) {
	path := fmt.Sprintf("/%s/users/%s", c.Client.AccountName, input.UserID)
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   path,
	}
	respReader, err := c.Client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errwrap.Wrapf("Error executing Get request: {{err}}", err)
	}

	var result *User
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding Get response: {{err}}", err)
	}

	return result, nil
}

type DeleteUserInput struct {
	UserID string
}

func (c *UsersClient) Delete(ctx context.Context, input *DeleteUserInput) error {
	path := fmt.Sprintf("/%s/users/%s", c.Client.AccountName, input.UserID)
	reqInputs := client.RequestInput{
		Method: http.MethodDelete,
		Path:   path,
	}
	respReader, err := c.Client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return errwrap.Wrapf("Error executing Delete request: {{err}}", err)
	}

	return nil
}

type CreateUserInput struct {
	Email       string `json:"email"`
	Login       string `json:"login"`
	Password    string `json:"password"`
	CompanyName string `json:"companyName,omitempty"`
	FirstName   string `json:"firstName,omitempty"`
	LastName    string `json:"lastName,omitempty"`
	Address     string `json:"address,omitempty"`
	PostalCode  string `json:"postalCode,omitempty"`
	City        string `json:"city,omitempty"`
	State       string `json:"state,omitempty"`
	Country     string `json:"country,omitempty"`
	Phone       string `json:"phone,omitempty"`
}

func (c *UsersClient) Create(ctx context.Context, input *CreateUserInput) (*User, error) {
	path := fmt.Sprintf("/%s/users", c.Client.AccountName)
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
		return nil, errwrap.Wrapf("Error executing Create request: {{err}}", err)
	}

	var result *User
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding Create response: {{err}}", err)
	}

	return result, nil
}

type UpdateUserInput struct {
	UserID      string
	Email       string `json:"email,omitempty"`
	Login       string `json:"login,omitempty"`
	CompanyName string `json:"companyName,omitempty"`
	FirstName   string `json:"firstName,omitempty"`
	LastName    string `json:"lastName,omitempty"`
	Address     string `json:"address,omitempty"`
	PostalCode  string `json:"postalCode,omitempty"`
	City        string `json:"city,omitempty"`
	State       string `json:"state,omitempty"`
	Country     string `json:"country,omitempty"`
	Phone       string `json:"phone,omitempty"`
}

func (c *UsersClient) Update(ctx context.Context, input *UpdateUserInput) (*User, error) {
	path := fmt.Sprintf("/%s/users/%s", c.Client.AccountName, input.UserID)
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
		return nil, errwrap.Wrapf("Error executing Update request: {{err}}", err)
	}

	var result *User
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding Update response: {{err}}", err)
	}

	return result, nil
}
