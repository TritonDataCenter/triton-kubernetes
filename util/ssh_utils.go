package util

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"

	"github.com/manifoldco/promptui"
	"golang.org/x/crypto/ssh"
)

// GetPublicKeyFingerprintFromPrivateKey takes in location of a private key and returns the md5 fingerprint
func GetPublicKeyFingerprintFromPrivateKey(privateKeyPath string) (string, error) {
	var fingerprint string
	var err error

	key, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return "", fmt.Errorf("Unable to read private key: %v", err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		prompt := promptui.Prompt{
			Label: "Private Key Password",
			Mask:  '*',
		}
		password, _ := prompt.Run()
		signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(password))
		if err != nil {
			return "", fmt.Errorf("Unable to parse private key: %v", err)
		}
	}
	h := md5.New()
	h.Write(signer.PublicKey().Marshal())
	for i, b := range h.Sum(nil) {
		fingerprint += fmt.Sprintf("%02x", b)
		if i < len(h.Sum(nil))-1 {
			fingerprint += fmt.Sprintf(":")
		}
	}
	return fingerprint, err
}
