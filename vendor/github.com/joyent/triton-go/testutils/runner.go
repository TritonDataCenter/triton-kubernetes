package testutils

import (
	"errors"
	"fmt"
	"log"
	"os"
	"testing"

	triton "github.com/joyent/triton-go"
	"github.com/joyent/triton-go/authentication"
	"github.com/joyent/triton-go/client"
)

const TestEnvVar = "TRITON_TEST"

type TestCase struct {
	Steps []Step
	State TritonStateBag
}

func AccTest(t *testing.T, c TestCase) {
	// We only run acceptance tests if an env var is set because they're
	// slow and generally require some outside configuration.
	if os.Getenv(TestEnvVar) == "" {
		t.Skip(fmt.Sprintf(
			"Acceptance tests skipped unless env '%s' set",
			TestEnvVar))
		return
	}

	// We require verbose mode so that the user knows what is going on.
	if !testing.Verbose() {
		t.Fatal("Acceptance tests must be run with the -v flag on tests")
		return
	}

	tritonURL := client.GetTritonEnv("URL")
	tritonAccount := client.GetTritonEnv("ACCOUNT")
	tritonKeyID := client.GetTritonEnv("KEY_ID")
	tritonKeyMaterial := client.GetTritonEnv("KEY_MATERIAL")
	userName := client.GetTritonEnv("USER")
	mantaURL := client.GetTritonEnv("MANTA_URL")

	var prerollErrors []error
	if tritonURL == "" {
		prerollErrors = append(prerollErrors,
			errors.New("The TRITON_URL environment variable must be set to run acceptance tests"))
	}
	if tritonAccount == "" {
		prerollErrors = append(prerollErrors,
			errors.New("The TRITON_ACCOUNT environment variable must be set to run acceptance tests"))
	}
	if tritonKeyID == "" {
		prerollErrors = append(prerollErrors,
			errors.New("The TRITON_KEY_ID environment variable must be set to run acceptance tests"))
	}
	if len(prerollErrors) > 0 {
		for _, err := range prerollErrors {
			t.Error(err)
		}
		t.FailNow()
	}

	var signer authentication.Signer
	var err error
	if tritonKeyMaterial != "" {
		log.Println("[INFO] Creating Triton Client with Private Key Signer...")
		input := authentication.PrivateKeySignerInput{
			KeyID:              tritonKeyID,
			PrivateKeyMaterial: []byte(tritonKeyMaterial),
			AccountName:        tritonAccount,
			Username:           userName,
		}
		signer, err = authentication.NewPrivateKeySigner(input)
		if err != nil {
			t.Fatalf("Error creating private key signer: %s", err)
		}
	} else {
		log.Println("[INFO] Creating Triton Client with SSH Key Signer...")
		input := authentication.SSHAgentSignerInput{
			KeyID:       tritonKeyID,
			AccountName: tritonAccount,
			Username:    userName,
		}
		signer, err = authentication.NewSSHAgentSigner(input)
		if err != nil {
			t.Fatalf("Error creating SSH Agent signer: %s", err)
		}
	}

	// Old world... we spun up a universal client. This is pushed deeper into
	// the process within `testutils.StepClient`.
	//
	// client, err := NewClient(tritonURL, tritonAccount, signer)
	// if err != nil {
	//         t.Fatalf("Error creating Triton Client: %s", err)
	// }

	config := &triton.ClientConfig{
		TritonURL:   tritonURL,
		MantaURL:    mantaURL,
		AccountName: tritonAccount,
		Username:    userName,
		Signers:     []authentication.Signer{signer},
	}

	state := &basicTritonStateBag{
		TritonConfig: config,
	}

	runner := &basicRunner{
		Steps: c.Steps,
	}

	runner.Run(state)

	if errs := state.ErrorsOrNil(); errs != nil {
		for _, err := range errs {
			t.Error(err)
		}
		t.Fatal("\n\nThere may be dangling resources in your Triton account!")
	}
}
