package provider

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestRenderConfig_OmitsAbsentFields(t *testing.T) {
	m := &configModel{
		Providers: map[string]configProviderModel{
			"openai": {
				BaseURL: types.StringValue("https://api.openai.com/v1"),
				Models:  []types.String{types.StringValue("openai/gpt-5.5")},
			},
		},
	}
	doc, err := renderConfig(m)
	if err != nil {
		t.Fatalf("renderConfig: %v", err)
	}
	provs, ok := doc["providers"].(map[string]any)
	if !ok {
		t.Fatalf("providers: %#v", doc["providers"])
	}
	openai, ok := provs["openai"].(map[string]any)
	if !ok {
		t.Fatalf("openai: %#v", provs["openai"])
	}
	if _, has := openai["apikey"]; has {
		t.Errorf("apikey should be omitted when null, got %#v", openai)
	}
	if _, has := doc["grants"]; has {
		t.Errorf("grants should be omitted when empty, got %#v", doc)
	}
	if _, has := doc["auto_cost_basis"]; has {
		t.Errorf("auto_cost_basis should be omitted when null, got %#v", doc)
	}
}

func TestRenderConfig_GrantShape(t *testing.T) {
	m := &configModel{
		Grants: []configGrantModel{
			{
				Src: []types.String{types.StringValue("*")},
				Capabilities: []configCapabilityModel{
					{Role: types.StringValue("user")},
					{Models: types.StringValue("openai/**")},
				},
			},
		},
	}
	doc, err := renderConfig(m)
	if err != nil {
		t.Fatalf("renderConfig: %v", err)
	}
	out, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	// Aperture grants live under app["tailscale.com/cap/aperture"].
	// Verify that capability key is present verbatim.
	want := `"tailscale.com/cap/aperture"`
	if !contains(string(out), want) {
		t.Errorf("expected %s in output, got %s", want, out)
	}
}

func TestRenderConfig_GrantWithoutRoleOrModelsErrors(t *testing.T) {
	m := &configModel{
		Grants: []configGrantModel{
			{
				Src:          []types.String{types.StringValue("*")},
				Capabilities: []configCapabilityModel{{}},
			},
		},
	}
	if _, err := renderConfig(m); err == nil {
		t.Fatal("expected error for empty capability, got nil")
	}
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
