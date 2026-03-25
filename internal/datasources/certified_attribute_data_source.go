package datasources

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/lmammino/terraform-provider-pdnd/internal/client/api"
	"github.com/lmammino/terraform-provider-pdnd/internal/providerdata"
)

var _ datasource.DataSource = &certifiedAttributeDataSource{}

type certifiedAttributeDataSource struct {
	client api.AttributesAPI
}

type certifiedAttributeDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Code        types.String `tfsdk:"code"`
	Origin      types.String `tfsdk:"origin"`
	CreatedAt   types.String `tfsdk:"created_at"`
}

func NewCertifiedAttributeDataSource() datasource.DataSource {
	return &certifiedAttributeDataSource{}
}

func (d *certifiedAttributeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certified_attribute"
}

func (d *certifiedAttributeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single PDND certified attribute by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Certified attribute UUID",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Attribute name",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Attribute description",
				Computed:    true,
			},
			"code": schema.StringAttribute{
				Description: "Unique identifier on source registry (e.g. IPA code)",
				Computed:    true,
			},
			"origin": schema.StringAttribute{
				Description: "Source registry (e.g. IPA)",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Creation timestamp",
				Computed:    true,
			},
		},
	}
}

func (d *certifiedAttributeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = pd.AttributesAPI
}

func (d *certifiedAttributeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data certifiedAttributeDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Could not parse certified attribute ID as UUID: %s", err))
		return
	}

	attr, err := d.client.GetCertifiedAttribute(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading certified attribute", err.Error())
		return
	}

	data.ID = types.StringValue(attr.ID.String())
	data.Name = types.StringValue(attr.Name)
	data.Description = types.StringValue(attr.Description)
	data.Code = types.StringValue(attr.Code)
	data.Origin = types.StringValue(attr.Origin)
	data.CreatedAt = types.StringValue(attr.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
