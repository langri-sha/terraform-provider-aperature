package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// aperature_config renders a JSON document matching the documented
// Aperture config schema. This data source is the workhorse of the
// pre-alpha provider: until upstream ships a management API, the
// canonical way to manage Aperture is to edit a single JSON document,
// and rendering it from typed HCL is genuinely useful.
//
// Field names mirror the upstream JSON keys verbatim (baseurl, apikey,
// add_headers, on_exceed, ...) — see AGENTS.md.

func newConfigDataSource() datasource.DataSource { return &configDataSource{} }

type configDataSource struct{}

type configModel struct {
	Providers       map[string]configProviderModel `tfsdk:"providers"`
	Grants          []configGrantModel             `tfsdk:"grants"`
	Quotas          map[string]configQuotaModel    `tfsdk:"quotas"`
	Hooks           map[string]configHookModel     `tfsdk:"hooks"`
	AutoCostBasis   types.Bool                     `tfsdk:"auto_cost_basis"`
	JSON            types.String                   `tfsdk:"json"`
}

type configProviderModel struct {
	BaseURL       types.String      `tfsdk:"baseurl"`
	Models        []types.String    `tfsdk:"models"`
	APIKey        types.String      `tfsdk:"apikey"`
	Authorization types.String      `tfsdk:"authorization"`
	Name          types.String      `tfsdk:"name"`
	Description   types.String      `tfsdk:"description"`
	CostBasis     types.String      `tfsdk:"cost_basis"`
	Preference    types.Int64       `tfsdk:"preference"`
	Disabled      types.Bool        `tfsdk:"disabled"`
	AddHeaders    map[string]string `tfsdk:"add_headers"`
}

type configGrantModel struct {
	Src          []types.String         `tfsdk:"src"`
	Capabilities []configCapabilityModel `tfsdk:"capabilities"`
}

// One capability entry maps to one element of the
// app["tailscale.com/cap/aperture"] array. Upstream allows either
// {role: ...} or {models: ...}; we model both as optional so callers
// emit one entry per capability.
type configCapabilityModel struct {
	Role   types.String `tfsdk:"role"`
	Models types.String `tfsdk:"models"`
}

type configQuotaModel struct {
	Capacity types.Float64 `tfsdk:"capacity"`
	Rate     types.Float64 `tfsdk:"rate"`
	OnExceed types.String  `tfsdk:"on_exceed"`
}

type configHookModel struct {
	URL           types.String `tfsdk:"url"`
	APIKey        types.String `tfsdk:"apikey"`
	Authorization types.String `tfsdk:"authorization"`
	Timeout       types.String `tfsdk:"timeout"`
	Disabled      types.Bool   `tfsdk:"disabled"`
	FailPolicy    types.String `tfsdk:"fail_policy"`
	Preference    types.Int64  `tfsdk:"preference"`
}

func (d *configDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config"
}

func (d *configDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Renders an Aperture configuration JSON document from typed HCL. Pipe `json` into a `local_file` to manage `aperture.json` declaratively.",
		Attributes: map[string]schema.Attribute{
			"providers": schema.MapNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"baseurl":       schema.StringAttribute{Required: true},
						"models":        schema.ListAttribute{ElementType: types.StringType, Required: true},
						"apikey":        schema.StringAttribute{Optional: true, Sensitive: true},
						"authorization": schema.StringAttribute{Optional: true, Sensitive: true},
						"name":          schema.StringAttribute{Optional: true},
						"description":   schema.StringAttribute{Optional: true},
						"cost_basis":    schema.StringAttribute{Optional: true},
						"preference":    schema.Int64Attribute{Optional: true},
						"disabled":      schema.BoolAttribute{Optional: true},
						"add_headers":   schema.MapAttribute{ElementType: types.StringType, Optional: true},
					},
				},
			},
			"grants": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"src": schema.ListAttribute{ElementType: types.StringType, Required: true},
						"capabilities": schema.ListNestedAttribute{
							Required: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"role":   schema.StringAttribute{Optional: true},
									"models": schema.StringAttribute{Optional: true},
								},
							},
						},
					},
				},
			},
			"quotas": schema.MapNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"capacity":  schema.Float64Attribute{Required: true},
						"rate":      schema.Float64Attribute{Required: true},
						"on_exceed": schema.StringAttribute{Optional: true},
					},
				},
			},
			"hooks": schema.MapNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"url":           schema.StringAttribute{Required: true},
						"apikey":        schema.StringAttribute{Optional: true, Sensitive: true},
						"authorization": schema.StringAttribute{Optional: true, Sensitive: true},
						"timeout":       schema.StringAttribute{Optional: true},
						"disabled":      schema.BoolAttribute{Optional: true},
						"fail_policy":   schema.StringAttribute{Optional: true},
						"preference":    schema.Int64Attribute{Optional: true},
					},
				},
			},
			"auto_cost_basis": schema.BoolAttribute{Optional: true},
			"json": schema.StringAttribute{
				Computed:    true,
				Description: "The rendered Aperture configuration JSON.",
			},
		},
	}
}

func (d *configDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data configModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	doc, err := renderConfig(&data)
	if err != nil {
		resp.Diagnostics.AddError("render aperture config", err.Error())
		return
	}
	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		resp.Diagnostics.AddError("marshal aperture config", err.Error())
		return
	}

	data.JSON = types.StringValue(string(out))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// renderConfig converts the Terraform model into a marshallable map
// matching the upstream schema. We use a map instead of a typed struct
// so unset fields are simply omitted rather than emitted as zero
// values, matching Aperture's "absent means default" semantics.
func renderConfig(m *configModel) (map[string]any, error) {
	doc := map[string]any{}

	if len(m.Providers) > 0 {
		providers := make(map[string]any, len(m.Providers))
		for name, p := range m.Providers {
			entry := map[string]any{
				"baseurl": p.BaseURL.ValueString(),
				"models":  stringsFromList(p.Models),
			}
			if !p.APIKey.IsNull() {
				entry["apikey"] = p.APIKey.ValueString()
			}
			if !p.Authorization.IsNull() {
				entry["authorization"] = p.Authorization.ValueString()
			}
			if !p.Name.IsNull() {
				entry["name"] = p.Name.ValueString()
			}
			if !p.Description.IsNull() {
				entry["description"] = p.Description.ValueString()
			}
			if !p.CostBasis.IsNull() {
				entry["cost_basis"] = p.CostBasis.ValueString()
			}
			if !p.Preference.IsNull() {
				entry["preference"] = p.Preference.ValueInt64()
			}
			if !p.Disabled.IsNull() {
				entry["disabled"] = p.Disabled.ValueBool()
			}
			if len(p.AddHeaders) > 0 {
				entry["add_headers"] = p.AddHeaders
			}
			providers[name] = entry
		}
		doc["providers"] = providers
	}

	if len(m.Grants) > 0 {
		grants := make([]any, 0, len(m.Grants))
		for _, g := range m.Grants {
			caps := make([]map[string]any, 0, len(g.Capabilities))
			for _, c := range g.Capabilities {
				cap := map[string]any{}
				if !c.Role.IsNull() {
					cap["role"] = c.Role.ValueString()
				}
				if !c.Models.IsNull() {
					cap["models"] = c.Models.ValueString()
				}
				if len(cap) == 0 {
					return nil, fmt.Errorf("grant capability must set role or models")
				}
				caps = append(caps, cap)
			}
			grants = append(grants, map[string]any{
				"src": stringsFromList(g.Src),
				"app": map[string]any{
					"tailscale.com/cap/aperture": caps,
				},
			})
		}
		doc["grants"] = grants
	}

	if len(m.Quotas) > 0 {
		quotas := make(map[string]any, len(m.Quotas))
		for name, q := range m.Quotas {
			entry := map[string]any{
				"capacity": q.Capacity.ValueFloat64(),
				"rate":     q.Rate.ValueFloat64(),
			}
			if !q.OnExceed.IsNull() {
				entry["on_exceed"] = q.OnExceed.ValueString()
			}
			quotas[name] = entry
		}
		doc["quotas"] = quotas
	}

	if len(m.Hooks) > 0 {
		hooks := make(map[string]any, len(m.Hooks))
		for name, h := range m.Hooks {
			entry := map[string]any{
				"url": h.URL.ValueString(),
			}
			if !h.APIKey.IsNull() {
				entry["apikey"] = h.APIKey.ValueString()
			}
			if !h.Authorization.IsNull() {
				entry["authorization"] = h.Authorization.ValueString()
			}
			if !h.Timeout.IsNull() {
				entry["timeout"] = h.Timeout.ValueString()
			}
			if !h.Disabled.IsNull() {
				entry["disabled"] = h.Disabled.ValueBool()
			}
			if !h.FailPolicy.IsNull() {
				entry["fail_policy"] = h.FailPolicy.ValueString()
			}
			if !h.Preference.IsNull() {
				entry["preference"] = h.Preference.ValueInt64()
			}
			hooks[name] = entry
		}
		doc["hooks"] = hooks
	}

	if !m.AutoCostBasis.IsNull() {
		doc["auto_cost_basis"] = m.AutoCostBasis.ValueBool()
	}

	return doc, nil
}

func stringsFromList(in []types.String) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		out = append(out, s.ValueString())
	}
	return out
}
