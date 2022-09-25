package provider

import (
	"context"

	"github.com/devopzilla/guku-client-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure GukuProvider satisfies various provider interfaces.
var _ provider.Provider = &GukuProvider{}
var _ provider.ProviderWithMetadata = &GukuProvider{}

const DEFAULT_ENDPOINT = "https://ztvgrcfy5bcvra2jmlfhsjw2ve.appsync-api.eu-north-1.amazonaws.com/graphql"

// GukuProvider defines the provider implementation.
type GukuProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// GukuProviderModel describes the provider data model.
type GukuProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func (p *GukuProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "guku"
	resp.Version = p.version
}

func (p *GukuProvider) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"endpoint": {
				MarkdownDescription: "Example provider attribute",
				Optional:            true,
				Type:                types.StringType,
			},
			"username": {
				MarkdownDescription: "Example provider attribute",
				Required:            true,
				Type:                types.StringType,
			},
			"password": {
				MarkdownDescription: "Example provider attribute",
				Required:            true,
				Type:                types.StringType,
				Sensitive:           true,
			},
		},
	}, nil
}

func (p *GukuProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data GukuProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.Endpoint.IsNull() {
		data.Endpoint = types.String{Value: DEFAULT_ENDPOINT}
		// resp.Diagnostics.AddError(
		// 	"Unable to find Endpoint",
		// 	"Endpoint cannot be empty",
		// )
		// return
	}
	if data.Username.IsNull() {
		resp.Diagnostics.AddError(
			"Unable to find Username",
			"Username cannot be empty",
		)
		return
	}
	if data.Password.IsNull() {
		resp.Diagnostics.AddError(
			"Unable to find Password",
			"Password cannot be empty",
		)
		return
	}

	client, err := guku.NewClient(context.TODO(), data.Endpoint.Value, data.Username.Value, data.Password.Value)
	tflog.Info(ctx, data.Username.Value)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create client",
			"Unable to create guku client:\n\n"+err.Error(),
		)
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *GukuProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewClusterResource,
		NewPlatformBindingResource,
	}
}

func (p *GukuProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewPlatformDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &GukuProvider{
			version: version,
		}
	}
}
