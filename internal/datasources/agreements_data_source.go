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
	"github.com/lmammino/terraform-provider-pdnd/internal/providerdata"
)

var _ datasource.DataSource = &agreementsDataSource{}

type agreementsDataSource struct {
	client api.AgreementsAPI
}

type agreementsDataSourceModel struct {
	States        types.List              `tfsdk:"states"`
	ProducerIDs   types.List              `tfsdk:"producer_ids"`
	ConsumerIDs   types.List              `tfsdk:"consumer_ids"`
	EServiceIDs   types.List              `tfsdk:"eservice_ids"`
	DescriptorIDs types.List              `tfsdk:"descriptor_ids"`
	Agreements    []agreementNestedModel  `tfsdk:"agreements"`
}

type agreementNestedModel struct {
	ID                  types.String `tfsdk:"id"`
	EServiceID          types.String `tfsdk:"eservice_id"`
	DescriptorID        types.String `tfsdk:"descriptor_id"`
	ProducerID          types.String `tfsdk:"producer_id"`
	ConsumerID          types.String `tfsdk:"consumer_id"`
	DelegationID        types.String `tfsdk:"delegation_id"`
	State               types.String `tfsdk:"state"`
	SuspendedByConsumer types.Bool   `tfsdk:"suspended_by_consumer"`
	SuspendedByProducer types.Bool   `tfsdk:"suspended_by_producer"`
	SuspendedByPlatform types.Bool   `tfsdk:"suspended_by_platform"`
	ConsumerNotes       types.String `tfsdk:"consumer_notes"`
	RejectionReason     types.String `tfsdk:"rejection_reason"`
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
	SuspendedAt         types.String `tfsdk:"suspended_at"`
}

func NewAgreementsDataSource() datasource.DataSource {
	return &agreementsDataSource{}
}

func (d *agreementsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agreements"
}

func agreementNestedSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id":                    schema.StringAttribute{Computed: true},
		"eservice_id":           schema.StringAttribute{Computed: true},
		"descriptor_id":         schema.StringAttribute{Computed: true},
		"producer_id":           schema.StringAttribute{Computed: true},
		"consumer_id":           schema.StringAttribute{Computed: true},
		"delegation_id":         schema.StringAttribute{Computed: true},
		"state":                 schema.StringAttribute{Computed: true},
		"suspended_by_consumer": schema.BoolAttribute{Computed: true},
		"suspended_by_producer": schema.BoolAttribute{Computed: true},
		"suspended_by_platform": schema.BoolAttribute{Computed: true},
		"consumer_notes":        schema.StringAttribute{Computed: true},
		"rejection_reason":      schema.StringAttribute{Computed: true},
		"created_at":            schema.StringAttribute{Computed: true},
		"updated_at":            schema.StringAttribute{Computed: true},
		"suspended_at":          schema.StringAttribute{Computed: true},
	}
}

func (d *agreementsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists PDND agreements with optional filters.",
		Attributes: map[string]schema.Attribute{
			"states": schema.ListAttribute{
				Description: "Filter by agreement states",
				Optional:    true,
				ElementType: types.StringType,
			},
			"producer_ids": schema.ListAttribute{
				Description: "Filter by producer UUIDs",
				Optional:    true,
				ElementType: types.StringType,
			},
			"consumer_ids": schema.ListAttribute{
				Description: "Filter by consumer UUIDs",
				Optional:    true,
				ElementType: types.StringType,
			},
			"eservice_ids": schema.ListAttribute{
				Description: "Filter by e-service UUIDs",
				Optional:    true,
				ElementType: types.StringType,
			},
			"descriptor_ids": schema.ListAttribute{
				Description: "Filter by descriptor UUIDs",
				Optional:    true,
				ElementType: types.StringType,
			},
			"agreements": schema.ListNestedAttribute{
				Description: "List of agreements",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: agreementNestedSchema(),
				},
			},
		},
	}
}

func (d *agreementsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = pd.AgreementsAPI
}

func (d *agreementsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data agreementsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := api.ListAgreementsParams{
		Offset: 0,
		Limit:  50,
	}

	// Extract filter values
	if !data.States.IsNull() && !data.States.IsUnknown() {
		var states []string
		resp.Diagnostics.Append(data.States.ElementsAs(ctx, &states, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		params.States = states
	}

	if !data.ProducerIDs.IsNull() && !data.ProducerIDs.IsUnknown() {
		var ids []string
		resp.Diagnostics.Append(data.ProducerIDs.ElementsAs(ctx, &ids, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		params.ProducerIDs = parseUUIDs(ids)
	}

	if !data.ConsumerIDs.IsNull() && !data.ConsumerIDs.IsUnknown() {
		var ids []string
		resp.Diagnostics.Append(data.ConsumerIDs.ElementsAs(ctx, &ids, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		params.ConsumerIDs = parseUUIDs(ids)
	}

	if !data.EServiceIDs.IsNull() && !data.EServiceIDs.IsUnknown() {
		var ids []string
		resp.Diagnostics.Append(data.EServiceIDs.ElementsAs(ctx, &ids, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		params.EServiceIDs = parseUUIDs(ids)
	}

	if !data.DescriptorIDs.IsNull() && !data.DescriptorIDs.IsUnknown() {
		var ids []string
		resp.Diagnostics.Append(data.DescriptorIDs.ElementsAs(ctx, &ids, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		params.DescriptorIDs = parseUUIDs(ids)
	}

	// Auto-paginate
	var allAgreements []api.Agreement
	for {
		page, err := d.client.ListAgreements(ctx, params)
		if err != nil {
			resp.Diagnostics.AddError("Error listing agreements", err.Error())
			return
		}

		allAgreements = append(allAgreements, page.Results...)

		if int32(len(allAgreements)) >= page.Pagination.TotalCount {
			break
		}
		params.Offset += params.Limit
	}

	// Convert to nested models
	data.Agreements = make([]agreementNestedModel, len(allAgreements))
	for i, a := range allAgreements {
		data.Agreements[i] = agreementToNestedModel(&a)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func agreementToNestedModel(a *api.Agreement) agreementNestedModel {
	m := agreementNestedModel{
		ID:           types.StringValue(a.ID.String()),
		EServiceID:   types.StringValue(a.EServiceID.String()),
		DescriptorID: types.StringValue(a.DescriptorID.String()),
		ProducerID:   types.StringValue(a.ProducerID.String()),
		ConsumerID:   types.StringValue(a.ConsumerID.String()),
		State:        types.StringValue(a.State),
		CreatedAt:    types.StringValue(a.CreatedAt.Format(time.RFC3339)),
	}

	if a.DelegationID != nil {
		m.DelegationID = types.StringValue(a.DelegationID.String())
	} else {
		m.DelegationID = types.StringNull()
	}

	if a.SuspendedByConsumer != nil {
		m.SuspendedByConsumer = types.BoolValue(*a.SuspendedByConsumer)
	} else {
		m.SuspendedByConsumer = types.BoolNull()
	}

	if a.SuspendedByProducer != nil {
		m.SuspendedByProducer = types.BoolValue(*a.SuspendedByProducer)
	} else {
		m.SuspendedByProducer = types.BoolNull()
	}

	if a.SuspendedByPlatform != nil {
		m.SuspendedByPlatform = types.BoolValue(*a.SuspendedByPlatform)
	} else {
		m.SuspendedByPlatform = types.BoolNull()
	}

	if a.ConsumerNotes != nil {
		m.ConsumerNotes = types.StringValue(*a.ConsumerNotes)
	} else {
		m.ConsumerNotes = types.StringNull()
	}

	if a.RejectionReason != nil {
		m.RejectionReason = types.StringValue(*a.RejectionReason)
	} else {
		m.RejectionReason = types.StringNull()
	}

	if a.UpdatedAt != nil {
		m.UpdatedAt = types.StringValue(a.UpdatedAt.Format(time.RFC3339))
	} else {
		m.UpdatedAt = types.StringNull()
	}

	if a.SuspendedAt != nil {
		m.SuspendedAt = types.StringValue(a.SuspendedAt.Format(time.RFC3339))
	} else {
		m.SuspendedAt = types.StringNull()
	}

	return m
}

func parseUUIDs(strs []string) []uuid.UUID {
	uuids := make([]uuid.UUID, len(strs))
	for i, s := range strs {
		uuids[i] = uuid.MustParse(s)
	}
	return uuids
}
