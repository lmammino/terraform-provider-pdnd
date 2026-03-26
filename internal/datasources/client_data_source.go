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
	"github.com/lmammino/terraform-provider-pdnd/internal/models"
	"github.com/lmammino/terraform-provider-pdnd/internal/providerdata"
)

var _ datasource.DataSource = &clientDataSource{}

type clientDataSource struct {
	client api.ClientsAPI
}

func NewClientDataSource() datasource.DataSource {
	return &clientDataSource{}
}

func (d *clientDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_client"
}

func (d *clientDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single PDND client by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Client UUID",
				Required:    true,
			},
			"consumer_id": schema.StringAttribute{
				Description: "Consumer UUID",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Client name",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Client description",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Creation timestamp",
				Computed:    true,
			},
		},
	}
}

func (d *clientDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *clientDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.ClientDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Could not parse client ID as UUID: %s", err))
		return
	}

	clientInfo, err := d.client.GetClient(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading client", err.Error())
		return
	}

	data.ID = types.StringValue(clientInfo.ID.String())
	data.ConsumerID = types.StringValue(clientInfo.ConsumerID.String())
	data.Name = types.StringValue(clientInfo.Name)
	data.CreatedAt = types.StringValue(clientInfo.CreatedAt.Format(time.RFC3339))

	if clientInfo.Description != nil {
		data.Description = types.StringValue(*clientInfo.Description)
	} else {
		data.Description = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
