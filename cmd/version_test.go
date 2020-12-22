package cmd

import (
	"regexp"
	"testing"

	"github.com/joyent/triton-kubernetes/test_pkg"
)

func TestVersion(t *testing.T) {
	tCase := test_pkg.NewT(t)
	cliVersion = "beta"

	outch, errch := test_pkg.AlterStdout(func() {
		versionCmd.Run(versionCmd, []string{})
	})

	expected := "triton-kubernetes 1.0.1-pre1 (beta)\n"

	select {
	case err := <-errch:
		tCase.Fatal("altering output", nil, err)
	case actual := <-outch:
		if expected != string(actual) {
			tCase.Fatal("output", expected, string(actual))
		}
	}
}

func TestMissingVersion(t *testing.T) {
	tCase := test_pkg.NewT(t)

	cliVersion = ""

	outch, errch := test_pkg.AlterStdout(func() {
		versionCmd.Run(versionCmd, []string{})
	})

	match := "no version set for this build"

	select {
	case err := <-errch:
		tCase.Fatal("altering output", nil, err)
	case output := <-outch:
		if match, err := regexp.Match(match, output); !match || err != nil {
			tCase.Fatal("output contents", match, string(output))
		}
	}
}
