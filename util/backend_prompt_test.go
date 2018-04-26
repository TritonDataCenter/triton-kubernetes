package util

import (
	"testing"
	"github.com/spf13/viper"
)


func TestBackendPromptWithUnsupportedBackendProviderNonInteractiveMode(t *testing.T) {
	viper.Set("non-interactive", true)
	viper.Set("backend_provider", "S3")

	_,err:=PromptForBackend()

	expected:= "Unsupported backend provider 'S3'"

	if err.Error() != expected {
		t.Error(err)
	}
}


func TestBackendPromptWithNoTritonAccountNonInteractiveMode(t *testing.T) {
	viper.Set("non-interactive", true)
	viper.Set("backend_provider", "manta")

	_,err:=PromptForBackend()

	expected:= "triton_account must be specified"

	if err.Error() != expected {
		t.Error(err)
	}
}

func TestBackendPromptWithNoTritonSSHKeyPathNonInteractiveMode(t *testing.T) {
	viper.Set("triton_account", "xyz")

	_,err:=PromptForBackend()

	expected:= "triton_key_path must be specified"

	if err.Error() != expected {
		t.Error(err)
	}
}

func TestBackendPromptWithInvalidTritonSSHKeyPathInteractive(t *testing.T) {
	viper.Set("non-interactive", false)
	viper.Set("triton_key_path", "")

	_,err:=PromptForBackend()

	expected:= "Unable to read private key: open : no such file or directory"

	if err.Error() != expected {
		t.Error(err)
	}
}

func TestNoTritonURLForNonInteractiveMode (t *testing.T) {
	viper.Set("non-interactive", true)
	viper.Set("triton_key_path", "")
	viper.Set("triton_key_id", "")

	_,err := PromptForBackend()

	expected := "triton_url must be specified"

	if err.Error() != expected {
		t.Error(err)
	}




}

func TestNoMantaURLForNonInteractiveMode (t *testing.T) {
	viper.Set("triton_url", "xyz.triton.com")

	_,err := PromptForBackend()

	expected := "manta_url must be specified"

	if err.Error() != expected {
		t.Error(err)
	}

}