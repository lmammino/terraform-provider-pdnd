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

var _ datasource.DataSource = &tenantDataSource{}

type tenantDataSource struct {
	client api.TenantsAPI
}

func NewTenantDataSource() datasource.DataSource {
	return &tenantDataSource{}
}

func (d *tenantDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant"
}

func (d *tenantDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single PDND tenant by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Tenant UUID",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Tenant name",
				Computed:    true,
			},
			"kind": schema.StringAttribute{
				Description: "Tenant kind",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Creation timestamp",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Last update timestamp",
				Computed:    true,
			},
			"onboarded_at": schema.StringAttribute{
				Description: "Onboarding timestamp",
				Computed:    true,
			},
		},
	}
}

func (d *tenantDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *tenantDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.TenantDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Could not parse tenant ID as UUID: %s", err))
		return
	}

	tenant, err := d.client.GetTenant(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading tenant", err.Error())
		return
	}

	data.ID = types.StringValue(tenant.ID.String())
	data.Name = types.StringValue(tenant.Name)
	data.CreatedAt = types.StringValue(tenant.CreatedAt.Format(time.RFC3339))

	if tenant.Kind != nil {
		data.Kind = types.StringValue(*tenant.Kind)
	} else {
		data.Kind = types.StringNull()
	}

	if tenant.UpdatedAt != nil {
		data.UpdatedAt = types.StringValue(tenant.UpdatedAt.Format(time.RFC3339))
	} else {
		data.UpdatedAt = types.StringNull()
	}

	if tenant.OnboardedAt != nil {
		data.OnboardedAt = types.StringValue(tenant.OnboardedAt.Format(time.RFC3339))
	} else {
		data.OnboardedAt = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
