package util

import (
	"testing"
	"github.com/spf13/viper"
)


func TestBackendPromptWithUnsupportedBackendProviderNonInteractiveMode(t *testing.T) {
	viper.Set("non-interactive", true)
	viper.Set("backend_provider", "S3")

	defer viper.Reset()

	_,err:=PromptForBackend()

	expected:= "Unsupported backend provider 'S3'"

	if err.Error() != expected {
		t.Errorf("Wrong output, expected %s, received %s", expected, err.Error())
	}
}


func TestBackendPromptWithNoTritonAccountNonInteractiveMode(t *testing.T) {
	viper.Set("non-interactive", true)
	viper.Set("backend_provider", "manta")

	defer viper.Reset()

	_,err:=PromptForBackend()

	expected:= "triton_account must be specified"

	if err.Error() != expected {
		t.Errorf("Wrong output, expected %s, received %s", expected, err.Error())
	}
}

func TestBackendPromptWithNoTritonSSHKeyPathNonInteractiveMode(t *testing.T) {
	viper.Set("non-interactive", true)
	viper.Set("backend_provider", "manta")
	viper.Set("triton_account", "xyz")

	defer viper.Reset()

	_,err:=PromptForBackend()

	expected:= "triton_key_path must be specified"

	if err.Error() != expected {
		t.Errorf("Wrong output, expected %s, received %s", expected, err.Error())
	}
}

func TestBackendPromptWithInvalidTritonSSHKeyPathInteractive(t *testing.T) {
	viper.Set("non-interactive", false)
	viper.Set("backend_provider", "manta")
	viper.Set("triton_account", "xyz")
	viper.Set("triton_key_path", "")

	defer viper.Reset()

	_,err:=PromptForBackend()

	expected:= "Unable to read private key: open : no such file or directory"

	if err.Error() != expected {
		t.Errorf("Wrong output, expected %s, received %s", expected, err.Error())
	}
}

func TestNoTritonURLForNonInteractiveMode (t *testing.T) {
	viper.Set("non-interactive", true)
	viper.Set("backend_provider", "manta")
	viper.Set("triton_account", "xyz")
	viper.Set("triton_key_path", "")
	viper.Set("triton_key_id", "")

	defer viper.Reset()

	_,err := PromptForBackend()

	expected := "triton_url must be specified"

	if err.Error() != expected {
		t.Errorf("Wrong output, expected %s, received %s", expected, err.Error())
	}
}

func TestNoMantaURLForNonInteractiveMode (t *testing.T) {
	viper.Set("non-interactive", true)
	viper.Set("backend_provider", "manta")
	viper.Set("triton_account", "xyz")
	viper.Set("triton_key_path", "")
	viper.Set("triton_key_id", "")
	viper.Set("triton_url", "xyz.triton.com")

	defer viper.Reset()

	_,err := PromptForBackend()

	expected := "manta_url must be specified"

	if err.Error() != expected {
		t.Errorf("Wrong output, expected %s, received %s", expected, err.Error())
	}
}