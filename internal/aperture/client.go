// Package aperture is a thin HTTP client for the Aperture admin API.
//
// As of 2026-05 Aperture does not expose a public administrative HTTP
// API; configuration is a single JSON document edited via the
// dashboard's JSON editor or supplied to aperture-cli. This package
// therefore exists as a placeholder: the public surface is settled
// (Config, NewClient, Do) so consumers can compile against it, but
// every Do() call returns ErrAPINotPublic until the upstream API
// arrives. When that happens, fill in Do (and add typed helpers like
// GetConfig, PutConfig) without changing the package boundary.
package aperture

import (
	"crypto/tls"
	"errors"
	"net/http"
	"time"
)

// ErrAPINotPublic is returned from every Do() call until Tailscale
// publishes an Aperture management API.
var ErrAPINotPublic = errors.New("aperture: management API is not yet public; configure via the Aperture dashboard JSON editor or aperture-cli")

// Config configures a Client.
type Config struct {
	Endpoint  string
	AuthToken string
	Insecure  bool
	UserAgent string
}

// Client talks to the Aperture admin endpoint.
type Client struct {
	cfg  Config
	http *http.Client
}

// NewClient returns a configured Client. Endpoint and AuthToken may be
// empty: Do() simply returns ErrAPINotPublic regardless until upstream
// catches up, and we want practitioners to be able to instantiate the
// provider without an endpoint just to use the offline data sources.
func NewClient(cfg Config) *Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if cfg.Insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return &Client{
		cfg: cfg,
		http: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
	}
}

// Endpoint returns the configured endpoint, or "" if none was set.
func (c *Client) Endpoint() string { return c.cfg.Endpoint }

// HasCredentials reports whether both endpoint and auth token are set.
// Resources that require live API access can short-circuit with a
// helpful error when this is false.
func (c *Client) HasCredentials() bool {
	return c.cfg.Endpoint != "" && c.cfg.AuthToken != ""
}

// Do is the single execution point for HTTP calls. Today it always
// returns ErrAPINotPublic. Once the upstream API ships, replace this
// body with a real round-trip that injects User-Agent and Authorization.
func (c *Client) Do(_ *http.Request) (*http.Response, error) {
	return nil, ErrAPINotPublic
}
