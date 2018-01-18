package compute

import (
	"context"
	"fmt"
	"testing"

	"github.com/abdullin/seq"
	triton "github.com/joyent/triton-go"
	"github.com/joyent/triton-go/testutils"
)

// Note that this is specific to Joyent Public Cloud and will not pass on
// private installations of Triton.
func TestAccDataCenters_Get(t *testing.T) {
	const dataCenterName = "us-east-1"
	const dataCenterURL = "https://us-east-1.api.joyentcloud.com"

	testutils.AccTest(t, testutils.TestCase{
		Steps: []testutils.Step{

			&testutils.StepClient{
				StateBagKey: "datacenter",
				CallFunc: func(config *triton.ClientConfig) (interface{}, error) {
					return NewClient(config)
				},
			},

			&testutils.StepAPICall{
				StateBagKey: "datacenter",
				CallFunc: func(client interface{}) (interface{}, error) {
					c := client.(*ComputeClient)
					ctx := context.Background()
					input := &GetDataCenterInput{
						Name: dataCenterName,
					}
					return c.Datacenters().Get(ctx, input)
				},
			},

			&testutils.StepAssert{
				StateBagKey: "datacenter",
				Assertions: seq.Map{
					"name": dataCenterName,
					"url":  dataCenterURL,
				},
			},
		},
	})
}

// Note that this is specific to Joyent Public Cloud and will not pass on
// private installations of Triton.
func TestAccDataCenters_List(t *testing.T) {
	testutils.AccTest(t, testutils.TestCase{
		Steps: []testutils.Step{

			&testutils.StepClient{
				StateBagKey: "datacenter",
				CallFunc: func(config *triton.ClientConfig) (interface{}, error) {
					return NewClient(config)
				},
			},

			&testutils.StepAPICall{
				StateBagKey: "datacenters",
				CallFunc: func(client interface{}) (interface{}, error) {
					c := client.(*ComputeClient)
					ctx := context.Background()
					input := &ListDataCentersInput{}
					return c.Datacenters().List(ctx, input)
				},
			},

			&testutils.StepAssertFunc{
				AssertFunc: func(state testutils.TritonStateBag) error {
					dcs, ok := state.GetOk("datacenters")
					if !ok {
						return fmt.Errorf("State key %q not found", "datacenters")
					}

					toFind := []string{"us-east-1", "eu-ams-1"}
					for _, dcName := range toFind {
						found := false
						for _, dc := range dcs.([]*DataCenter) {
							if dc.Name == dcName {
								found = true
								if dc.URL == "" {
									return fmt.Errorf("%q has no URL", dc.Name)
								}
							}
						}
						if !found {
							return fmt.Errorf("Did not find DC %q", dcName)
						}
					}

					return nil
				},
			},
		},
	})
}
