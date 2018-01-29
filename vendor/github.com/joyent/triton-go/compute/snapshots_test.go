package compute_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"io/ioutil"

	"github.com/joyent/triton-go/compute"
	"github.com/joyent/triton-go/testutils"
)

const accountUrl = "testing"

var (
	listSnapshotErrorType             = errors.New("Error executing List request:")
	getSnapshotErrorType              = errors.New("Error executing Get request:")
	deleteSnapshotErrorType           = errors.New("Error executing Delete request:")
	createSnapshotErrorType           = errors.New("Error executing Create request:")
	startMachineFromSnapshotErrorType = errors.New("Error executing StartMachine request:")
)

func MockIdentityClient() *compute.ComputeClient {
	return &compute.ComputeClient{
		Client: testutils.NewMockClient(testutils.MockClientInput{
			AccountName: accountUrl,
		}),
	}
}

func TestListSnapshots(t *testing.T) {
	computeClient := MockIdentityClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) ([]*compute.Snapshot, error) {
		defer testutils.DeactivateClient()

		ping, err := cc.Snapshots().List(ctx, &compute.ListSnapshotsInput{
			MachineID: "123-3456-2335",
		})
		if err != nil {
			return nil, err
		}
		return ping, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", fmt.Sprintf("/%s/machines/%s/snapshots", accountUrl, "123-3456-2335"), listSnapshotsSuccess)

		resp, err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}

		if resp == nil {
			t.Fatalf("Expected an output but got nil")
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("GET", fmt.Sprintf("/%s/machines/%s/snapshots", accountUrl, "123-3456-2335"), listSnapshotEmpty)

		_, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %s", err)
		}
	})

	t.Run("bad_decode", func(t *testing.T) {
		testutils.RegisterResponder("GET", fmt.Sprintf("/%s/machines/%s/snapshots", accountUrl, "123-3456-2335"), listSnapshotBadDecode)

		_, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %s", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("GET", fmt.Sprintf("/%s/machines/%s/snapshots", accountUrl, "123-3456-2335"), listSnapshotError)

		resp, err := do(context.Background(), computeClient)
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

func TestGetSnapshot(t *testing.T) {
	computeClient := MockIdentityClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) (*compute.Snapshot, error) {
		defer testutils.DeactivateClient()

		snapshot, err := cc.Snapshots().Get(ctx, &compute.GetSnapshotInput{
			MachineID: "123-3456-2335",
			Name:      "sample-snapshot",
		})
		if err != nil {
			return nil, err
		}
		return snapshot, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", fmt.Sprintf("/%s/machines/%s/snapshots/%s", accountUrl, "123-3456-2335", "sample-snapshot"), getSnapshotSuccess)

		resp, err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}

		if resp == nil {
			t.Fatalf("Expected an output but got nil")
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("GET", fmt.Sprintf("/%s/machines/%s/snapshots/%s", accountUrl, "123-3456-2335", "sample-snapshot"), getSnapshotEmpty)

		_, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %s", err)
		}
	})

	t.Run("bad_decode", func(t *testing.T) {
		testutils.RegisterResponder("GET", fmt.Sprintf("/%s/machines/%s/snapshots/%s", accountUrl, "123-3456-2335", "sample-snapshot"), getSnapshotBadDecode)

		_, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %s", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("GET", fmt.Sprintf("/%s/machines/%s/snapshots/%s", accountUrl, "123-3456-2335", "sample-snapshot"), getSnapshotError)

		resp, err := do(context.Background(), computeClient)
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

func TestDeleteSnapshot(t *testing.T) {
	computeClient := MockIdentityClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) error {
		defer testutils.DeactivateClient()

		return cc.Snapshots().Delete(ctx, &compute.DeleteSnapshotInput{
			MachineID: "123-3456-2335",
			Name:      "sample-snapshot",
		})
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("DELETE", fmt.Sprintf("/%s/machines/%s/snapshots/%s", accountUrl, "123-3456-2335", "sample-snapshot"), deleteSnapshotSuccess)

		err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("DELETE", fmt.Sprintf("/%s/machines/%s/snapshots/%s", accountUrl, "123-3456-2335", "sample-snapshot"), deleteSnapshotError)

		err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "Error executing Delete request:") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestStartMachineFromSnapshot(t *testing.T) {
	computeClient := MockIdentityClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) error {
		defer testutils.DeactivateClient()

		return cc.Snapshots().StartMachine(ctx, &compute.StartMachineFromSnapshotInput{
			MachineID: "123-3456-2335",
			Name:      "sample-snapshot",
		})
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("POST", fmt.Sprintf("/%s/machines/%s/snapshots/%s", accountUrl, "123-3456-2335", "sample-snapshot"), startMachineFromSnapshotSuccess)

		err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("POST", fmt.Sprintf("/%s/machines/%s/snapshots/%s", accountUrl, "123-3456-2335", "sample-snapshot"), startMachineFromSnapshotError)

		err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "Error executing StartMachine request:") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestCreateSnapshot(t *testing.T) {
	computeClient := MockIdentityClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) (*compute.Snapshot, error) {
		defer testutils.DeactivateClient()

		snapshot, err := cc.Snapshots().Create(ctx, &compute.CreateSnapshotInput{
			MachineID: "123-3456-2335",
			Name:      "sample-snapshot",
		})
		if err != nil {
			return nil, err
		}
		return snapshot, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("POST", fmt.Sprintf("/%s/machines/%s/snapshots", accountUrl, "123-3456-2335"), createSnapshotSuccess)

		_, err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("POST", fmt.Sprintf("/%s/machines/%s/snapshots", accountUrl, "123-3456-2335"), createSnapshotError)

		_, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "Error executing Create request:") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func listSnapshotsSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`[
	{
    "name": "sample-snapshot",
	"state": "queued",
    "updated": "2015-12-23T06:41:11.032Z",
    "created": "2015-12-23T06:41:11.032Z"
  },
  {
    "name": "sample-snapshot-2",
	"state": "queued",
    "updated": "2015-12-23T06:41:11.032Z",
    "created": "2015-12-23T06:41:11.032Z"
  }
]`)
	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listSnapshotEmpty(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")
	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}, nil
}

func listSnapshotBadDecode(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`[{
    "name": "sample-snapshot",
	"state": "queued",
    "updated": "2015-12-23T06:41:11.032Z",
    "created": "2015-12-23T06:41:11.032Z",}]`)
	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listSnapshotError(req *http.Request) (*http.Response, error) {
	return nil, listSnapshotErrorType
}

func getSnapshotSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
    "name": "sample-snapshot",
	"state": "queued",
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

func getSnapshotBadDecode(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
    "name": "sample-snapshot",
	"state": "queued",
    "updated": "2015-12-23T06:41:11.032Z",
    "created": "2015-12-23T06:41:11.032Z",}`)
	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func getSnapshotEmpty(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")
	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}, nil
}

func getSnapshotError(req *http.Request) (*http.Response, error) {
	return nil, getSnapshotErrorType
}

func deleteSnapshotSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: 204,
		Header:     header,
	}, nil
}

func deleteSnapshotError(req *http.Request) (*http.Response, error) {
	return nil, deleteSnapshotErrorType
}

func startMachineFromSnapshotSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: 202,
		Header:     header,
	}, nil
}

func startMachineFromSnapshotError(req *http.Request) (*http.Response, error) {
	return nil, startMachineFromSnapshotErrorType
}

func createSnapshotSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
    "name": "sample-snapshot",
	"state": "queued",
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

func createSnapshotError(req *http.Request) (*http.Response, error) {
	return nil, createSnapshotErrorType
}
