package datasources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/lmammino/terraform-provider-pdnd/internal/client/api"
	"github.com/lmammino/terraform-provider-pdnd/internal/providerdata"
)

var _ datasource.DataSource = &certifiedAttributesDataSource{}

type certifiedAttributesDataSource struct {
	client api.AttributesAPI
}

type certifiedAttributesDataSourceModel struct {
	Name       types.String                        `tfsdk:"name"`
	Attributes []certifiedAttributeNestedModel     `tfsdk:"attributes"`
}

type certifiedAttributeNestedModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Code        types.String `tfsdk:"code"`
	Origin      types.String `tfsdk:"origin"`
	CreatedAt   types.String `tfsdk:"created_at"`
}

func NewCertifiedAttributesDataSource() datasource.DataSource {
	return &certifiedAttributesDataSource{}
}

func (d *certifiedAttributesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certified_attributes"
}

func (d *certifiedAttributesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists PDND certified attributes with optional name filter.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Filter by attribute name (case-insensitive contains match)",
				Optional:    true,
			},
			"attributes": schema.ListNestedAttribute{
				Description: "List of certified attributes",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":          schema.StringAttribute{Computed: true},
						"name":        schema.StringAttribute{Computed: true},
						"description": schema.StringAttribute{Computed: true},
						"code":        schema.StringAttribute{Computed: true},
						"origin":      schema.StringAttribute{Computed: true},
						"created_at":  schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *certifiedAttributesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *certifiedAttributesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data certifiedAttributesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := api.PaginationParams{
		Offset: 0,
		Limit:  50,
	}

	// Auto-paginate
	var allAttrs []api.CertifiedAttribute
	for {
		page, err := d.client.ListCertifiedAttributes(ctx, params)
		if err != nil {
			resp.Diagnostics.AddError("Error listing certified attributes", err.Error())
			return
		}

		allAttrs = append(allAttrs, page.Results...)

		if int32(len(allAttrs)) >= page.Pagination.TotalCount {
			break
		}
		params.Offset += params.Limit
	}

	// Client-side name filter
	var nameFilter string
	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		nameFilter = data.Name.ValueString()
	}

	var filtered []api.CertifiedAttribute
	for _, attr := range allAttrs {
		if nameFilter != "" && !strings.Contains(strings.ToLower(attr.Name), strings.ToLower(nameFilter)) {
			continue
		}
		filtered = append(filtered, attr)
	}

	// Convert to nested models
	data.Attributes = make([]certifiedAttributeNestedModel, len(filtered))
	for i, attr := range filtered {
		data.Attributes[i] = certifiedAttributeNestedModel{
			ID:          types.StringValue(attr.ID.String()),
			Name:        types.StringValue(attr.Name),
			Description: types.StringValue(attr.Description),
			Code:        types.StringValue(attr.Code),
			Origin:      types.StringValue(attr.Origin),
			CreatedAt:   types.StringValue(attr.CreatedAt.UTC().Format("2006-01-02T15:04:05Z")),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
