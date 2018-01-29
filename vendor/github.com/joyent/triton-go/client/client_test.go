package client

import (
	"os"
	"strings"
	"testing"

	auth "github.com/joyent/triton-go/authentication"
)

const BadURL = "**ftp://man($$"

func TestNew(t *testing.T) {
	tritonURL := "http://triton.test.org"
	mantaURL := "http://manta.test.org"
	accountName := "test.user"
	signer, _ := auth.NewTestSigner()

	tests := []struct {
		name        string
		tritonURL   string
		mantaURL    string
		accountName string
		signer      auth.Signer
		err         interface{}
	}{
		{"default", tritonURL, mantaURL, accountName, signer, nil},
		{"missing url", "", "", accountName, signer, ErrMissingURL},
		{"bad tritonURL", BadURL, mantaURL, accountName, signer, BadTritonURL},
		{"bad mantaURL", tritonURL, BadURL, accountName, signer, BadMantaURL},
		{"missing accountName", tritonURL, mantaURL, "", signer, ErrAccountName},
		{"missing signer", tritonURL, mantaURL, accountName, nil, ErrDefaultAuth},
	}

	for _, test := range tests {
		os.Unsetenv("TRITON_KEY_ID")
		os.Unsetenv("SDC_KEY_ID")
		os.Unsetenv("MANTA_KEY_ID")
		os.Unsetenv("SSH_AUTH_SOCK")

		t.Run(test.name, func(t *testing.T) {
			_, err := New(
				test.tritonURL,
				test.mantaURL,
				test.accountName,
				test.signer,
			)
			if test.err != nil {
				if err == nil {
					t.Error("expected error not to be nil")
					return
				}

				switch test.err.(type) {
				case error:
					testErr := test.err.(error)
					if err.Error() != testErr.Error() {
						t.Errorf("expected error: received %v", err)
					}
				case string:
					testErr := test.err.(string)
					if !strings.Contains(err.Error(), testErr) {
						t.Errorf("expected error: received %v", err)
					}
				}
				return
			}
			if err != nil {
				t.Errorf("expected error to be nil: received %v", err)
			}
		})
	}

	t.Run("default SSH agent auth", func(t *testing.T) {
		os.Unsetenv("SSH_AUTH_SOCK")
		err := os.Setenv("TRITON_KEY_ID", auth.Dummy.Fingerprint)
		defer os.Unsetenv("TRITON_KEY_ID")
		if err != nil {
			t.Errorf("expected error to not be nil: received %v", err)
		}

		_, err = New(
			tritonURL,
			mantaURL,
			accountName,
			nil,
		)
		if err == nil {
			t.Error("expected error to not be nil")
		}
		if !strings.Contains(err.Error(), "problem initializing NewSSHAgentSigner") {
			t.Errorf("expected error to be from NewSSHAgentSigner: received '%v'", err)
		}
	})
}
