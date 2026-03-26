package datasources

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/lmammino/terraform-provider-pdnd/internal/client/api"
	"github.com/lmammino/terraform-provider-pdnd/internal/providerdata"
)

var _ datasource.DataSource = &clientsDataSource{}

type clientsDataSource struct {
	client api.ClientsAPI
}

type clientsDataSourceModel struct {
	Name       types.String      `tfsdk:"name"`
	ConsumerID types.String      `tfsdk:"consumer_id"`
	Clients    []clientItemModel `tfsdk:"clients"`
}

type clientItemModel struct {
	ID          types.String `tfsdk:"id"`
	ConsumerID  types.String `tfsdk:"consumer_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	CreatedAt   types.String `tfsdk:"created_at"`
}

func NewClientsDataSource() datasource.DataSource {
	return &clientsDataSource{}
}

func (d *clientsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_clients"
}

func clientNestedSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id":          schema.StringAttribute{Computed: true},
		"consumer_id": schema.StringAttribute{Computed: true},
		"name":        schema.StringAttribute{Computed: true},
		"description": schema.StringAttribute{Computed: true},
		"created_at":  schema.StringAttribute{Computed: true},
	}
}

func (d *clientsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists PDND clients with optional filters.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Filter by client name",
				Optional:    true,
			},
			"consumer_id": schema.StringAttribute{
				Description: "Filter by consumer UUID",
				Optional:    true,
			},
			"clients": schema.ListNestedAttribute{
				Description: "List of clients",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: clientNestedSchema(),
				},
			},
		},
	}
}

func (d *clientsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	pd, ok := req.ProviderData.(*providerdata.ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *providerdata.ProviderData, got: %T", req.ProviderData),
		)
		return
	}

	d.client = pd.ClientsAPI
}

func (d *clientsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data clientsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := api.ListClientsParams{
		Offset: 0,
		Limit:  50,
	}

	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		name := data.Name.ValueString()
		params.Name = &name
	}

	if !data.ConsumerID.IsNull() && !data.ConsumerID.IsUnknown() {
		id := uuid.MustParse(data.ConsumerID.ValueString())
		params.ConsumerID = &id
	}

	// Auto-paginate
	var allClients []api.ClientInfo
	for {
		page, err := d.client.ListClients(ctx, params)
		if err != nil {
			resp.Diagnostics.AddError("Error listing clients", err.Error())
			return
		}

		allClients = append(allClients, page.Results...)

		if int32(len(allClients)) >= page.Pagination.TotalCount {
			break
		}
		params.Offset += params.Limit
	}

	data.Clients = make([]clientItemModel, len(allClients))
	for i, c := range allClients {
		data.Clients[i] = clientToItemModel(&c)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func clientToItemModel(c *api.ClientInfo) clientItemModel {
	m := clientItemModel{
		ID:         types.StringValue(c.ID.String()),
		ConsumerID: types.StringValue(c.ConsumerID.String()),
		Name:       types.StringValue(c.Name),
		CreatedAt:  types.StringValue(c.CreatedAt.Format(time.RFC3339)),
	}

	if c.Description != nil {
		m.Description = types.StringValue(*c.Description)
	} else {
		m.Description = types.StringNull()
	}

	return m
}
