// Package provider implements the terraform-provider-aperature plugin.
package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/langri-sha/aperature/internal/aperture"
)

// Default name of the env var holding the Aperture admin auth token.
// We avoid the obvious APERTURE_AUTH_TOKEN to keep the env-var spelling
// aligned with the (mis)spelled provider for now; users can override
// via the provider block.
const envAuthToken = "APERATURE_AUTH_TOKEN"

// New returns a providerserver.NewProtocol6 factory.
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
	Endpoint  types.String `tfsdk:"endpoint"`
	AuthToken types.String `tfsdk:"auth_token"`
	Insecure  types.Bool   `tfsdk:"insecure"`
}

func (p *apertureProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "aperature"
	resp.Version = p.version
}

func (p *apertureProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for Aperture by Tailscale.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Optional:    true,
				Description: "Base URL of the Aperture admin endpoint, e.g. http://aperture.tailnet.ts.net.",
			},
			"auth_token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Admin auth token. Defaults to $" + envAuthToken + ".",
			},
			"insecure": schema.BoolAttribute{
				Optional:    true,
				Description: "Skip TLS verification. Useful for in-tailnet HTTP endpoints with self-signed certs.",
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

	endpoint := data.Endpoint.ValueString()
	token := data.AuthToken.ValueString()
	if token == "" {
		token = os.Getenv(envAuthToken)
	}
	insecure := data.Insecure.ValueBool()

	client := aperture.NewClient(aperture.Config{
		Endpoint:  endpoint,
		AuthToken: token,
		Insecure:  insecure,
		UserAgent: "terraform-provider-aperature/" + p.version,
	})

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *apertureProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		newGrantResource,
		newLLMProviderResource,
	}
}

func (p *apertureProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		newConfigDataSource,
	}
}
