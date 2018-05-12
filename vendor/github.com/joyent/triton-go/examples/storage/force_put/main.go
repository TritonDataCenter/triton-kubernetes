package main

import (
	"encoding/pem"
	"io/ioutil"
	"log"
	"os"

	"context"

	"fmt"

	triton "github.com/joyent/triton-go"
	"github.com/joyent/triton-go/authentication"
	"github.com/joyent/triton-go/storage"
)

func main() {
	keyID := os.Getenv("TRITON_KEY_ID")
	keyMaterial := os.Getenv("TRITON_KEY_MATERIAL")
	mantaUser := os.Getenv("MANTA_USER")
	userName := os.Getenv("TRITON_USER")

	var signer authentication.Signer
	var err error

	if keyMaterial == "" {
		input := authentication.SSHAgentSignerInput{
			KeyID:       keyID,
			AccountName: mantaUser,
			Username:    userName,
		}
		signer, err = authentication.NewSSHAgentSigner(input)
		if err != nil {
			log.Fatalf("Error Creating SSH Agent Signer: %s", err.Error())
		}
	} else {
		var keyBytes []byte
		if _, err = os.Stat(keyMaterial); err == nil {
			keyBytes, err = ioutil.ReadFile(keyMaterial)
			if err != nil {
				log.Fatalf("Error reading key material from %s: %s",
					keyMaterial, err)
			}
			block, _ := pem.Decode(keyBytes)
			if block == nil {
				log.Fatalf(
					"Failed to read key material '%s': no key found", keyMaterial)
			}

			if block.Headers["Proc-Type"] == "4,ENCRYPTED" {
				log.Fatalf(
					"Failed to read key '%s': password protected keys are\n"+
						"not currently supported. Please decrypt the key prior to use.", keyMaterial)
			}

		} else {
			keyBytes = []byte(keyMaterial)
		}

		input := authentication.PrivateKeySignerInput{
			KeyID:              keyID,
			PrivateKeyMaterial: keyBytes,
			AccountName:        mantaUser,
			Username:           userName,
		}
		signer, err = authentication.NewPrivateKeySigner(input)
		if err != nil {
			log.Fatalf("Error Creating SSH Private Key Signer: %s", err.Error())
		}
	}

	config := &triton.ClientConfig{
		MantaURL:    os.Getenv("MANTA_URL"),
		AccountName: mantaUser,
		Username:    userName,
		Signers:     []authentication.Signer{signer},
	}

	client, err := storage.NewClient(config)
	if err != nil {
		log.Fatalf("NewClient: %s", err)
	}

	reader, err := os.Open("/tmp/foo.txt")
	if err != nil {
		log.Fatalf("os.Open: %s", err)
	}
	defer reader.Close()

	err = client.Objects().Put(context.Background(), &storage.PutObjectInput{
		ObjectPath:   "/stor/folder1/folder2/folder3/folder4/foo.txt",
		ObjectReader: reader,
		ForceInsert:  true,
	})

	if err != nil {
		log.Fatalf("Error creating nested folder structure: %s", err.Error())
	}
	fmt.Println("Successfully uploaded /tmp/foo.txt to /stor/folder1/folder2/folder3/folder4/foo.txt")
}
