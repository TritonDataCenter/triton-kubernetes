package network_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"io/ioutil"

	"github.com/joyent/triton-go/network"
	"github.com/joyent/triton-go/testutils"
)

const accountUrl = "testing"

var (
	listRuleMachinesErrorType = errors.New("Error executing ListRuleMachines request:")
)

func MockNetworkClient() *network.NetworkClient {
	return &network.NetworkClient{
		Client: testutils.NewMockClient(testutils.MockClientInput{
			AccountName: accountUrl,
		}),
	}
}

func TestListRuleMachines(t *testing.T) {
	networkClient := MockNetworkClient()

	do := func(ctx context.Context, nc *network.NetworkClient) ([]*network.Machine, error) {
		defer testutils.DeactivateClient()

		machines, err := nc.Firewall().ListRuleMachines(ctx, &network.ListRuleMachinesInput{
			ID: "38de17c4-39e8-48c7-a168-0f58083de860",
		})
		if err != nil {
			return nil, err
		}
		return machines, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", fmt.Sprintf("/%s/fwrules/%s/machines", accountUrl, "38de17c4-39e8-48c7-a168-0f58083de860"), listRuleMachinesSuccess)

		resp, err := do(context.Background(), networkClient)
		if err != nil {
			t.Fatal(err)
		}

		if resp == nil {
			t.Fatalf("Expected an output but got nil")
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("GET", fmt.Sprintf("/%s/fwrules/%s/machines", accountUrl, "38de17c4-39e8-48c7-a168-0f58083de860"), listRuleMachinesEmpty)

		_, err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %s", err)
		}
	})

	t.Run("bad_decode", func(t *testing.T) {
		testutils.RegisterResponder("GET", fmt.Sprintf("/%s/fwrules/%s/machines", accountUrl, "38de17c4-39e8-48c7-a168-0f58083de860"), listRuleMachinesBadDecode)

		_, err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %s", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("GET", fmt.Sprintf("/%s/fwrules/%s/machines", accountUrl, "38de17c4-39e8-48c7-a168-0f58083de860"), listRuleMachinesError)

		resp, err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}
		if resp != nil {
			t.Error("expected resp to be nil")
		}

		if !strings.Contains(err.Error(), "Error executing ListRuleMachines request:") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func listRuleMachinesEmpty(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")
	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}, nil
}

func listRuleMachinesSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`[
	{
    "id": "b6979942-7d5d-4fe6-a2ec-b812e950625a",
    "name": "test",
    "type": "smartmachine",
    "brand": "joyent",
    "state": "running",
    "image": "2b683a82-a066-11e3-97ab-2faa44701c5a",
    "ips": [
      "10.88.88.26",
      "192.168.128.5"
    ],
    "memory": 128,
    "disk": 12288,
    "metadata": {
      "root_authorized_keys": "test-key"
    },
    "tags": {},
    "created": "2016-01-04T12:55:50.539Z",
    "updated": "2016-01-21T08:56:59.000Z",
    "networks": [
      "a9c130da-e3ba-40e9-8b18-112aba2d3ba7",
      "45607081-4cd2-45c8-baf7-79da760fffaa"
    ],
    "primaryIp": "10.88.88.26",
    "firewall_enabled": false,
    "compute_node": "564d0b8e-6099-7648-351e-877faf6c56f6",
    "package": "sdc_128"
  }
]`)
	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listRuleMachinesBadDecode(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`[
	{
    "id": "b6979942-7d5d-4fe6-a2ec-b812e950625a",
    "name": "test",
    "type": "smartmachine",
    "brand": "joyent",
    "state": "running",
    "image": "2b683a82-a066-11e3-97ab-2faa44701c5a",
    "ips": [
      "10.88.88.26",
      "192.168.128.5"
    ],
    "memory": 128,
    "disk": 12288,
    "metadata": {
      "root_authorized_keys": "test-key"
    },
    "tags": {},
    "created": "2016-01-04T12:55:50.539Z",
    "updated": "2016-01-21T08:56:59.000Z",
    "networks": [
      "a9c130da-e3ba-40e9-8b18-112aba2d3ba7",
      "45607081-4cd2-45c8-baf7-79da760fffaa"
    ],
    "primaryIp": "10.88.88.26",
    "firewall_enabled": false,
    "compute_node": "564d0b8e-6099-7648-351e-877faf6c56f6",
    "package": "sdc_128",
  }
]`)
	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listRuleMachinesError(req *http.Request) (*http.Response, error) {
	return nil, listRuleMachinesErrorType
}
