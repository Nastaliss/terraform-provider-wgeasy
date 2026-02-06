package datasource_client

import (
	"context"

	"github.com/Nastaliss/terraform-provider-wgeasy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// clientModel maps a single wg-easy client to Terraform state.
type clientModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Enabled             types.Bool   `tfsdk:"enabled"`
	IPv4Address         types.String `tfsdk:"ipv4_address"`
	IPv6Address         types.String `tfsdk:"ipv6_address"`
	PublicKey           types.String `tfsdk:"public_key"`
	PrivateKey          types.String `tfsdk:"private_key"`
	PresharedKey        types.String `tfsdk:"preshared_key"`
	ExpiresAt           types.String `tfsdk:"expires_at"`
	AllowedIPs          types.List   `tfsdk:"allowed_ips"`
	ServerAllowedIPs    types.List   `tfsdk:"server_allowed_ips"`
	DNS                 types.List   `tfsdk:"dns"`
	MTU                 types.Int64  `tfsdk:"mtu"`
	PersistentKeepalive types.Int64  `tfsdk:"persistent_keepalive"`
	ServerEndpoint      types.String `tfsdk:"server_endpoint"`
	PreUp               types.String `tfsdk:"pre_up"`
	PostUp              types.String `tfsdk:"post_up"`
	PreDown             types.String `tfsdk:"pre_down"`
	PostDown            types.String `tfsdk:"post_down"`
	JC                  types.Int64  `tfsdk:"jc"`
	JMin                types.Int64  `tfsdk:"j_min"`
	JMax                types.Int64  `tfsdk:"j_max"`
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
}

func mapClientToModel(ctx context.Context, apiClient *client.Client, model *clientModel, diags *diag.Diagnostics) {
	model.ID = types.StringValue(apiClient.ID.String())
	model.Name = types.StringValue(apiClient.Name)
	model.Enabled = types.BoolValue(apiClient.Enabled)
	model.IPv4Address = types.StringValue(apiClient.IPv4Address)
	model.IPv6Address = types.StringValue(apiClient.IPv6Address)
	model.PublicKey = types.StringValue(apiClient.PublicKey)
	model.PrivateKey = types.StringValue(apiClient.PrivateKey)
	model.PresharedKey = types.StringValue(apiClient.PresharedKey)
	model.CreatedAt = types.StringValue(apiClient.CreatedAt)
	model.UpdatedAt = types.StringValue(apiClient.UpdatedAt)
	model.PreUp = types.StringValue(apiClient.PreUp)
	model.PostUp = types.StringValue(apiClient.PostUp)
	model.PreDown = types.StringValue(apiClient.PreDown)
	model.PostDown = types.StringValue(apiClient.PostDown)
	model.JC = types.Int64Value(apiClient.JC)
	model.JMin = types.Int64Value(apiClient.JMin)
	model.JMax = types.Int64Value(apiClient.JMax)

	if apiClient.ExpiresAt != nil {
		model.ExpiresAt = types.StringValue(*apiClient.ExpiresAt)
	} else {
		model.ExpiresAt = types.StringNull()
	}

	model.MTU = types.Int64Value(apiClient.MTU)
	model.PersistentKeepalive = types.Int64Value(apiClient.PersistentKeepalive)
	if apiClient.ServerEndpoint != nil {
		model.ServerEndpoint = types.StringValue(*apiClient.ServerEndpoint)
	} else {
		model.ServerEndpoint = types.StringNull()
	}

	// Ensure nil slices become empty lists (not null) for consistency.
	allowedIPsSlice := apiClient.AllowedIPs
	if allowedIPsSlice == nil {
		allowedIPsSlice = []string{}
	}
	allowedIPs, d := types.ListValueFrom(ctx, types.StringType, allowedIPsSlice)
	diags.Append(d...)
	model.AllowedIPs = allowedIPs

	serverAllowedIPsSlice := apiClient.ServerAllowedIPs
	if serverAllowedIPsSlice == nil {
		serverAllowedIPsSlice = []string{}
	}
	serverAllowedIPs, d := types.ListValueFrom(ctx, types.StringType, serverAllowedIPsSlice)
	diags.Append(d...)
	model.ServerAllowedIPs = serverAllowedIPs

	dnsSlice := apiClient.DNS
	if dnsSlice == nil {
		dnsSlice = []string{}
	}
	dns, d := types.ListValueFrom(ctx, types.StringType, dnsSlice)
	diags.Append(d...)
	model.DNS = dns
}
