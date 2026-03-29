// Package forme provides a Go client for the Forme hosted PDF API.
//
// The client supports synchronous rendering, async rendering, S3 upload,
// PDF merging, and embedded data extraction. It uses only the Go standard
// library — no third-party dependencies.
package forme

import (
	"net/http"
)

const defaultBaseURL = "https://api.formepdf.com"

// Forme is the client for the Forme hosted PDF API.
type Forme struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// Option configures a Forme client.
type Option func(*Forme)

// WithBaseURL sets a custom base URL for the API.
func WithBaseURL(url string) Option {
	return func(f *Forme) {
		f.baseURL = url
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(f *Forme) {
		f.client = client
	}
}

// New creates a new Forme API client.
func New(apiKey string, opts ...Option) *Forme {
	f := &Forme{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		client:  http.DefaultClient,
	}
	for _, opt := range opts {
		opt(f)
	}
	// Strip trailing slash
	for len(f.baseURL) > 0 && f.baseURL[len(f.baseURL)-1] == '/' {
		f.baseURL = f.baseURL[:len(f.baseURL)-1]
	}
	return f
}

// FormeError is returned on non-2xx responses from the Forme API.
type FormeError struct {
	Status  int
	Message string
}

func (e *FormeError) Error() string {
	return e.Message
}
