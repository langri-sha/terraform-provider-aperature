package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/langri-sha/aperature/internal/aperture"
)

// aperature_llm_provider is a resource scaffold for one entry of the
// top-level `providers` map (e.g. providers.openai). See grant_resource.go
// for the CRUD pattern; both resources are scaffolds today.

func newLLMProviderResource() resource.Resource { return &llmProviderResource{} }

type llmProviderResource struct {
	client *aperture.Client
}

type llmProviderModel struct {
	ID            types.String      `tfsdk:"id"`
	Name          types.String      `tfsdk:"name"`
	BaseURL       types.String      `tfsdk:"baseurl"`
	Models        []types.String    `tfsdk:"models"`
	APIKey        types.String      `tfsdk:"apikey"`
	Authorization types.String      `tfsdk:"authorization"`
	Description   types.String      `tfsdk:"description"`
	CostBasis     types.String      `tfsdk:"cost_basis"`
	Preference    types.Int64       `tfsdk:"preference"`
	Disabled      types.Bool        `tfsdk:"disabled"`
	AddHeaders    map[string]string `tfsdk:"add_headers"`
}

func (r *llmProviderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_llm_provider"
}

func (r *llmProviderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "An Aperture LLM provider — one entry in the top-level `providers` map.",
		Attributes: map[string]schema.Attribute{
			"id":            schema.StringAttribute{Computed: true},
			"name":          schema.StringAttribute{Required: true, Description: "Map key under top-level `providers` (e.g. \"openai\")."},
			"baseurl":       schema.StringAttribute{Required: true},
			"models":        schema.ListAttribute{ElementType: types.StringType, Required: true},
			"apikey":        schema.StringAttribute{Optional: true, Sensitive: true},
			"authorization": schema.StringAttribute{Optional: true, Sensitive: true},
			"description":   schema.StringAttribute{Optional: true},
			"cost_basis":    schema.StringAttribute{Optional: true},
			"preference":    schema.Int64Attribute{Optional: true},
			"disabled":      schema.BoolAttribute{Optional: true},
			"add_headers":   schema.MapAttribute{ElementType: types.StringType, Optional: true},
		},
	}
}

func (r *llmProviderResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*aperture.Client)
	if !ok {
		resp.Diagnostics.AddError("provider data type", "expected *aperture.Client")
		return
	}
	r.client = c
}

func (r *llmProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError("upstream API not yet public", apiNotPublicMsg(r.client))
}

func (r *llmProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.AddError("upstream API not yet public", apiNotPublicMsg(r.client))
}

func (r *llmProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("upstream API not yet public", apiNotPublicMsg(r.client))
}

func (r *llmProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError("upstream API not yet public", apiNotPublicMsg(r.client))
}
