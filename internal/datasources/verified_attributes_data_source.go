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

var _ datasource.DataSource = &verifiedAttributesDataSource{}

type verifiedAttributesDataSource struct {
	client api.AttributesAPI
}

type verifiedAttributesDataSourceModel struct {
	Name       types.String                    `tfsdk:"name"`
	Attributes []verifiedAttributeNestedModel  `tfsdk:"attributes"`
}

type verifiedAttributeNestedModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	CreatedAt   types.String `tfsdk:"created_at"`
}

func NewVerifiedAttributesDataSource() datasource.DataSource {
	return &verifiedAttributesDataSource{}
}

func (d *verifiedAttributesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_verified_attributes"
}

func (d *verifiedAttributesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists PDND verified attributes with optional name filter.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Filter by attribute name (case-insensitive contains match)",
				Optional:    true,
			},
			"attributes": schema.ListNestedAttribute{
				Description: "List of verified attributes",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":          schema.StringAttribute{Computed: true},
						"name":        schema.StringAttribute{Computed: true},
						"description": schema.StringAttribute{Computed: true},
						"created_at":  schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *verifiedAttributesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *verifiedAttributesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data verifiedAttributesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := api.PaginationParams{
		Offset: 0,
		Limit:  50,
	}

	// Auto-paginate
	var allAttrs []api.VerifiedAttribute
	for {
		page, err := d.client.ListVerifiedAttributes(ctx, params)
		if err != nil {
			resp.Diagnostics.AddError("Error listing verified attributes", err.Error())
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

	var filtered []api.VerifiedAttribute
	for _, attr := range allAttrs {
		if nameFilter != "" && !strings.Contains(strings.ToLower(attr.Name), strings.ToLower(nameFilter)) {
			continue
		}
		filtered = append(filtered, attr)
	}

	// Convert to nested models
	data.Attributes = make([]verifiedAttributeNestedModel, len(filtered))
	for i, attr := range filtered {
		data.Attributes[i] = verifiedAttributeNestedModel{
			ID:          types.StringValue(attr.ID.String()),
			Name:        types.StringValue(attr.Name),
			Description: types.StringValue(attr.Description),
			CreatedAt:   types.StringValue(attr.CreatedAt.UTC().Format("2006-01-02T15:04:05Z")),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
