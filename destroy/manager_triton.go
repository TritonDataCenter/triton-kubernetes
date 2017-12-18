package destroy

import (
	"os"
	"os/exec"
)

func DeleteTritonManager() error {
	// Run terraform apply
	tfDestroy := exec.Command("terraform", []string{"destroy", "-force"}...)
	tfDestroy.Stdin = os.Stdin
	tfDestroy.Stdout = os.Stdout
	tfDestroy.Stderr = os.Stderr

	if err := tfDestroy.Start(); err != nil {
		return err
	}

	err := tfDestroy.Wait()
	if err != nil {
		return err
	}

	// TODO: Delete terraform backend tfstate path

	return nil
}
