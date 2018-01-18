package network

import (
	"context"
	"fmt"
	"testing"

	triton "github.com/joyent/triton-go"
	"github.com/joyent/triton-go/testutils"
)

// Note that this is specific to Joyent Public Cloud and will not pass on
// private installations of Triton.
func TestAccNetworks_List(t *testing.T) {
	testutils.AccTest(t, testutils.TestCase{
		Steps: []testutils.Step{

			&testutils.StepClient{
				StateBagKey: "datacenter",
				CallFunc: func(config *triton.ClientConfig) (interface{}, error) {
					return NewClient(config)
				},
			},

			&testutils.StepAPICall{
				StateBagKey: "networks",
				CallFunc: func(client interface{}) (interface{}, error) {
					ctx := context.Background()
					input := &ListInput{}
					if c, ok := client.(*NetworkClient); ok {
						return c.List(ctx, input)
					}
					return nil, fmt.Errorf("Bad client initialization")
				},
			},

			&testutils.StepAssertFunc{
				AssertFunc: func(state testutils.TritonStateBag) error {
					dcs, ok := state.GetOk("networks")
					if !ok {
						return fmt.Errorf("State key %q not found", "networks")
					}

					toFind := []string{"Joyent-SDC-Private", "Joyent-SDC-Public"}
					for _, dcName := range toFind {
						found := false
						for _, dc := range dcs.([]*Network) {
							if dc.Name == dcName {
								found = true
								if dc.Id == "" {
									return fmt.Errorf("%q has no ID", dc.Name)
								}
							}
						}
						if !found {
							return fmt.Errorf("Did not find Network %q", dcName)
						}
					}

					return nil
				},
			},
		},
	})
}
