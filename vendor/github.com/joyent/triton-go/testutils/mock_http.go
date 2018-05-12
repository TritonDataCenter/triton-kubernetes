package testutils

import (
	"errors"
	"net/http"

	"github.com/joyent/triton-go/authentication"
	"github.com/joyent/triton-go/client"
)

// Responders are callbacks that receive http requests and return a mocked
// response.
type Responder func(*http.Request) (*http.Response, error)

// NoResponderFound is returned when no responders are found for a given HTTP
// method and URL.
var NoResponderFound = errors.New("no responder found")

// MockTransport implements http.RoundTripper, which fulfills single http
// requests issued by an http.Client.  This implementation doesn't actually make
// the call, instead defering to the registered list of responders.
type MockTransport struct {
	FailNoResponder bool
	responders      map[string]Responder
}

// RoundTrip is required to implement http.MockTransport.  Instead of fulfilling
// the given request, the internal list of responders is consulted to handle the
// request.  If no responder is found an error is returned, which is the
// equivalent of a network error.
func (m *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	key := req.Method + " " + req.URL.String()

	// scan through the responders and find one that matches our key
	for k, r := range m.responders {
		if k != key {
			continue
		}
		return r(req)
	}

	// if we've been told to error when no match was found
	if m.FailNoResponder {
		return nil, NoResponderFound
	}

	// fallback to the default roundtripper
	return http.DefaultTransport.RoundTrip(req)
}

// RegisterResponder adds a new responder, associated with a given HTTP method
// and URL.  When a request comes in that matches, the responder will be called
// and the response returned to the client.
func (m *MockTransport) RegisterResponder(method, url string, responder Responder) {
	m.responders[method+" "+url] = responder
}

// Clear clears out all the mock responders that have been set on a particular
// MockTransport object. This comes in especially handy when utilizing the
// global DefaultMockTRansport and is utilized by the DeactivateClient func.
func (m *MockTransport) Clear() {
	m.responders = make(map[string]Responder)
}

// DefaultMockTransport allows users to easily and globally alter the default
// RoundTripper for all http requests.
var DefaultMockTransport = &MockTransport{
	responders: make(map[string]Responder),
}

// Activate replaces the `Transport` on the `http.Client` with our
// `DefaultMockTransport`.
func ActivateClient(failNoResponder bool) {
	DefaultMockTransport.FailNoResponder = failNoResponder
	http.DefaultClient.Transport = DefaultMockTransport
}

// Deactivate replaces our `DefaultMockTransport` with the
// `http.DefaultTransport`.
func DeactivateClient() {
	DefaultMockTransport.Clear()
	http.DefaultClient.Transport = http.DefaultTransport
}

// RegisterResponder adds a responder to the `DefaultMockTransport` responder
// table.
func RegisterResponder(method, url string, responder Responder) {
	DefaultMockTransport.RegisterResponder(method, url, responder)
}

// DefaultMockClient uses NewMockClient to construct a mocked out client.Client
var DefaultMockClient = NewMockClient(MockClientInput{})

type MockClientInput struct {
	AccountName string
}

// NewMockClient returns a new client.Client that includes our
// DefaultMockTransport which allows us to attach custom HTTP client request
// responders.
func NewMockClient(input MockClientInput) *client.Client {
	testSigner, _ := authentication.NewTestSigner()

	httpClient := &http.Client{
		Transport: DefaultMockTransport,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	client := &client.Client{
		Authorizers: []authentication.Signer{testSigner},
		HTTPClient:  httpClient,
	}

	if input.AccountName != "" {
		client.AccountName = input.AccountName
	}

	return client
}
