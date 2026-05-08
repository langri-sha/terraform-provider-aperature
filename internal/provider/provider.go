// Package provider implements the terraform-provider-aperature plugin.
package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/langri-sha/aperature/internal/aperture"
)

// New returns a providerserver factory.
func New(version, commit string) func() provider.Provider {
	return func() provider.Provider {
		return &apertureProvider{version: version, commit: commit}
	}
}

type apertureProvider struct {
	version string
	commit  string
}

type providerModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
}

func (p *apertureProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "aperature"
	resp.Version = p.version
}

func (p *apertureProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for Aperture by Tailscale. Authentication is by Tailscale identity at the network layer — the caller must be on the tailnet with the admin role granted in Aperture's configuration.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Required:    true,
				Description: "Full base URL of the Aperture admin API including the /aperture path prefix, e.g. http://ai.<tailnet>.ts.net/aperture.",
			},
		},
	}
}

func (p *apertureProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client := aperture.NewClient(aperture.ClientConfig{
		Endpoint:  data.Endpoint.ValueString(),
		UserAgent: "terraform-provider-aperature/" + p.version,
	})

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *apertureProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		newConfigResource,
	}
}

func (p *apertureProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}
