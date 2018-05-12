package testutils

import (
	"sync"

	triton "github.com/joyent/triton-go"
)

type TritonStateBag interface {
	Get(string) interface{}
	GetOk(string) (interface{}, bool)
	Put(string, interface{})
	Remove(string)

	Config() *triton.ClientConfig

	Client() interface{}
	PutClient(interface{})

	AppendError(error)
	ErrorsOrNil() []error
}

// TritonStateBag implements StateBag by using a normal map underneath
// protected by a RWMutex.
type basicTritonStateBag struct {
	TritonConfig *triton.ClientConfig
	TritonClient interface{}

	errors []error
	data   map[string]interface{}

	l    sync.RWMutex
	once sync.Once
}

func (b *basicTritonStateBag) Config() *triton.ClientConfig {
	b.l.RLock()
	defer b.l.RUnlock()

	return b.TritonConfig
}

func (b *basicTritonStateBag) Client() interface{} {
	b.l.RLock()
	defer b.l.RUnlock()

	return b.TritonClient
}

func (b *basicTritonStateBag) PutClient(client interface{}) {
	b.l.Lock()
	defer b.l.Unlock()

	b.TritonClient = client
}

func (b *basicTritonStateBag) AppendError(err error) {
	b.l.Lock()
	defer b.l.Unlock()

	if b.errors == nil {
		b.errors = make([]error, 0, 1)
	}

	b.errors = append(b.errors, err)
}

func (b *basicTritonStateBag) ErrorsOrNil() []error {
	b.l.RLock()
	defer b.l.RUnlock()

	return b.errors
}

func (b *basicTritonStateBag) Get(k string) interface{} {
	result, _ := b.GetOk(k)
	return result
}

func (b *basicTritonStateBag) GetOk(k string) (interface{}, bool) {
	b.l.RLock()
	defer b.l.RUnlock()

	result, ok := b.data[k]
	return result, ok
}

func (b *basicTritonStateBag) Put(k string, v interface{}) {
	b.l.Lock()
	defer b.l.Unlock()

	// Make sure the map is initialized one time, on write
	b.once.Do(func() {
		b.data = make(map[string]interface{})
	})

	// Write the data
	b.data[k] = v
}

func (b *basicTritonStateBag) Remove(k string) {
	b.l.Lock()
	defer b.l.Unlock()

	delete(b.data, k)
}
