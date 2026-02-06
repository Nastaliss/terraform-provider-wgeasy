package datasource_client

import (
	"context"
	"fmt"

	"github.com/Nastaliss/terraform-provider-wgeasy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var _ datasource.DataSource = &clientsDataSource{}

type clientsDataSource struct {
	apiClient *client.WGEasyClient
}

type clientsDataSourceModel struct {
	Clients []clientModel `tfsdk:"clients"`
}

func NewClientsDataSource() datasource.DataSource {
	return &clientsDataSource{}
}

func (d *clientsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_clients"
}

func (d *clientsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all WireGuard clients/peers from a wg-easy instance.",
		Attributes: map[string]schema.Attribute{
			"clients": schema.ListNestedAttribute{
				Description: "List of all WireGuard clients.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: clientDataSourceAttributes(false),
				},
			},
		},
	}
}

func (d *clientsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *clientsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	apiClients, err := d.apiClient.GetClients()
	if err != nil {
		resp.Diagnostics.AddError("Error reading clients", err.Error())
		return
	}

	var state clientsDataSourceModel
	state.Clients = make([]clientModel, len(apiClients))

	for i, apiClient := range apiClients {
		mapClientToModel(ctx, &apiClient, &state.Clients[i], &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
