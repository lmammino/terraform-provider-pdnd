package datasources

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/lmammino/terraform-provider-pdnd/internal/client/api"
	"github.com/lmammino/terraform-provider-pdnd/internal/models"
	"github.com/lmammino/terraform-provider-pdnd/internal/providerdata"
)

var _ datasource.DataSource = &consumerDelegationDataSource{}

type consumerDelegationDataSource struct {
	client api.DelegationsAPI
}

func NewConsumerDelegationDataSource() datasource.DataSource {
	return &consumerDelegationDataSource{}
}

func (d *consumerDelegationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_consumer_delegation"
}

func (d *consumerDelegationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single PDND consumer delegation by ID.",
		Attributes: map[string]schema.Attribute{
			"id":               schema.StringAttribute{Description: "Delegation UUID", Required: true},
			"eservice_id":      schema.StringAttribute{Description: "E-Service UUID", Computed: true},
			"delegate_id":      schema.StringAttribute{Description: "Delegate UUID", Computed: true},
			"delegator_id":     schema.StringAttribute{Description: "Delegator UUID", Computed: true},
			"state":            schema.StringAttribute{Description: "Delegation state", Computed: true},
			"created_at":       schema.StringAttribute{Description: "Creation timestamp", Computed: true},
			"submitted_at":     schema.StringAttribute{Description: "Submission timestamp", Computed: true},
			"updated_at":       schema.StringAttribute{Description: "Last update timestamp", Computed: true},
			"activated_at":     schema.StringAttribute{Description: "Activation timestamp", Computed: true},
			"rejected_at":      schema.StringAttribute{Description: "Rejection timestamp", Computed: true},
			"revoked_at":       schema.StringAttribute{Description: "Revocation timestamp", Computed: true},
			"rejection_reason": schema.StringAttribute{Description: "Reason for rejection", Computed: true},
		},
	}
}

func (d *consumerDelegationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = pd.DelegationsAPI
}

func (d *consumerDelegationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.DelegationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Could not parse delegation ID as UUID: %s", err))
		return
	}

	delegation, err := d.client.GetDelegation(ctx, "consumer", id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading consumer delegation", err.Error())
		return
	}

	populateDelegationDataSourceModel(&data, delegation)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
