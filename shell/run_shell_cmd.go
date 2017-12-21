package shell

import (
	"os"
	"os/exec"
)

func RunShellCommand(options *ShellOptions, command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if (options != nil) {
		cmd.Dir = options.WorkingDir
	}

	err := cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}
