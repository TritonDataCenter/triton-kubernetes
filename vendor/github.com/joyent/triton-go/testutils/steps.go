package testutils

import (
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"

	"github.com/abdullin/seq"
	"github.com/hashicorp/errwrap"
	triton "github.com/joyent/triton-go"
	"github.com/joyent/triton-go/client"
)

type StepClient struct {
	StateBagKey string
	ErrorKey    string
	CallFunc    func(config *triton.ClientConfig) (interface{}, error)
	CleanupFunc func(client interface{}, callState interface{})
}

func (s *StepClient) Run(state TritonStateBag) StepAction {
	client, err := s.CallFunc(state.Config())
	if err != nil {
		if s.ErrorKey == "" {
			state.AppendError(err)
			return Halt
		}

		state.Put(s.ErrorKey, err)
		return Continue
	}

	state.PutClient(client)
	return Continue
}

func (s *StepClient) Cleanup(state TritonStateBag) {
	return
}

type StepAPICall struct {
	StateBagKey string
	ErrorKey    string
	CallFunc    func(client interface{}) (interface{}, error)
	CleanupFunc func(client interface{}, callState interface{})
}

func (s *StepAPICall) Run(state TritonStateBag) StepAction {
	result, err := s.CallFunc(state.Client())
	if err != nil {
		if s.ErrorKey == "" {
			state.AppendError(err)
			return Halt
		}

		state.Put(s.ErrorKey, err)
		return Continue
	}

	state.Put(s.StateBagKey, result)
	return Continue
}

func (s *StepAPICall) Cleanup(state TritonStateBag) {
	if s.CleanupFunc == nil {
		return
	}

	if callState, ok := state.GetOk(s.StateBagKey); ok {
		s.CleanupFunc(state.Client(), callState)
	} else {
		log.Print("[INFO] No state for API call, calling cleanup with nil call state")
		s.CleanupFunc(state.Client(), nil)
	}
}

type AssertFunc func(TritonStateBag) error

type StepAssertFunc struct {
	AssertFunc AssertFunc
}

func (s *StepAssertFunc) Run(state TritonStateBag) StepAction {
	if s.AssertFunc == nil {
		state.AppendError(errors.New("StepAssertFunc may not have a nil AssertFunc"))
		return Halt
	}

	err := s.AssertFunc(state)
	if err != nil {
		state.AppendError(err)
		return Halt
	}

	return Continue
}

func (s *StepAssertFunc) Cleanup(state TritonStateBag) {
	return
}

type StepAssert struct {
	StateBagKey string
	Assertions  seq.Map
}

func (s *StepAssert) Run(state TritonStateBag) StepAction {
	actual, ok := state.GetOk(s.StateBagKey)
	if !ok {
		state.AppendError(fmt.Errorf("Key %q not found in state", s.StateBagKey))
	}

	for k, v := range s.Assertions {
		path := fmt.Sprintf("%s.%s", s.StateBagKey, k)
		if os.Getenv("TRITON_VERBOSE_TESTS") != "" {
			log.Printf("[INFO] Asserting %q has value \"%v\"", path, v)
		} else {
			vPrefix := fmt.Sprintf("%v", v)
			if len(vPrefix) > 15 {
				vPrefix = fmt.Sprintf("%s...", vPrefix[:15])
			}
			log.Printf("[INFO] Asserting %q has value \"%s\"", path, vPrefix)
		}
	}

	result := s.Assertions.Test(actual)

	if result.Ok() {
		return Continue
	}

	for _, v := range result.Issues {
		err := fmt.Sprintf("Expected %q to be \"%v\" but got %q",
			v.Path,
			v.ExpectedValue,
			v.ActualValue,
		)
		state.AppendError(fmt.Errorf(err))
	}

	return Halt
}

func (s *StepAssert) Cleanup(state TritonStateBag) {
	return
}

type StepAssertSet struct {
	StateBagKey string
	Keys        []string
}

func (s *StepAssertSet) Run(state TritonStateBag) StepAction {
	actual, ok := state.GetOk(s.StateBagKey)
	if !ok {
		state.AppendError(fmt.Errorf("Key %q not found in state", s.StateBagKey))
	}

	var pass = true
	for _, key := range s.Keys {
		r := reflect.ValueOf(actual)
		f := reflect.Indirect(r).FieldByName(key)

		log.Printf("[INFO] Asserting %q has a non-zero value", key)
		if f.Interface() == reflect.Zero(reflect.TypeOf(f)).Interface() {
			err := fmt.Sprintf("Expected %q to have a non-zero value", key)
			state.AppendError(fmt.Errorf(err))
			pass = false
		}
	}

	if !pass {
		return Halt
	}

	return Continue
}

func (s *StepAssertSet) Cleanup(state TritonStateBag) {
	return
}

type StepAssertTritonError struct {
	ErrorKey string
	Code     string
}

func (s *StepAssertTritonError) Run(state TritonStateBag) StepAction {
	err, ok := state.GetOk(s.ErrorKey)
	if !ok {
		state.AppendError(fmt.Errorf("Expected TritonError %q to be in state", s.Code))
		return Halt
	}

	tritonErrorInterface := errwrap.GetType(err.(error), &client.TritonError{})
	if tritonErrorInterface == nil {
		state.AppendError(errors.New("Expected a TritonError in wrapped error chain"))
		return Halt
	}

	tritonErr := tritonErrorInterface.(*client.TritonError)
	if tritonErr.Code == s.Code {
		return Continue
	}

	state.AppendError(fmt.Errorf("Expected TritonError code %q to be in state key %q, was %q", s.Code, s.ErrorKey, tritonErr.Code))
	return Halt
}

func (s *StepAssertTritonError) Cleanup(state TritonStateBag) {
	return
}
