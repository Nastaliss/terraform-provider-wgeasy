// Package resourceclient implements the wgeasy_client resource for the Terraform provider.
package resourceclient

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

// NewClientResource creates a new wgeasy_client resource instance.
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
		updateReq := buildUpdateRequest(ctx, plan, current)
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

	updateReq := buildUpdateRequest(ctx, plan, current)

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
	if hasNonEmptyList(plan.AllowedIPs) || hasNonEmptyList(plan.ServerAllowedIPs) || hasNonEmptyList(plan.DNS) {
		return true
	}
	if isSetInt64(plan.MTU) || isSetInt64(plan.PersistentKeepalive) || isSetString(plan.ServerEndpoint) {
		return true
	}
	if !plan.Enabled.IsNull() && !plan.Enabled.IsUnknown() && !plan.Enabled.ValueBool() {
		return true
	}
	return false
}

func hasNonEmptyList(list types.List) bool {
	return !list.IsNull() && !list.IsUnknown() && len(list.Elements()) > 0
}

func isSetInt64(val types.Int64) bool {
	return !val.IsNull() && !val.IsUnknown()
}

func isSetString(val types.String) bool {
	return !val.IsNull() && !val.IsUnknown()
}

// buildUpdateRequest builds an update request starting from the current API state,
// then overlays any values from the plan.
func buildUpdateRequest(ctx context.Context, plan clientResourceModel, current *client.Client) client.UpdateClientRequest {
	req := initUpdateRequestFromCurrent(current)
	applyPlanToUpdateRequest(ctx, plan, &req)
	return req
}

func initUpdateRequestFromCurrent(current *client.Client) client.UpdateClientRequest {
	return client.UpdateClientRequest{
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
		ServerEndpoint:      current.ServerEndpoint,
		ExpiresAt:           current.ExpiresAt,
		I1:                  current.I1,
		I2:                  current.I2,
		I3:                  current.I3,
		I4:                  current.I4,
		I5:                  current.I5,
	}
}

func applyPlanToUpdateRequest(ctx context.Context, plan clientResourceModel, req *client.UpdateClientRequest) {
	req.Name = plan.Name.ValueString()
	req.Enabled = plan.Enabled.ValueBool()

	if isSetString(plan.ExpiresAt) {
		v := plan.ExpiresAt.ValueString()
		req.ExpiresAt = &v
	}

	applyListField(ctx, plan.AllowedIPs, &req.AllowedIPs)
	applyListField(ctx, plan.ServerAllowedIPs, &req.ServerAllowedIPs)
	applyDNSField(ctx, plan.DNS, &req.DNS)

	applyInt64Field(plan.MTU, &req.MTU)
	applyInt64Field(plan.PersistentKeepalive, &req.PersistentKeepalive)
	applyStringPtrField(plan.ServerEndpoint, &req.ServerEndpoint)
	applyStringField(plan.PreUp, &req.PreUp)
	applyStringField(plan.PostUp, &req.PostUp)
	applyStringField(plan.PreDown, &req.PreDown)
	applyStringField(plan.PostDown, &req.PostDown)
	applyInt64Field(plan.JC, &req.JC)
	applyInt64Field(plan.JMin, &req.JMin)
	applyInt64Field(plan.JMax, &req.JMax)
}

func applyListField(ctx context.Context, list types.List, target *[]string) {
	if list.IsNull() || list.IsUnknown() {
		return
	}
	var values []string
	list.ElementsAs(ctx, &values, false)
	if len(values) > 0 {
		*target = values
	} else {
		// Empty list means "use global config" - send null to API.
		*target = nil
	}
}

func applyDNSField(ctx context.Context, list types.List, target *[]string) {
	if list.IsNull() || list.IsUnknown() {
		return
	}
	var values []string
	list.ElementsAs(ctx, &values, false)
	*target = values
}

func applyInt64Field(val types.Int64, target *int64) {
	if isSetInt64(val) {
		*target = val.ValueInt64()
	}
}

func applyStringField(val types.String, target *string) {
	if isSetString(val) {
		*target = val.ValueString()
	}
}

func applyStringPtrField(val types.String, target **string) {
	if isSetString(val) {
		v := val.ValueString()
		*target = &v
	}
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

	state.AllowedIPs = sliceToList(ctx, apiClient.AllowedIPs, diags)
	state.ServerAllowedIPs = sliceToList(ctx, apiClient.ServerAllowedIPs, diags)
	state.DNS = sliceToList(ctx, apiClient.DNS, diags)
}

func sliceToList(ctx context.Context, slice []string, diags *diag.Diagnostics) types.List {
	if slice == nil {
		slice = []string{}
	}
	list, d := types.ListValueFrom(ctx, types.StringType, slice)
	diags.Append(d...)
	return list
}
