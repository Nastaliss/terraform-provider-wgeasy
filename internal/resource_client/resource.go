package resource_client

import (
	"context"
	"fmt"

	"github.com/Nastaliss/terraform-provider-wgeasy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &clientResource{}
	_ resource.ResourceWithImportState = &clientResource{}
)

type clientResource struct {
	apiClient *client.WGEasyClient
}

func NewClientResource() resource.Resource {
	return &clientResource{}
}

func (r *clientResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_client"
}

func (r *clientResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a WireGuard client/peer on a wg-easy instance.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the client.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the client.",
				Required:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the client is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"ipv4_address": schema.StringAttribute{
				Description: "The IPv4 address assigned to the client.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ipv6_address": schema.StringAttribute{
				Description: "The IPv6 address assigned to the client.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"public_key": schema.StringAttribute{
				Description: "The public key of the client.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"private_key": schema.StringAttribute{
				Description: "The private key of the client.",
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"preshared_key": schema.StringAttribute{
				Description: "The preshared key of the client.",
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"expires_at": schema.StringAttribute{
				Description: "The expiration date of the client (ISO 8601 format).",
				Optional:    true,
			},
			"allowed_ips": schema.ListAttribute{
				Description: "List of allowed IPs for the client. Empty list means use server default.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"server_allowed_ips": schema.ListAttribute{
				Description: "List of server-side allowed IPs. Empty list means use server default.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"dns": schema.ListAttribute{
				Description: "List of DNS servers for the client. Empty list means use server default.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"mtu": schema.Int64Attribute{
				Description: "MTU value for the client (minimum 1024).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"persistent_keepalive": schema.Int64Attribute{
				Description: "Persistent keepalive interval in seconds.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"server_endpoint": schema.StringAttribute{
				Description: "The server endpoint for the client.",
				Optional:    true,
			},
			"pre_up": schema.StringAttribute{
				Description: "Command to run before bringing up the interface.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"post_up": schema.StringAttribute{
				Description: "Command to run after bringing up the interface.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"pre_down": schema.StringAttribute{
				Description: "Command to run before bringing down the interface.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"post_down": schema.StringAttribute{
				Description: "Command to run after bringing down the interface.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"jc": schema.Int64Attribute{
				Description: "Jitter coefficient (jC) for WireGuard.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"j_min": schema.Int64Attribute{
				Description: "Minimum jitter value (jMin) for WireGuard.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"j_max": schema.Int64Attribute{
				Description: "Maximum jitter value (jMax) for WireGuard.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "The creation timestamp.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "The last update timestamp.",
				Computed:    true,
			},
		},
	}
}

func (r *clientResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	apiClient, ok := req.ProviderData.(*client.WGEasyClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.WGEasyClient, got: %T", req.ProviderData),
		)
		return
	}
	r.apiClient = apiClient
}

func (r *clientResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan clientResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Step 1: Create the client (only name + expiresAt).
	createReq := client.CreateClientRequest{
		Name: plan.Name.ValueString(),
	}
	if !plan.ExpiresAt.IsNull() && !plan.ExpiresAt.IsUnknown() {
		v := plan.ExpiresAt.ValueString()
		createReq.ExpiresAt = &v
	}

	clientID, err := r.apiClient.CreateClient(createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating client", err.Error())
		return
	}

	// Step 2: If there are additional fields to set, fetch and update.
	if needsUpdate(plan) {
		current, err := r.apiClient.GetClient(clientID)
		if err != nil {
			resp.Diagnostics.AddError("Error reading client after creation", err.Error())
			return
		}
		updateReq := buildUpdateRequestFromCurrent(ctx, plan, current)
		_, err = r.apiClient.UpdateClient(clientID, updateReq)
		if err != nil {
			resp.Diagnostics.AddError("Error updating client after creation", err.Error())
			return
		}
	}

	// Step 3: Read back to get server-authoritative values.
	readBack, err := r.apiClient.GetClient(clientID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading client after creation", err.Error())
		return
	}

	mapClientToState(ctx, readBack, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *clientResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state clientResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiClient, err := r.apiClient.GetClient(state.ID.ValueString())
	if err != nil {
		if _, ok := err.(*client.NotFoundError); ok {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading client", err.Error())
		return
	}

	mapClientToState(ctx, apiClient, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *clientResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan clientResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state clientResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch current client state to merge with planned changes.
	current, err := r.apiClient.GetClient(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading client before update", err.Error())
		return
	}

	updateReq := buildUpdateRequestFromCurrent(ctx, plan, current)

	_, err = r.apiClient.UpdateClient(state.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating client", err.Error())
		return
	}

	readBack, err := r.apiClient.GetClient(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading client after update", err.Error())
		return
	}

	mapClientToState(ctx, readBack, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *clientResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state clientResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.apiClient.DeleteClient(state.ID.ValueString())
	if err != nil {
		if _, ok := err.(*client.NotFoundError); !ok {
			resp.Diagnostics.AddError("Error deleting client", err.Error())
		}
	}
}

func (r *clientResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// needsUpdate returns true if the plan has optional fields that need a follow-up update call.
func needsUpdate(plan clientResourceModel) bool {
	// Check if list fields have non-empty values.
	hasAllowedIPs := !plan.AllowedIPs.IsNull() && !plan.AllowedIPs.IsUnknown() && len(plan.AllowedIPs.Elements()) > 0
	hasServerAllowedIPs := !plan.ServerAllowedIPs.IsNull() && !plan.ServerAllowedIPs.IsUnknown() && len(plan.ServerAllowedIPs.Elements()) > 0
	hasDNS := !plan.DNS.IsNull() && !plan.DNS.IsUnknown() && len(plan.DNS.Elements()) > 0

	return hasAllowedIPs ||
		hasServerAllowedIPs ||
		hasDNS ||
		(!plan.MTU.IsNull() && !plan.MTU.IsUnknown()) ||
		(!plan.PersistentKeepalive.IsNull() && !plan.PersistentKeepalive.IsUnknown()) ||
		(!plan.ServerEndpoint.IsNull() && !plan.ServerEndpoint.IsUnknown()) ||
		(!plan.Enabled.IsNull() && !plan.Enabled.IsUnknown() && !plan.Enabled.ValueBool())
}

// buildUpdateRequestFromCurrent builds an update request starting from the current API state,
// then overlays any values from the plan.
func buildUpdateRequestFromCurrent(ctx context.Context, plan clientResourceModel, current *client.Client) client.UpdateClientRequest {
	updateReq := client.UpdateClientRequest{
		Name:                current.Name,
		Enabled:             current.Enabled,
		IPv4Address:         current.IPv4Address,
		IPv6Address:         current.IPv6Address,
		AllowedIPs:          current.AllowedIPs,
		ServerAllowedIPs:    current.ServerAllowedIPs,
		DNS:                 current.DNS,
		MTU:                 current.MTU,
		PersistentKeepalive: current.PersistentKeepalive,
		PreUp:               current.PreUp,
		PostUp:              current.PostUp,
		PreDown:             current.PreDown,
		PostDown:            current.PostDown,
		JC:                  current.JC,
		JMin:                current.JMin,
		JMax:                current.JMax,
		// Optional pointer fields - pass through as-is (nil means omit)
		ServerEndpoint: current.ServerEndpoint,
		ExpiresAt:      current.ExpiresAt,
		I1:             current.I1,
		I2:             current.I2,
		I3:             current.I3,
		I4:             current.I4,
		I5:             current.I5,
	}

	// Overlay plan values.
	updateReq.Name = plan.Name.ValueString()
	updateReq.Enabled = plan.Enabled.ValueBool()

	if !plan.ExpiresAt.IsNull() && !plan.ExpiresAt.IsUnknown() {
		v := plan.ExpiresAt.ValueString()
		updateReq.ExpiresAt = &v
	}

	if !plan.AllowedIPs.IsNull() && !plan.AllowedIPs.IsUnknown() {
		var ips []string
		plan.AllowedIPs.ElementsAs(ctx, &ips, false)
		if len(ips) > 0 {
			updateReq.AllowedIPs = ips
		} else {
			// Empty list means "use global config" - send null to API
			updateReq.AllowedIPs = nil
		}
	}

	if !plan.ServerAllowedIPs.IsNull() && !plan.ServerAllowedIPs.IsUnknown() {
		var ips []string
		plan.ServerAllowedIPs.ElementsAs(ctx, &ips, false)
		if len(ips) > 0 {
			updateReq.ServerAllowedIPs = ips
		} else {
			// Empty list means "use global config" - send null to API
			updateReq.ServerAllowedIPs = nil
		}
	}

	if !plan.DNS.IsNull() && !plan.DNS.IsUnknown() {
		var dns []string
		plan.DNS.ElementsAs(ctx, &dns, false)
		updateReq.DNS = dns
	}

	if !plan.MTU.IsNull() && !plan.MTU.IsUnknown() {
		updateReq.MTU = plan.MTU.ValueInt64()
	}

	if !plan.PersistentKeepalive.IsNull() && !plan.PersistentKeepalive.IsUnknown() {
		updateReq.PersistentKeepalive = plan.PersistentKeepalive.ValueInt64()
	}

	if !plan.ServerEndpoint.IsNull() && !plan.ServerEndpoint.IsUnknown() {
		v := plan.ServerEndpoint.ValueString()
		updateReq.ServerEndpoint = &v
	}

	if !plan.PreUp.IsNull() && !plan.PreUp.IsUnknown() {
		updateReq.PreUp = plan.PreUp.ValueString()
	}

	if !plan.PostUp.IsNull() && !plan.PostUp.IsUnknown() {
		updateReq.PostUp = plan.PostUp.ValueString()
	}

	if !plan.PreDown.IsNull() && !plan.PreDown.IsUnknown() {
		updateReq.PreDown = plan.PreDown.ValueString()
	}

	if !plan.PostDown.IsNull() && !plan.PostDown.IsUnknown() {
		updateReq.PostDown = plan.PostDown.ValueString()
	}

	if !plan.JC.IsNull() && !plan.JC.IsUnknown() {
		updateReq.JC = plan.JC.ValueInt64()
	}

	if !plan.JMin.IsNull() && !plan.JMin.IsUnknown() {
		updateReq.JMin = plan.JMin.ValueInt64()
	}

	if !plan.JMax.IsNull() && !plan.JMax.IsUnknown() {
		updateReq.JMax = plan.JMax.ValueInt64()
	}

	return updateReq
}

func mapClientToState(ctx context.Context, apiClient *client.Client, state *clientResourceModel, diags *diag.Diagnostics) {
	state.ID = types.StringValue(apiClient.ID.String())
	state.Name = types.StringValue(apiClient.Name)
	state.Enabled = types.BoolValue(apiClient.Enabled)
	state.IPv4Address = types.StringValue(apiClient.IPv4Address)
	state.IPv6Address = types.StringValue(apiClient.IPv6Address)
	state.PublicKey = types.StringValue(apiClient.PublicKey)
	state.PrivateKey = types.StringValue(apiClient.PrivateKey)
	state.PresharedKey = types.StringValue(apiClient.PresharedKey)
	state.CreatedAt = types.StringValue(apiClient.CreatedAt)
	state.UpdatedAt = types.StringValue(apiClient.UpdatedAt)
	state.PreUp = types.StringValue(apiClient.PreUp)
	state.PostUp = types.StringValue(apiClient.PostUp)
	state.PreDown = types.StringValue(apiClient.PreDown)
	state.PostDown = types.StringValue(apiClient.PostDown)
	state.JC = types.Int64Value(apiClient.JC)
	state.JMin = types.Int64Value(apiClient.JMin)
	state.JMax = types.Int64Value(apiClient.JMax)

	if apiClient.ExpiresAt != nil {
		state.ExpiresAt = types.StringValue(*apiClient.ExpiresAt)
	} else {
		state.ExpiresAt = types.StringNull()
	}

	state.MTU = types.Int64Value(apiClient.MTU)
	state.PersistentKeepalive = types.Int64Value(apiClient.PersistentKeepalive)
	if apiClient.ServerEndpoint != nil {
		state.ServerEndpoint = types.StringValue(*apiClient.ServerEndpoint)
	} else {
		state.ServerEndpoint = types.StringNull()
	}

	// Ensure nil slices become empty lists (not null) for consistency with plan.
	allowedIPsSlice := apiClient.AllowedIPs
	if allowedIPsSlice == nil {
		allowedIPsSlice = []string{}
	}
	allowedIPs, d := types.ListValueFrom(ctx, types.StringType, allowedIPsSlice)
	diags.Append(d...)
	state.AllowedIPs = allowedIPs

	serverAllowedIPsSlice := apiClient.ServerAllowedIPs
	if serverAllowedIPsSlice == nil {
		serverAllowedIPsSlice = []string{}
	}
	serverAllowedIPs, d := types.ListValueFrom(ctx, types.StringType, serverAllowedIPsSlice)
	diags.Append(d...)
	state.ServerAllowedIPs = serverAllowedIPs

	dnsSlice := apiClient.DNS
	if dnsSlice == nil {
		dnsSlice = []string{}
	}
	dns, d := types.ListValueFrom(ctx, types.StringType, dnsSlice)
	diags.Append(d...)
	state.DNS = dns
}
