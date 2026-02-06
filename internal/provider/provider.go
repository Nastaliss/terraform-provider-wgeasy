package provider

import (
	"context"
	"os"

	"github.com/Nastaliss/terraform-provider-wgeasy/internal/client"
	"github.com/Nastaliss/terraform-provider-wgeasy/internal/datasource_client"
	"github.com/Nastaliss/terraform-provider-wgeasy/internal/resource_client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &wgeasyProvider{}

type wgeasyProvider struct{}

type wgeasyProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func New() provider.Provider {
	return &wgeasyProvider{}
}

func (p *wgeasyProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "wgeasy"
}

func (p *wgeasyProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for managing WireGuard clients on a wg-easy instance.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Description: "The URL of the wg-easy instance (e.g. http://localhost:51821). Can also be set via WGEASY_ENDPOINT environment variable.",
				Optional:    true,
			},
			"username": schema.StringAttribute{
				Description: "The username for wg-easy authentication. Can also be set via WGEASY_USERNAME environment variable.",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "The password for wg-easy authentication. Can also be set via WGEASY_PASSWORD environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

func (p *wgeasyProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config wgeasyProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := stringValueOrEnv(config.Endpoint, "WGEASY_ENDPOINT")
	username := stringValueOrEnv(config.Username, "WGEASY_USERNAME")
	password := stringValueOrEnv(config.Password, "WGEASY_PASSWORD")

	if endpoint == "" {
		resp.Diagnostics.AddError(
			"Missing endpoint",
			"The wg-easy endpoint must be set in the provider configuration or via the WGEASY_ENDPOINT environment variable.",
		)
	}
	if password == "" {
		resp.Diagnostics.AddError(
			"Missing password",
			"The wg-easy password must be set in the provider configuration or via the WGEASY_PASSWORD environment variable.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	apiClient, err := client.NewWGEasyClient(endpoint, username, password)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create API client", err.Error())
		return
	}

	resp.ResourceData = apiClient
	resp.DataSourceData = apiClient
}

func (p *wgeasyProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resource_client.NewClientResource,
	}
}

func (p *wgeasyProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasource_client.NewClientDataSource,
		datasource_client.NewClientsDataSource,
	}
}

func stringValueOrEnv(val types.String, envVar string) string {
	if !val.IsNull() && !val.IsUnknown() {
		return val.ValueString()
	}
	return os.Getenv(envVar)
}
