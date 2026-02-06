package resource_client

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// clientResourceModel maps the resource schema to a Go struct.
type clientResourceModel struct {
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
