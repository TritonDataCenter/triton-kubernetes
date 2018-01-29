package compute

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"reflect"
	"testing"
	"time"

	triton "github.com/joyent/triton-go"
	"github.com/joyent/triton-go/network"
	"github.com/joyent/triton-go/testutils"
)

func getAnyInstanceID(t *testing.T, client *ComputeClient) (string, error) {
	ctx := context.Background()
	input := &ListInstancesInput{}
	instances, err := client.Instances().List(ctx, input)
	if err != nil {
		return "", err
	}

	for _, m := range instances {
		if len(m.ID) > 0 {
			return m.ID, nil
		}
	}

	t.Skip()
	return "", errors.New("no machines configured")
}

func RandInt() int {
	reseed()
	return rand.New(rand.NewSource(time.Now().UnixNano())).Int()
}

func RandWithPrefix(name string) string {
	return fmt.Sprintf("%s-%d", name, RandInt())
}

// Seeds random with current timestamp
func reseed() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func TestAccInstances_Create(t *testing.T) {
	testInstanceName := RandWithPrefix("acctest")

	testutils.AccTest(t, testutils.TestCase{
		Steps: []testutils.Step{

			&testutils.StepClient{
				StateBagKey: "instances",
				CallFunc: func(config *triton.ClientConfig) (interface{}, error) {
					computeClient, err := NewClient(config)
					if err != nil {
						return nil, err
					}

					networkClient, err := network.NewClient(config)
					if err != nil {
						return nil, err
					}

					return []interface{}{
						computeClient,
						networkClient,
					}, nil
				},
			},

			&testutils.StepAPICall{
				StateBagKey: "instances",
				CallFunc: func(client interface{}) (interface{}, error) {
					clients := client.([]interface{})
					c := clients[0].(*ComputeClient)
					n := clients[1].(*network.NetworkClient)

					images, err := c.Images().List(context.Background(), &ListImagesInput{
						Name:    "ubuntu-16.04",
						Version: "20170403",
					})
					img := images[0]

					var net *network.Network
					networkName := "Joyent-SDC-Private"
					nets, err := n.List(context.Background(), &network.ListInput{})
					if err != nil {
						return nil, err
					}
					for _, found := range nets {
						if found.Name == networkName {
							net = found
						}
					}

					input := &CreateInstanceInput{
						Name:     testInstanceName,
						Package:  "g4-highcpu-128M",
						Image:    img.ID,
						Networks: []string{net.Id},
						Metadata: map[string]string{
							"metadata1": "value1",
						},
						Tags: map[string]string{
							"tag1": "value1",
						},
						CNS: InstanceCNS{
							Services: []string{"testapp", "testweb"},
						},
					}
					created, err := c.Instances().Create(context.Background(), input)
					if err != nil {
						return nil, err
					}

					state := make(chan *Instance, 1)
					go func(createdID string, c *ComputeClient) {
						for {
							time.Sleep(1 * time.Second)
							instance, err := c.Instances().Get(context.Background(), &GetInstanceInput{
								ID: createdID,
							})
							if err != nil {
								log.Fatalf("Get(): %v", err)
							}
							if instance.State == "running" {
								state <- instance
							}
						}
					}(created.ID, c)

					select {
					case instance := <-state:
						return instance, nil
					case <-time.After(5 * time.Minute):
						return nil, fmt.Errorf("Timed out waiting for instance to provision")
					}
				},
				CleanupFunc: func(client interface{}, stateBag interface{}) {
					instance, instOk := stateBag.(*Instance)
					if !instOk {
						log.Println("Expected instance to be Instance")
						return
					}

					if instance.Name != testInstanceName {
						log.Printf("Expected instance to be named %s: found %s\n",
							testInstanceName, instance.Name)
						return
					}

					clients := client.([]interface{})
					c, clientOk := clients[0].(*ComputeClient)
					if !clientOk {
						log.Println("Expected client to be ComputeClient")
						return
					}

					err := c.Instances().Delete(context.Background(), &DeleteInstanceInput{
						ID: instance.ID,
					})
					if err != nil {
						log.Printf("Could not delete instance %s\n", instance.Name)
					}
					return
				},
			},

			&testutils.StepAssertFunc{
				AssertFunc: func(state testutils.TritonStateBag) error {
					instanceRaw, found := state.GetOk("instances")
					if !found {
						return fmt.Errorf("State key %q not found", "instances")
					}
					instance, ok := instanceRaw.(*Instance)
					if !ok {
						return errors.New("Expected state to include instance")
					}

					if instance.State != "running" {
						return fmt.Errorf("Expected instance state to be \"running\": found %s",
							instance.State)
					}
					if instance.ID == "" {
						return fmt.Errorf("Expected instance ID: found \"\"")
					}
					if instance.Name == "" {
						return fmt.Errorf("Expected instance Name: found \"\"")
					}
					if instance.Memory != 128 {
						return fmt.Errorf("Expected instance Memory to be 128: found \"%d\"",
							instance.Memory)
					}

					metadataVal, metaOk := instance.Metadata["metadata1"]
					if !metaOk {
						return fmt.Errorf("Expected instance to have Metadata: found \"%v\"",
							instance.Metadata)
					}
					if metadataVal != "value1" {
						return fmt.Errorf("Expected instance Metadata \"metadata1\" to equal \"value1\": found \"%s\"",
							metadataVal)
					}

					tagVal, tagOk := instance.Tags["tag1"]
					if !tagOk {
						return fmt.Errorf("Expected instance to have Tags: found \"%v\"",
							instance.Tags)
					}
					if tagVal != "value1" {
						return fmt.Errorf("Expected instance Tag \"tag1\" to equal \"value1\": found \"%s\"",
							tagVal)
					}

					services := []string{"testapp", "testweb"}
					if !reflect.DeepEqual(instance.CNS.Services, services) {
						return fmt.Errorf("Expected instance CNS Services \"%s\", to equal \"%v\"",
							instance.CNS.Services, services)
					}
					return nil
				},
			},
		},
	})
}

func TestAccInstances_Get(t *testing.T) {
	testutils.AccTest(t, testutils.TestCase{
		Steps: []testutils.Step{

			&testutils.StepClient{
				StateBagKey: "instances",
				CallFunc: func(config *triton.ClientConfig) (interface{}, error) {
					return NewClient(config)
				},
			},

			&testutils.StepAPICall{
				StateBagKey: "instances",
				CallFunc: func(client interface{}) (interface{}, error) {
					c := client.(*ComputeClient)

					instanceID, err := getAnyInstanceID(t, c)
					if err != nil {
						return nil, err
					}

					ctx := context.Background()
					input := &GetInstanceInput{
						ID: instanceID,
					}
					return c.Instances().Get(ctx, input)
				},
			},

			&testutils.StepAssertSet{
				StateBagKey: "instances",
				Keys:        []string{"ID", "Name", "Type", "Tags"},
			},
		},
	})
}

// FIXME(seanc@): TestAccMachine_ListMachineTags assumes that any machine ID
// returned from getAnyInstanceID will have at least one tag.
func TestAccInstances_ListTags(t *testing.T) {
	testutils.AccTest(t, testutils.TestCase{
		Steps: []testutils.Step{

			&testutils.StepClient{
				StateBagKey: "instances",
				CallFunc: func(config *triton.ClientConfig) (interface{}, error) {
					return NewClient(config)
				},
			},

			&testutils.StepAPICall{
				StateBagKey: "instances",
				CallFunc: func(client interface{}) (interface{}, error) {
					c := client.(*ComputeClient)

					instanceID, err := getAnyInstanceID(t, c)
					if err != nil {
						return nil, err
					}

					ctx := context.Background()
					input := &ListTagsInput{
						ID: instanceID,
					}
					return c.Instances().ListTags(ctx, input)
				},
			},

			&testutils.StepAssertFunc{
				AssertFunc: func(state testutils.TritonStateBag) error {
					tagsRaw, found := state.GetOk("instances")
					if !found {
						return fmt.Errorf("State key %q not found", "instances")
					}

					tags := tagsRaw.(map[string]interface{})
					if len(tags) == 0 {
						return errors.New("Expected at least one tag on machine")
					}
					return nil
				},
			},
		},
	})
}

func TestAccInstances_UpdateMetadata(t *testing.T) {
	testutils.AccTest(t, testutils.TestCase{
		Steps: []testutils.Step{

			&testutils.StepClient{
				StateBagKey: "instances",
				CallFunc: func(config *triton.ClientConfig) (interface{}, error) {
					return NewClient(config)
				},
			},

			&testutils.StepAPICall{
				StateBagKey: "instances",
				CallFunc: func(client interface{}) (interface{}, error) {
					c := client.(*ComputeClient)

					instanceID, err := getAnyInstanceID(t, c)
					if err != nil {
						return nil, err
					}

					ctx := context.Background()
					input := &UpdateMetadataInput{
						ID: instanceID,
						Metadata: map[string]string{
							"tester": os.Getenv("USER"),
						},
					}
					return c.Instances().UpdateMetadata(ctx, input)
				},
			},

			&testutils.StepAssertFunc{
				AssertFunc: func(state testutils.TritonStateBag) error {
					mdataRaw, found := state.GetOk("instances")
					if !found {
						return fmt.Errorf("State key %q not found", "instances")
					}

					mdata := mdataRaw.(map[string]string)
					if len(mdata) == 0 {
						return errors.New("Expected metadata on machine")
					}

					if mdata["tester"] != os.Getenv("USER") {
						return errors.New("Expected test metadata to equal environ $USER")
					}
					return nil
				},
			},
		},
	})
}

func TestAccInstances_ListMetadata(t *testing.T) {
	testutils.AccTest(t, testutils.TestCase{
		Steps: []testutils.Step{

			&testutils.StepClient{
				StateBagKey: "instances",
				CallFunc: func(config *triton.ClientConfig) (interface{}, error) {
					return NewClient(config)
				},
			},

			&testutils.StepAPICall{
				StateBagKey: "instances",
				CallFunc: func(client interface{}) (interface{}, error) {
					c := client.(*ComputeClient)

					instanceID, err := getAnyInstanceID(t, c)
					if err != nil {
						return nil, err
					}

					ctx := context.Background()
					input := &ListMetadataInput{
						ID: instanceID,
					}
					return c.Instances().ListMetadata(ctx, input)
				},
			},

			&testutils.StepAssertFunc{
				AssertFunc: func(state testutils.TritonStateBag) error {
					mdataRaw, found := state.GetOk("instances")
					if !found {
						return fmt.Errorf("State key %q not found", "instances")
					}

					mdata := mdataRaw.(map[string]string)
					if len(mdata) == 0 {
						return errors.New("Expected metadata on machine")
					}

					if mdata["root_authorized_keys"] == "" {
						return errors.New("Expected test metadata to have key")
					}
					return nil
				},
			},
		},
	})
}

func TestAccInstances_GetMetadata(t *testing.T) {
	testutils.AccTest(t, testutils.TestCase{
		Steps: []testutils.Step{

			&testutils.StepClient{
				StateBagKey: "instances",
				CallFunc: func(config *triton.ClientConfig) (interface{}, error) {
					return NewClient(config)
				},
			},

			&testutils.StepAPICall{
				StateBagKey: "instances",
				CallFunc: func(client interface{}) (interface{}, error) {
					c := client.(*ComputeClient)

					instanceID, err := getAnyInstanceID(t, c)
					if err != nil {
						return nil, err
					}

					ctx := context.Background()
					input := &UpdateMetadataInput{
						ID: instanceID,
						Metadata: map[string]string{
							"testkey": os.Getenv("USER"),
						},
					}
					_, err = c.Instances().UpdateMetadata(ctx, input)
					if err != nil {
						return nil, err
					}

					ctx2 := context.Background()
					input2 := &GetMetadataInput{
						ID:  instanceID,
						Key: "testkey",
					}
					return c.Instances().GetMetadata(ctx2, input2)
				},
			},

			&testutils.StepAssertFunc{
				AssertFunc: func(state testutils.TritonStateBag) error {
					mdataValue := state.Get("instances")
					retValue := fmt.Sprintf("\"%s\"", os.Getenv("USER"))
					if mdataValue != retValue {
						return errors.New("Expected test metadata to equal environ \"$USER\"")
					}
					return nil
				},
			},
		},
	})
}
