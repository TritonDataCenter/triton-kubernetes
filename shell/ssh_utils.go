package shell

import (
	"errors"
	"os/exec"
	"strings"
)

func GetPublicKeyFingerprintFromPrivateKey(privateKeyPath string) (string, error) {
	out, err := exec.Command("ssh-keygen", "-E", "md5", "-lf", privateKeyPath).Output()
	if err != nil {
		return "", err
	}

	parts := strings.Split(string(out), " ")
	if len(parts) != 4 {
		return "", errors.New("Could not get ssh key id")
	}

	return strings.TrimPrefix(parts[1], "MD5:"), nil
}
