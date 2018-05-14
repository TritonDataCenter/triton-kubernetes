package account

import (
	"context"
	"testing"

	"github.com/abdullin/seq"
	triton "github.com/joyent/triton-go"
	"github.com/joyent/triton-go/testutils"
)

func TestAccKey_Create(t *testing.T) {
	keyName := testutils.RandPrefixString("TestAccCreateKey", 32)

	testutils.AccTest(t, testutils.TestCase{
		Steps: []testutils.Step{

			&testutils.StepClient{
				StateBagKey: "key",
				CallFunc: func(config *triton.ClientConfig) (interface{}, error) {
					return NewClient(config)
				},
			},

			&testutils.StepAPICall{
				StateBagKey: "key",
				CallFunc: func(client interface{}) (interface{}, error) {
					c := client.(*AccountClient)
					ctx := context.Background()
					input := &CreateKeyInput{
						Name: keyName,
						Key:  testAccCreateKeyMaterial,
					}
					return c.Keys().Create(ctx, input)
				},
				CleanupFunc: func(client interface{}, callState interface{}) {
					c := client.(*AccountClient)
					ctx := context.Background()
					input := &DeleteKeyInput{
						KeyName: keyName,
					}
					c.Keys().Delete(ctx, input)
				},
			},
			&testutils.StepAssert{
				StateBagKey: "key",
				Assertions: seq.Map{
					"name":        keyName,
					"key":         testAccCreateKeyMaterial,
					"fingerprint": testAccCreateKeyFingerprint,
				},
			},
		},
	})
}

func TestAccKey_Get(t *testing.T) {
	keyName := testutils.RandPrefixString("TestAccGetKey", 32)

	testutils.AccTest(t, testutils.TestCase{
		Steps: []testutils.Step{

			&testutils.StepClient{
				StateBagKey: "key",
				CallFunc: func(config *triton.ClientConfig) (interface{}, error) {
					return NewClient(config)
				},
			},

			&testutils.StepAPICall{
				StateBagKey: "key",
				CallFunc: func(client interface{}) (interface{}, error) {
					c := client.(*AccountClient)
					ctx := context.Background()
					input := &CreateKeyInput{
						Name: keyName,
						Key:  testAccCreateKeyMaterial,
					}
					return c.Keys().Create(ctx, input)
				},
				CleanupFunc: func(client interface{}, callState interface{}) {
					c := client.(*AccountClient)
					ctx := context.Background()
					input := &DeleteKeyInput{
						KeyName: keyName,
					}
					c.Keys().Delete(ctx, input)
				},
			},

			&testutils.StepAPICall{
				StateBagKey: "getKey",
				CallFunc: func(client interface{}) (interface{}, error) {
					c := client.(*AccountClient)
					ctx := context.Background()
					input := &GetKeyInput{
						KeyName: keyName,
					}
					return c.Keys().Get(ctx, input)
				},
			},

			&testutils.StepAssert{
				StateBagKey: "getKey",
				Assertions: seq.Map{
					"name":        keyName,
					"key":         testAccCreateKeyMaterial,
					"fingerprint": testAccCreateKeyFingerprint,
				},
			},
		},
	})
}

func TestAccKey_Delete(t *testing.T) {
	keyName := testutils.RandPrefixString("TestAccGetKey", 32)

	testutils.AccTest(t, testutils.TestCase{
		Steps: []testutils.Step{

			&testutils.StepClient{
				StateBagKey: "key",
				CallFunc: func(config *triton.ClientConfig) (interface{}, error) {
					return NewClient(config)
				},
			},

			&testutils.StepAPICall{
				StateBagKey: "key",
				CallFunc: func(client interface{}) (interface{}, error) {
					c := client.(*AccountClient)
					ctx := context.Background()
					input := &CreateKeyInput{
						Name: keyName,
						Key:  testAccCreateKeyMaterial,
					}
					return c.Keys().Create(ctx, input)
				},
				CleanupFunc: func(client interface{}, callState interface{}) {
					c := client.(*AccountClient)
					ctx := context.Background()
					input := &DeleteKeyInput{
						KeyName: keyName,
					}
					c.Keys().Delete(ctx, input)
				},
			},

			&testutils.StepAPICall{
				StateBagKey: "noop",
				CallFunc: func(client interface{}) (interface{}, error) {
					c := client.(*AccountClient)
					ctx := context.Background()
					input := &DeleteKeyInput{
						KeyName: keyName,
					}
					return nil, c.Keys().Delete(ctx, input)
				},
			},

			&testutils.StepAPICall{
				ErrorKey: "getKeyError",
				CallFunc: func(client interface{}) (interface{}, error) {
					c := client.(*AccountClient)
					ctx := context.Background()
					input := &GetKeyInput{
						KeyName: keyName,
					}
					return c.Keys().Get(ctx, input)
				},
			},

			&testutils.StepAssertTritonError{
				ErrorKey: "getKeyError",
				Code:     "ResourceNotFound",
			},
		},
	})
}

const testAccCreateKeyMaterial = `ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDBOJ5z6jTdY3SYK2Nc+MQLSQstAOzxFqDN00MJ9SMhJea8ZQbZFlhCAZBFE4TUBDI3zXBxFjygh84lb1QlNu1dmZeoQ10MThuowZllBAfg9Eb5RkXqLvDdYh9+rLdEdUL4+aiYZ8JYtQ+K5ZnogZoxdzNQ3WnVhMGJIrj1zcRveUSvQ6tMhaEQDxDWrAMDLxnLI/6SNmkhdF1ZKE8iQ+BnazYp0vg5jAzkHzEYJY9kFUOubupOxio93B9OTkpQ0jZD+J9iR1t8Me3JdhHy85inaAFc0fkjznDYluV8aqfIprD/WE9grQ/GfEYfsvQdQr1ljLBJZdad7DvnKqU0M4vJ James@jn-mpb15`
const testAccCreateKeyFingerprint = `ab:f4:8f:bc:26:e1:cf:1d:06:a3:9d:40:39:7c:5a:78`
