package identity_test

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/joyent/triton-go/identity"
	"github.com/joyent/triton-go/testutils"
)

const accountUrl = "testing"

var (
	listUserErrorType   = errors.New("Error executing List request:")
	getUserErrorType    = errors.New("Error executing Get request:")
	deleteUserErrorType = errors.New("Error executing Delete request:")
	createUserErrorType = errors.New("Error executing Create request:")
	updateUserErrorType = errors.New("Error executing Update request:")
)

func MockIdentityClient() *identity.IdentityClient {
	return &identity.IdentityClient{
		Client: testutils.NewMockClient(testutils.MockClientInput{
			AccountName: accountUrl,
		}),
	}
}

func TestListUsers(t *testing.T) {
	identityClient := MockIdentityClient()

	do := func(ctx context.Context, ic *identity.IdentityClient) ([]*identity.User, error) {
		defer testutils.DeactivateClient()

		ping, err := ic.Users().List(ctx, &identity.ListUsersInput{})
		if err != nil {
			return nil, err
		}
		return ping, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", fmt.Sprintf("/%s/users", accountUrl), listUsersSuccess)

		resp, err := do(context.Background(), identityClient)
		if err != nil {
			t.Fatal(err)
		}

		if resp == nil {
			t.Fatalf("Expected an output but got nil")
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("GET", fmt.Sprintf("/%s/users", accountUrl), listUsersEmpty)

		_, err := do(context.Background(), identityClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %s", err)
		}
	})

	t.Run("bad_decode", func(t *testing.T) {
		testutils.RegisterResponder("GET", fmt.Sprintf("/%s/users", accountUrl), listUsersBadeDecode)

		_, err := do(context.Background(), identityClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %s", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("GET", fmt.Sprintf("/%s/users", accountUrl), listUserError)

		resp, err := do(context.Background(), identityClient)
		if err == nil {
			t.Fatal(err)
		}
		if resp != nil {
			t.Error("expected resp to be nil")
		}

		if !strings.Contains(err.Error(), "Error executing List request:") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestGetUser(t *testing.T) {
	identityClient := MockIdentityClient()

	do := func(ctx context.Context, ic *identity.IdentityClient) (*identity.User, error) {
		defer testutils.DeactivateClient()

		user, err := ic.Users().Get(ctx, &identity.GetUserInput{
			UserID: "123-3456-2335",
		})
		if err != nil {
			return nil, err
		}
		return user, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", fmt.Sprintf("/%s/users/%s", accountUrl, "123-3456-2335"), getUserSuccess)

		resp, err := do(context.Background(), identityClient)
		if err != nil {
			t.Fatal(err)
		}

		if resp == nil {
			t.Fatalf("Expected an output but got nil")
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("GET", fmt.Sprintf("/%s/users/%s", accountUrl, "123-3456-2335"), getUserEmpty)

		_, err := do(context.Background(), identityClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %s", err)
		}
	})

	t.Run("bad_decode", func(t *testing.T) {
		testutils.RegisterResponder("GET", fmt.Sprintf("/%s/users/%s", accountUrl, "123-3456-2335"), getUserBadeDecode)

		_, err := do(context.Background(), identityClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %s", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("GET", fmt.Sprintf("/%s/users", accountUrl), getUserError)

		resp, err := do(context.Background(), identityClient)
		if err == nil {
			t.Fatal(err)
		}
		if resp != nil {
			t.Error("expected resp to be nil")
		}

		if !strings.Contains(err.Error(), "Error executing Get request:") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestDeleteUser(t *testing.T) {
	identityClient := MockIdentityClient()

	do := func(ctx context.Context, ic *identity.IdentityClient) error {
		defer testutils.DeactivateClient()

		return ic.Users().Delete(ctx, &identity.DeleteUserInput{
			UserID: "123-3456-2335",
		})
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("DELETE", fmt.Sprintf("/%s/users/%s", accountUrl, "123-3456-2335"), deleteUserSuccess)

		err := do(context.Background(), identityClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("DELETE", fmt.Sprintf("/%s/users", accountUrl), deleteUserError)

		err := do(context.Background(), identityClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "Error executing Delete request:") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestCreateUser(t *testing.T) {
	identityClient := MockIdentityClient()

	do := func(ctx context.Context, ic *identity.IdentityClient) (*identity.User, error) {
		defer testutils.DeactivateClient()

		user, err := ic.Users().Create(ctx, &identity.CreateUserInput{
			Email:    "fake@fake.com",
			Login:    "testuser",
			Password: "Password123",
		})

		if err != nil {
			return nil, err
		}
		return user, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("POST", fmt.Sprintf("/%s/users", accountUrl), createUserSuccess)

		_, err := do(context.Background(), identityClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("POST", fmt.Sprintf("/%s/users", accountUrl), createUserError)

		_, err := do(context.Background(), identityClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "Error executing Create request:") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestUpdateUser(t *testing.T) {
	identityClient := MockIdentityClient()

	do := func(ctx context.Context, ic *identity.IdentityClient) (*identity.User, error) {
		defer testutils.DeactivateClient()

		user, err := ic.Users().Update(ctx, &identity.UpdateUserInput{
			UserID: "123-3456-2335",
			Login:  "testuser1",
		})
		if err != nil {
			return nil, err
		}
		return user, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("POST", fmt.Sprintf("/%s/users/%s", accountUrl, "123-3456-2335"), updateUserSuccess)

		_, err := do(context.Background(), identityClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("POST", fmt.Sprintf("/%s/users/%s", accountUrl, "123-3456-2335"), updateUserError)

		_, err := do(context.Background(), identityClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "Error executing Update request:") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func getUserSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
    "id": "4fc13ac6-1e7d-cd79-f3d2-96276af0d638",
    "login": "barbar",
    "email": "barbar@example.com",
    "companyName": "Example",
    "firstName": "BarBar",
    "lastName": "Jinks",
    "phone": "(123)457-6890",
    "updated": "2015-12-23T06:41:11.032Z",
    "created": "2015-12-23T06:41:11.032Z"
  }
`)
	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func getUserError(req *http.Request) (*http.Response, error) {
	return nil, getUserErrorType
}

func getUserBadeDecode(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
    "id": "4fc13ac6-1e7d-cd79-f3d2-96276af0d638",
    "login": "barbar",
    "email": "barbar@example.com",
    "companyName": "Example",
    "firstName": "BarBar",
    "lastName": "Jinks",
    "phone": "(123)457-6890",
    "updated": "2015-12-23T06:41:11.032Z",
    "created": "2015-12-23T06:41:11.032Z",}`)
	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func getUserEmpty(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")
	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}, nil
}

func listUsersEmpty(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")
	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}, nil
}

func listUsersSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`[
	{
    "id": "4fc13ac6-1e7d-cd79-f3d2-96276af0d638",
    "login": "barbar",
    "email": "barbar@example.com",
    "companyName": "Example",
    "firstName": "BarBar",
    "lastName": "Jinks",
    "phone": "(123)457-6890",
    "updated": "2015-12-23T06:41:11.032Z",
    "created": "2015-12-23T06:41:11.032Z"
  },
  {
    "id": "332ce629-fcc5-45c3-e34f-e7cfbeab1327",
    "login": "san",
    "email": "san@example.com",
    "companyName": "Example Inc",
    "firstName": "San",
    "lastName": "Holo",
    "phone": "(123)456-0987",
    "updated": "2015-12-23T06:41:56.102Z",
    "created": "2015-12-23T06:41:56.102Z"
  }
]`)
	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listUsersBadeDecode(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{[
	{
    "id": "4fc13ac6-1e7d-cd79-f3d2-96276af0d638",
    "login": "barbar",
    "email": "barbar@example.com",
    "companyName": "Example",
    "firstName": "BarBar",
    "lastName": "Jinks",
    "phone": "(123)457-6890",
    "updated": "2015-12-23T06:41:11.032Z",
    "created": "2015-12-23T06:41:11.032Z"
  },
  {
    "id": "332ce629-fcc5-45c3-e34f-e7cfbeab1327",
    "login": "san",
    "email": "san@example.com",
    "companyName": "Example Inc",
    "firstName": "San",
    "lastName": "Holo",
    "phone": "(123)456-0987",
    "updated": "2015-12-23T06:41:56.102Z",
    "created": "2015-12-23T06:41:56.102Z"
  }
]}`)
	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listUserError(req *http.Request) (*http.Response, error) {
	return nil, listUserErrorType
}

func deleteUserSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: 204,
		Header:     header,
	}, nil
}

func deleteUserError(req *http.Request) (*http.Response, error) {
	return nil, deleteUserErrorType
}

func createUserSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
    "id": "123-3456-2335",
    "login": "testuser",
    "email": "barbar@example.com",
    "updated": "2015-12-23T06:41:11.032Z",
    "created": "2015-12-23T06:41:11.032Z"
  }
`)

	return &http.Response{
		StatusCode: 201,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func createUserError(req *http.Request) (*http.Response, error) {
	return nil, createUserErrorType
}

func updateUserSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
    "id": "123-3456-2335",
    "login": "testuser",
    "email": "barbar@example.com",
    "updated": "2015-12-23T06:41:11.032Z",
    "created": "2015-12-23T06:41:11.032Z"
  }
`)

	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func updateUserError(req *http.Request) (*http.Response, error) {
	return nil, updateUserErrorType
}
