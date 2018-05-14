package main

import (
	"context"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	triton "github.com/joyent/triton-go"
	"github.com/joyent/triton-go/authentication"
	"github.com/joyent/triton-go/network"
)

func main() {
	keyID := os.Getenv("TRITON_KEY_ID")
	accountName := os.Getenv("TRITON_ACCOUNT")
	keyMaterial := os.Getenv("TRITON_KEY_MATERIAL")
	userName := os.Getenv("TRITON_USER")

	var signer authentication.Signer
	var err error

	if keyMaterial == "" {
		input := authentication.SSHAgentSignerInput{
			KeyID:       keyID,
			AccountName: accountName,
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
			AccountName:        accountName,
			Username:           userName,
		}
		signer, err = authentication.NewPrivateKeySigner(input)
		if err != nil {
			log.Fatalf("Error Creating SSH Private Key Signer: %s", err.Error())
		}
	}

	config := &triton.ClientConfig{
		TritonURL:   os.Getenv("TRITON_URL"),
		AccountName: accountName,
		Username:    userName,
		Signers:     []authentication.Signer{signer},
	}

	n, err := network.NewClient(config)
	if err != nil {
		log.Fatalf("Network NewClient(): %s", err)
	}

	fabric, err := n.Fabrics().Create(context.Background(), &network.CreateFabricInput{
		FabricVLANID:     2,
		Name:             "testnet",
		Description:      "This is a test network",
		Subnet:           "10.50.1.0/24",
		ProvisionStartIP: "10.50.1.10",
		ProvisionEndIP:   "10.50.1.240",
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("Fabric was successfully created!")
	fmt.Println("Name:", fabric.Name)
	time.Sleep(5 * time.Second)

	err = n.Fabrics().Delete(context.Background(), &network.DeleteFabricInput{
		FabricVLANID: 2,
		NetworkID:    fabric.Id,
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("Fabric was successfully deleted!")
	time.Sleep(5 * time.Second)

	fwrule, err := n.Firewall().CreateRule(context.Background(), &network.CreateRuleInput{
		Enabled: false,
		Rule:    "FROM any TO tag \"bone-thug\" = \"basket-ball\" ALLOW udp PORT 8600",
	})

	fmt.Println("Firewall Rule was successfully added!")
	time.Sleep(5 * time.Second)

	err = n.Firewall().DeleteRule(context.Background(), &network.DeleteRuleInput{
		ID: fwrule.ID,
	})

	fmt.Println("Firewall Rule was successfully deleted!")
	time.Sleep(5 * time.Second)

}
