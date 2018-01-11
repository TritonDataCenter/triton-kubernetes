package shell

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

func GetPublicKeyFingerprintFromPrivateKey(privateKeyPath string) (string, error) {
	// ssh-keygen -E md5 -lf PATH_TO_FILE
	// Sample output:
	// 2048 MD5:68:9f:9a:c4:76:3a:f4:62:77:47:3e:47:d4:34:4a:b7 njalali@Nimas-MacBook-Pro.local (RSA)
	out, err := exec.Command("ssh-keygen", "-E", "md5", "-lf", privateKeyPath).Output()
	if err != nil {
		return "", fmt.Errorf("Failed to exec ssh-keygen: %s", err)
	}

	parts := strings.Split(string(out), " ")
	if len(parts) != 4 {
		return "", errors.New("Could not get ssh key id")
	}

	return strings.TrimPrefix(parts[1], "MD5:"), nil
}
