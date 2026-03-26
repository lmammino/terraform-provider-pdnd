package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/lmammino/terraform-provider-pdnd/internal/client/api"
	"github.com/lmammino/terraform-provider-pdnd/internal/providerdata"
)

var _ datasource.DataSource = &producerDelegationsDataSource{}

type producerDelegationsDataSource struct {
	client api.DelegationsAPI
}

type producerDelegationsDataSourceModel struct {
	States       types.List            `tfsdk:"states"`
	DelegatorIDs types.List            `tfsdk:"delegator_ids"`
	DelegateIDs  types.List            `tfsdk:"delegate_ids"`
	EServiceIDs  types.List            `tfsdk:"eservice_ids"`
	Delegations  []delegationItemModel `tfsdk:"delegations"`
}

func NewProducerDelegationsDataSource() datasource.DataSource {
	return &producerDelegationsDataSource{}
}

func (d *producerDelegationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_producer_delegations"
}

func (d *producerDelegationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = delegationsFilterSchema("producer")
}

func (d *producerDelegationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *producerDelegationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data producerDelegationsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := api.ListDelegationsParams{
		Offset: 0,
		Limit:  50,
	}

	if !data.States.IsNull() && !data.States.IsUnknown() {
		var states []string
		resp.Diagnostics.Append(data.States.ElementsAs(ctx, &states, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		params.States = states
	}

	if !data.DelegatorIDs.IsNull() && !data.DelegatorIDs.IsUnknown() {
		var ids []string
		resp.Diagnostics.Append(data.DelegatorIDs.ElementsAs(ctx, &ids, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		params.DelegatorIDs = parseUUIDs(ids)
	}

	if !data.DelegateIDs.IsNull() && !data.DelegateIDs.IsUnknown() {
		var ids []string
		resp.Diagnostics.Append(data.DelegateIDs.ElementsAs(ctx, &ids, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		params.DelegateIDs = parseUUIDs(ids)
	}

	if !data.EServiceIDs.IsNull() && !data.EServiceIDs.IsUnknown() {
		var ids []string
		resp.Diagnostics.Append(data.EServiceIDs.ElementsAs(ctx, &ids, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		params.EServiceIDs = parseUUIDs(ids)
	}

	// Auto-paginate
	var allDelegations []api.Delegation
	for {
		page, err := d.client.ListDelegations(ctx, "producer", params)
		if err != nil {
			resp.Diagnostics.AddError("Error listing producer delegations", err.Error())
			return
		}

		allDelegations = append(allDelegations, page.Results...)

		if int32(len(allDelegations)) >= page.Pagination.TotalCount {
			break
		}
		params.Offset += params.Limit
	}

	data.Delegations = make([]delegationItemModel, len(allDelegations))
	for i, del := range allDelegations {
		data.Delegations[i] = delegationToItemModel(&del)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
