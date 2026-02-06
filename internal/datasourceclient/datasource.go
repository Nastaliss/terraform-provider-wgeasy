// Package datasourceclient implements the wgeasy_client and wgeasy_clients data sources.
package datasourceclient

import (
	"context"
	"fmt"

	"github.com/Nastaliss/terraform-provider-wgeasy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &clientDataSource{}

type clientDataSource struct {
	apiClient *client.WGEasyClient
}

type clientDataSourceModel struct {
	clientModel
}

// NewClientDataSource creates a new wgeasy_client data source instance.
func NewClientDataSource() datasource.DataSource {
	return &clientDataSource{}
}

func (d *clientDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_client"
}

func (d *clientDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single WireGuard client/peer by ID from a wg-easy instance.",
		Attributes:  clientDataSourceAttributes(true),
	}
}

func (d *clientDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	apiClient, ok := req.ProviderData.(*client.WGEasyClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.WGEasyClient, got: %T", req.ProviderData),
		)
		return
	}
	d.apiClient = apiClient
}

func (d *clientDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state clientDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiClient, err := d.apiClient.GetClient(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading client", err.Error())
		return
	}

	mapClientToModel(ctx, apiClient, &state.clientModel, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// clientDataSourceAttributes returns the common attributes for client data sources.
// If idRequired is true, id is a required input; otherwise it's computed.
func clientDataSourceAttributes(idRequired bool) map[string]schema.Attribute {
	attrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "The ID of the client.",
			Required:    idRequired,
			Computed:    !idRequired,
		},
		"name": schema.StringAttribute{
			Description: "The name of the client.",
			Computed:    true,
		},
		"enabled": schema.BoolAttribute{
			Description: "Whether the client is enabled.",
			Computed:    true,
		},
		"ipv4_address": schema.StringAttribute{
			Description: "The IPv4 address assigned to the client.",
			Computed:    true,
		},
		"ipv6_address": schema.StringAttribute{
			Description: "The IPv6 address assigned to the client.",
			Computed:    true,
		},
		"public_key": schema.StringAttribute{
			Description: "The public key of the client.",
			Computed:    true,
		},
		"private_key": schema.StringAttribute{
			Description: "The private key of the client.",
			Computed:    true,
			Sensitive:   true,
		},
		"preshared_key": schema.StringAttribute{
			Description: "The preshared key of the client.",
			Computed:    true,
			Sensitive:   true,
		},
		"expires_at": schema.StringAttribute{
			Description: "The expiration date of the client.",
			Computed:    true,
		},
		"allowed_ips": schema.ListAttribute{
			Description: "List of allowed IPs for the client.",
			Computed:    true,
			ElementType: types.StringType,
		},
		"server_allowed_ips": schema.ListAttribute{
			Description: "List of server-side allowed IPs.",
			Computed:    true,
			ElementType: types.StringType,
		},
		"dns": schema.ListAttribute{
			Description: "List of DNS servers for the client.",
			Computed:    true,
			ElementType: types.StringType,
		},
		"mtu": schema.Int64Attribute{
			Description: "MTU value for the client.",
			Computed:    true,
		},
		"persistent_keepalive": schema.Int64Attribute{
			Description: "Persistent keepalive interval in seconds.",
			Computed:    true,
		},
		"server_endpoint": schema.StringAttribute{
			Description: "The server endpoint for the client.",
			Computed:    true,
		},
		"pre_up": schema.StringAttribute{
			Description: "Command to run before bringing up the interface.",
			Computed:    true,
		},
		"post_up": schema.StringAttribute{
			Description: "Command to run after bringing up the interface.",
			Computed:    true,
		},
		"pre_down": schema.StringAttribute{
			Description: "Command to run before bringing down the interface.",
			Computed:    true,
		},
		"post_down": schema.StringAttribute{
			Description: "Command to run after bringing down the interface.",
			Computed:    true,
		},
		"jc": schema.Int64Attribute{
			Description: "Jitter coefficient (jC) for WireGuard.",
			Computed:    true,
		},
		"j_min": schema.Int64Attribute{
			Description: "Minimum jitter value (jMin) for WireGuard.",
			Computed:    true,
		},
		"j_max": schema.Int64Attribute{
			Description: "Maximum jitter value (jMax) for WireGuard.",
			Computed:    true,
		},
		"created_at": schema.StringAttribute{
			Description: "The creation timestamp.",
			Computed:    true,
		},
		"updated_at": schema.StringAttribute{
			Description: "The last update timestamp.",
			Computed:    true,
		},
	}
	return attrs
}
