package aperture

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetConfig(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/aperture/config" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("method = %q", r.Method)
		}
		w.Header().Set("ETag", `"abc123"`)
		_, _ = w.Write([]byte(`{"config":"// hello\n{\"providers\": {}}"}`))
	}))
	defer srv.Close()

	c := NewClient(ClientConfig{Endpoint: srv.URL + "/aperture"})
	cfg, etag, err := c.GetConfig(context.Background())
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if etag != `"abc123"` {
		t.Errorf("etag = %q", etag)
	}
	if !strings.Contains(cfg, "providers") {
		t.Errorf("config = %q (no providers)", cfg)
	}
}

func TestSetConfig_IfMatch(t *testing.T) {
	const want = `"prev-etag"`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("If-Match"); got != want {
			t.Errorf("If-Match = %q, want %q", got, want)
		}
		raw, _ := io.ReadAll(r.Body)
		var env configEnvelope
		_ = json.Unmarshal(raw, &env)
		if !strings.Contains(env.Config, "providers") {
			t.Errorf("body config = %q", env.Config)
		}
		w.Header().Set("ETag", `"new-etag"`)
		_, _ = w.Write([]byte(`{"config":"{}"}`))
	}))
	defer srv.Close()

	c := NewClient(ClientConfig{Endpoint: srv.URL})
	_, etag, err := c.SetConfig(context.Background(), `{"providers":{}}`, want)
	if err != nil {
		t.Fatalf("SetConfig: %v", err)
	}
	if etag != `"new-etag"` {
		t.Errorf("etag = %q", etag)
	}
}

func TestSetConfig_PreconditionFailed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusPreconditionFailed)
		_, _ = w.Write([]byte(`{"title":"Precondition Failed","status":412}`))
	}))
	defer srv.Close()

	c := NewClient(ClientConfig{Endpoint: srv.URL})
	if _, _, err := c.SetConfig(context.Background(), `{}`, `"stale"`); !errors.Is(err, ErrPreconditionFailed) {
		t.Errorf("err = %v, want ErrPreconditionFailed", err)
	}
}

func TestValidateConfig(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/config:validate" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"valid":false,"errors":["providers: at least one is required"]}`))
	}))
	defer srv.Close()

	c := NewClient(ClientConfig{Endpoint: srv.URL})
	vr, err := c.ValidateConfig(context.Background(), `{}`)
	if err != nil {
		t.Fatalf("ValidateConfig: %v", err)
	}
	if vr.Valid {
		t.Errorf("valid = true, want false")
	}
	if len(vr.Errors) != 1 {
		t.Errorf("errors = %v", vr.Errors)
	}
}

func TestProblemError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"title":"Forbidden","detail":"admin role required","status":403}`))
	}))
	defer srv.Close()

	c := NewClient(ClientConfig{Endpoint: srv.URL})
	_, _, err := c.GetConfig(context.Background())
	if err == nil {
		t.Fatal("err is nil, want a problem-formatted error")
	}
	if !strings.Contains(err.Error(), "Forbidden") || !strings.Contains(err.Error(), "admin role required") {
		t.Errorf("err = %q (missing Problem Details fields)", err)
	}
}

func TestParseAndMarshal_Roundtrip(t *testing.T) {
	in := []byte(`// top comment
{
  "providers": {
    "openai": {
      "baseurl": "https://api.openai.com/v1",
      "models": ["openai/gpt-5.5"],
    },
  },
  "auto_cost_basis": true,
}`)
	cfg, err := ParseConfigDocument(string(in))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got := cfg.Providers["openai"].BaseURL; got != "https://api.openai.com/v1" {
		t.Errorf("baseurl = %q", got)
	}
	if cfg.AutoCostBasis == nil || !*cfg.AutoCostBasis {
		t.Errorf("auto_cost_basis = %v", cfg.AutoCostBasis)
	}
	out, err := MarshalConfigDocument(cfg)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !strings.Contains(out, `"openai/gpt-5.5"`) {
		t.Errorf("marshalled = %q", out)
	}
}
