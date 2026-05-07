package aperture

import (
	"errors"
	"net/http/httptest"
	"testing"
)

func TestClient_DoReturnsAPINotPublic(t *testing.T) {
	c := NewClient(Config{Endpoint: "https://example.invalid", AuthToken: "x"})
	req := httptest.NewRequest("GET", "/", nil)
	if _, err := c.Do(req); !errors.Is(err, ErrAPINotPublic) {
		t.Errorf("expected ErrAPINotPublic, got %v", err)
	}
}

func TestClient_HasCredentials(t *testing.T) {
	cases := []struct {
		name     string
		cfg      Config
		wantHas  bool
	}{
		{"both set", Config{Endpoint: "x", AuthToken: "y"}, true},
		{"endpoint only", Config{Endpoint: "x"}, false},
		{"token only", Config{AuthToken: "y"}, false},
		{"empty", Config{}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := NewClient(tc.cfg)
			if got := c.HasCredentials(); got != tc.wantHas {
				t.Errorf("HasCredentials() = %v, want %v", got, tc.wantHas)
			}
		})
	}
}
