package datasources

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/lmammino/terraform-provider-pdnd/internal/client/api"
	"github.com/lmammino/terraform-provider-pdnd/internal/providerdata"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
)

var _ datasource.DataSource = &consumerDelegationsDataSource{}

type consumerDelegationsDataSource struct {
	client api.DelegationsAPI
}

type consumerDelegationsDataSourceModel struct {
	States       types.List            `tfsdk:"states"`
	DelegatorIDs types.List            `tfsdk:"delegator_ids"`
	DelegateIDs  types.List            `tfsdk:"delegate_ids"`
	EServiceIDs  types.List            `tfsdk:"eservice_ids"`
	Delegations  []delegationItemModel `tfsdk:"delegations"`
}

type delegationItemModel struct {
	ID              types.String `tfsdk:"id"`
	EServiceID      types.String `tfsdk:"eservice_id"`
	DelegateID      types.String `tfsdk:"delegate_id"`
	DelegatorID     types.String `tfsdk:"delegator_id"`
	State           types.String `tfsdk:"state"`
	CreatedAt       types.String `tfsdk:"created_at"`
	SubmittedAt     types.String `tfsdk:"submitted_at"`
	UpdatedAt       types.String `tfsdk:"updated_at"`
	ActivatedAt     types.String `tfsdk:"activated_at"`
	RejectedAt      types.String `tfsdk:"rejected_at"`
	RevokedAt       types.String `tfsdk:"revoked_at"`
	RejectionReason types.String `tfsdk:"rejection_reason"`
}

func NewConsumerDelegationsDataSource() datasource.DataSource {
	return &consumerDelegationsDataSource{}
}

func (d *consumerDelegationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_consumer_delegations"
}

func delegationNestedSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id":               schema.StringAttribute{Computed: true},
		"eservice_id":      schema.StringAttribute{Computed: true},
		"delegate_id":      schema.StringAttribute{Computed: true},
		"delegator_id":     schema.StringAttribute{Computed: true},
		"state":            schema.StringAttribute{Computed: true},
		"created_at":       schema.StringAttribute{Computed: true},
		"submitted_at":     schema.StringAttribute{Computed: true},
		"updated_at":       schema.StringAttribute{Computed: true},
		"activated_at":     schema.StringAttribute{Computed: true},
		"rejected_at":      schema.StringAttribute{Computed: true},
		"revoked_at":       schema.StringAttribute{Computed: true},
		"rejection_reason": schema.StringAttribute{Computed: true},
	}
}

func delegationsFilterSchema(delegationType string) schema.Schema {
	return schema.Schema{
		Description: fmt.Sprintf("Lists PDND %s delegations with optional filters.", delegationType),
		Attributes: map[string]schema.Attribute{
			"states": schema.ListAttribute{
				Description: "Filter by delegation states",
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(
						stringvalidator.OneOf("WAITING_FOR_APPROVAL", "ACTIVE", "REJECTED", "REVOKED"),
					),
				},
			},
			"delegator_ids": schema.ListAttribute{
				Description: "Filter by delegator UUIDs",
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(
						stringvalidator.LengthAtLeast(1),
					),
				},
			},
			"delegate_ids": schema.ListAttribute{
				Description: "Filter by delegate UUIDs",
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(
						stringvalidator.LengthAtLeast(1),
					),
				},
			},
			"eservice_ids": schema.ListAttribute{
				Description: "Filter by e-service UUIDs",
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(
						stringvalidator.LengthAtLeast(1),
					),
				},
			},
			"delegations": schema.ListNestedAttribute{
				Description: "List of delegations",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: delegationNestedSchema(),
				},
			},
		},
	}
}

func (d *consumerDelegationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = delegationsFilterSchema("consumer")
}

func (d *consumerDelegationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *consumerDelegationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data consumerDelegationsDataSourceModel
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
		page, err := d.client.ListDelegations(ctx, "consumer", params)
		if err != nil {
			resp.Diagnostics.AddError("Error listing consumer delegations", err.Error())
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

func delegationToItemModel(d *api.Delegation) delegationItemModel {
	m := delegationItemModel{
		ID:          types.StringValue(d.ID.String()),
		EServiceID:  types.StringValue(d.EServiceID.String()),
		DelegateID:  types.StringValue(d.DelegateID.String()),
		DelegatorID: types.StringValue(d.DelegatorID.String()),
		State:       types.StringValue(d.State),
		CreatedAt:   types.StringValue(d.CreatedAt.Format(time.RFC3339)),
		SubmittedAt: types.StringValue(d.SubmittedAt.Format(time.RFC3339)),
	}

	if d.UpdatedAt != nil {
		m.UpdatedAt = types.StringValue(d.UpdatedAt.Format(time.RFC3339))
	} else {
		m.UpdatedAt = types.StringNull()
	}

	if d.ActivatedAt != nil {
		m.ActivatedAt = types.StringValue(d.ActivatedAt.Format(time.RFC3339))
	} else {
		m.ActivatedAt = types.StringNull()
	}

	if d.RejectedAt != nil {
		m.RejectedAt = types.StringValue(d.RejectedAt.Format(time.RFC3339))
	} else {
		m.RejectedAt = types.StringNull()
	}

	if d.RevokedAt != nil {
		m.RevokedAt = types.StringValue(d.RevokedAt.Format(time.RFC3339))
	} else {
		m.RevokedAt = types.StringNull()
	}

	if d.RejectionReason != nil {
		m.RejectionReason = types.StringValue(*d.RejectionReason)
	} else {
		m.RejectionReason = types.StringNull()
	}

	return m
}
