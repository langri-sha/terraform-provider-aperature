package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/langri-sha/aperature/internal/aperture"
)

// aperature_grant is a resource scaffold. Aperture has no public
// management API yet (see internal/aperture), so Create returns the
// canonical "API not yet public" error. The schema is finalized so
// users can write against it today; once upstream ships, only the CRUD
// methods need filling in.

func newGrantResource() resource.Resource { return &grantResource{} }

type grantResource struct {
	client *aperture.Client
}

type grantModel struct {
	ID           types.String            `tfsdk:"id"`
	Src          []types.String          `tfsdk:"src"`
	Capabilities []configCapabilityModel `tfsdk:"capabilities"`
}

func (r *grantResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_grant"
}

func (r *grantResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "An Aperture grant — a single entry in the top-level `grants` array.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Synthetic ID assigned by the provider.",
			},
			"src": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: "Source matchers: emails, group:NAME, or '*' for all.",
			},
			"capabilities": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"role":   schema.StringAttribute{Optional: true, Description: "user | admin"},
						"models": schema.StringAttribute{Optional: true, Description: "Glob in provider/model form, e.g. anthropic/**"},
					},
				},
			},
		},
	}
}

func (r *grantResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *grantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError("upstream API not yet public", apiNotPublicMsg(r.client))
}

func (r *grantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.AddError("upstream API not yet public", apiNotPublicMsg(r.client))
}

func (r *grantResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("upstream API not yet public", apiNotPublicMsg(r.client))
}

func (r *grantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError("upstream API not yet public", apiNotPublicMsg(r.client))
}

func apiNotPublicMsg(c *aperture.Client) string {
	if c != nil && c.HasCredentials() {
		return errors.Join(
			aperture.ErrAPINotPublic,
			errors.New("until then, render an aperture.json with `data.aperature_config` and ship it via the dashboard JSON editor or aperture-cli"),
		).Error()
	}
	return aperture.ErrAPINotPublic.Error()
}
