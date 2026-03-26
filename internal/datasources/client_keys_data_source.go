package datasources

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/lmammino/terraform-provider-pdnd/internal/client/api"
	"github.com/lmammino/terraform-provider-pdnd/internal/models"
	"github.com/lmammino/terraform-provider-pdnd/internal/providerdata"
)

var _ datasource.DataSource = &clientKeysDataSource{}

type clientKeysDataSource struct {
	client api.ClientsAPI
}

type clientKeysDataSourceModel struct {
	ClientID types.String                   `tfsdk:"client_id"`
	Keys     []models.ClientKeyDataSourceModel `tfsdk:"keys"`
}

func NewClientKeysDataSource() datasource.DataSource {
	return &clientKeysDataSource{}
}

func (d *clientKeysDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_client_keys"
}

func clientKeyNestedSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"kid": schema.StringAttribute{Computed: true},
		"kty": schema.StringAttribute{Computed: true},
		"alg": schema.StringAttribute{Computed: true},
		"use": schema.StringAttribute{Computed: true},
	}
}

func (d *clientKeysDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists public keys for a PDND client.",
		Attributes: map[string]schema.Attribute{
			"client_id": schema.StringAttribute{
				Description: "Client UUID",
				Required:    true,
			},
			"keys": schema.ListNestedAttribute{
				Description: "List of client keys",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: clientKeyNestedSchema(),
				},
			},
		},
	}
}

func (d *clientKeysDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *clientKeysDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data clientKeysDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clientID, err := uuid.Parse(data.ClientID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid client_id", fmt.Sprintf("Could not parse client_id as UUID: %s", err))
		return
	}

	// Auto-paginate
	var allKeys []api.ClientKey
	var offset int32
	const limit int32 = 50

	for {
		page, err := d.client.ListClientKeys(ctx, clientID, offset, limit)
		if err != nil {
			resp.Diagnostics.AddError("Error listing client keys", err.Error())
			return
		}

		allKeys = append(allKeys, page.Results...)

		if int32(len(allKeys)) >= page.Pagination.TotalCount {
			break
		}
		offset += limit
	}

	data.Keys = make([]models.ClientKeyDataSourceModel, len(allKeys))
	for i, k := range allKeys {
		data.Keys[i] = models.ClientKeyDataSourceModel{
			Kid: types.StringValue(k.Kid),
			Kty: types.StringValue(k.Kty),
		}
		if k.Alg != nil {
			data.Keys[i].Alg = types.StringValue(*k.Alg)
		} else {
			data.Keys[i].Alg = types.StringNull()
		}
		if k.Use != nil {
			data.Keys[i].Use = types.StringValue(*k.Use)
		} else {
			data.Keys[i].Use = types.StringNull()
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
