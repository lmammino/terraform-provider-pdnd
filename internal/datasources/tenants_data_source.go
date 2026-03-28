package datasources

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/lmammino/terraform-provider-pdnd/internal/client/api"
	"github.com/lmammino/terraform-provider-pdnd/internal/providerdata"
)

var _ datasource.DataSource = &tenantsDataSource{}

type tenantsDataSource struct {
	client api.TenantsAPI
}

type tenantsDataSourceModel struct {
	IPACode types.String      `tfsdk:"ipa_code"`
	TaxCode types.String      `tfsdk:"tax_code"`
	Tenants []tenantItemModel `tfsdk:"tenants"`
}

type tenantItemModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Kind        types.String `tfsdk:"kind"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
	OnboardedAt types.String `tfsdk:"onboarded_at"`
}

func NewTenantsDataSource() datasource.DataSource {
	return &tenantsDataSource{}
}

func (d *tenantsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenants"
}

func tenantNestedSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id":           schema.StringAttribute{Computed: true},
		"name":         schema.StringAttribute{Computed: true},
		"kind":         schema.StringAttribute{Computed: true},
		"created_at":   schema.StringAttribute{Computed: true},
		"updated_at":   schema.StringAttribute{Computed: true},
		"onboarded_at": schema.StringAttribute{Computed: true},
	}
}

func (d *tenantsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists PDND tenants with optional filters.",
		Attributes: map[string]schema.Attribute{
			"ipa_code": schema.StringAttribute{
				Description: "Filter by IPA code",
				Optional:    true,
			},
			"tax_code": schema.StringAttribute{
				Description: "Filter by tax code",
				Optional:    true,
			},
			"tenants": schema.ListNestedAttribute{
				Description: "List of tenants",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: tenantNestedSchema(),
				},
			},
		},
	}
}

func (d *tenantsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = pd.TenantsAPI
}

func (d *tenantsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data tenantsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := api.ListTenantsParams{
		Offset: 0,
		Limit:  50,
	}

	if !data.IPACode.IsNull() && !data.IPACode.IsUnknown() {
		code := data.IPACode.ValueString()
		params.IPACode = &code
	}

	if !data.TaxCode.IsNull() && !data.TaxCode.IsUnknown() {
		code := data.TaxCode.ValueString()
		params.TaxCode = &code
	}

	// Auto-paginate
	var allTenants []api.TenantInfo
	for {
		page, err := d.client.ListTenants(ctx, params)
		if err != nil {
			resp.Diagnostics.AddError("Error listing tenants", err.Error())
			return
		}

		allTenants = append(allTenants, page.Results...)

		if int32(len(allTenants)) >= page.Pagination.TotalCount {
			break
		}
		params.Offset += params.Limit
	}

	data.Tenants = make([]tenantItemModel, len(allTenants))
	for i, t := range allTenants {
		data.Tenants[i] = tenantToItemModel(&t)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func tenantToItemModel(t *api.TenantInfo) tenantItemModel {
	m := tenantItemModel{
		ID:        types.StringValue(t.ID.String()),
		Name:      types.StringValue(t.Name),
		CreatedAt: types.StringValue(t.CreatedAt.Format(time.RFC3339)),
	}

	if t.Kind != nil {
		m.Kind = types.StringValue(*t.Kind)
	} else {
		m.Kind = types.StringNull()
	}

	if t.UpdatedAt != nil {
		m.UpdatedAt = types.StringValue(t.UpdatedAt.Format(time.RFC3339))
	} else {
		m.UpdatedAt = types.StringNull()
	}

	if t.OnboardedAt != nil {
		m.OnboardedAt = types.StringValue(t.OnboardedAt.Format(time.RFC3339))
	} else {
		m.OnboardedAt = types.StringNull()
	}

	return m
}
