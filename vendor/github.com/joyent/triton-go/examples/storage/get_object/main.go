package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"encoding/pem"

	triton "github.com/joyent/triton-go"
	"github.com/joyent/triton-go/authentication"
	"github.com/joyent/triton-go/storage"
)

// This file stored in Manta is used in the example below.
const path = "/stor/books/dracula.txt"

func main() {
	var (
		signer authentication.Signer
		err    error

		keyID       = os.Getenv("MANTA_KEY_ID")
		accountName = os.Getenv("MANTA_USER")
		keyMaterial = os.Getenv("MANTA_KEY_MATERIAL")
		userName    = os.Getenv("TRITON_USER")
	)

	if keyMaterial == "" {
		input := authentication.SSHAgentSignerInput{
			KeyID:       keyID,
			AccountName: accountName,
			Username:    userName,
		}
		signer, err = authentication.NewSSHAgentSigner(input)
		if err != nil {
			log.Fatalf("error creating SSH agent signer: %v", err.Error())
		}
	} else {
		var keyBytes []byte
		if _, err = os.Stat(keyMaterial); err == nil {
			keyBytes, err = ioutil.ReadFile(keyMaterial)
			if err != nil {
				log.Fatalf("error reading key material from %q: %v",
					keyMaterial, err)
			}
			block, _ := pem.Decode(keyBytes)
			if block == nil {
				log.Fatalf(
					"failed to read key material %q: no key found", keyMaterial)
			}

			if block.Headers["Proc-Type"] == "4,ENCRYPTED" {
				log.Fatalf("failed to read key %q: password protected keys are\n"+
					"not currently supported, decrypt key prior to use",
					keyMaterial)
			}

		} else {
			keyBytes = []byte(keyMaterial)
		}

		input := authentication.PrivateKeySignerInput{
			KeyID:              keyID,
			PrivateKeyMaterial: keyBytes,
			AccountName:        accountName,
			Username:           userName,
		}
		signer, err = authentication.NewPrivateKeySigner(input)
		if err != nil {
			log.Fatalf("error creating SSH private key signer: %v", err.Error())
		}
	}

	config := &triton.ClientConfig{
		MantaURL:    os.Getenv("MANTA_URL"),
		AccountName: accountName,
		Username:    userName,
		Signers:     []authentication.Signer{signer},
	}

	client, err := storage.NewClient(config)
	if err != nil {
		log.Fatalf("failed to init storage client: %v", err)
	}

	ctx := context.Background()
	info, err := client.Objects().GetInfo(ctx, &storage.GetInfoInput{
		ObjectPath: path,
	})
	if err != nil {
		fmt.Printf("could not find %q\n", path)
		return
	}

	fmt.Println("--- HEAD ---")
	fmt.Printf("Content-Length: %d\n", info.ContentLength)
	fmt.Printf("Content-MD5: %s\n", info.ContentMD5)
	fmt.Printf("Content-Type: %s\n", info.ContentType)
	fmt.Printf("ETag: %s\n", info.ETag)
	fmt.Printf("Date-Modified: %s\n", info.LastModified.String())

	ctx = context.Background()
	isDir, err := client.Objects().IsDir(ctx, path)
	if err != nil {
		log.Fatalf("failed to detect directory %q: %v\n", path, err)
		return
	}

	if isDir {
		fmt.Printf("%q is a directory\n", path)
	} else {
		fmt.Printf("%q is a file\n", path)
	}

	ctx = context.Background()
	obj, err := client.Objects().Get(ctx, &storage.GetObjectInput{
		ObjectPath: path,
	})
	if err != nil {
		log.Fatalf("failed to get %q: %v", path, err)
	}

	body, err := ioutil.ReadAll(obj.ObjectReader)
	if err != nil {
		log.Fatalf("failed to read response body: %v", err)
	}
	defer obj.ObjectReader.Close()

	fmt.Println("--- GET ---")
	fmt.Printf("Content-Length: %d\n", obj.ContentLength)
	fmt.Printf("Content-MD5: %s\n", obj.ContentMD5)
	fmt.Printf("Content-Type: %s\n", obj.ContentType)
	fmt.Printf("ETag: %s\n", obj.ETag)
	fmt.Printf("Date-Modified: %s\n", obj.LastModified.String())
	fmt.Printf("Length: %d\n", len(body))
}
