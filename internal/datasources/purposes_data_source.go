package datasources

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/lmammino/terraform-provider-pdnd/internal/client/api"
	"github.com/lmammino/terraform-provider-pdnd/internal/providerdata"
)

var _ datasource.DataSource = &purposesDataSource{}

type purposesDataSource struct {
	client api.PurposesAPI
}

type purposesDataSourceModel struct {
	EServiceIDs types.List          `tfsdk:"eservice_ids"`
	Title       types.String        `tfsdk:"title"`
	ConsumerIDs types.List          `tfsdk:"consumer_ids"`
	States      types.List          `tfsdk:"states"`
	Purposes    []purposeItemModel  `tfsdk:"purposes"`
}

type purposeItemModel struct {
	ID                  types.String `tfsdk:"id"`
	EServiceID          types.String `tfsdk:"eservice_id"`
	ConsumerID          types.String `tfsdk:"consumer_id"`
	Title               types.String `tfsdk:"title"`
	Description         types.String `tfsdk:"description"`
	DailyCalls          types.Int64  `tfsdk:"daily_calls"`
	State               types.String `tfsdk:"state"`
	IsFreeOfCharge      types.Bool   `tfsdk:"is_free_of_charge"`
	FreeOfChargeReason  types.String `tfsdk:"free_of_charge_reason"`
	IsRiskAnalysisValid types.Bool   `tfsdk:"is_risk_analysis_valid"`
	SuspendedByConsumer types.Bool   `tfsdk:"suspended_by_consumer"`
	SuspendedByProducer types.Bool   `tfsdk:"suspended_by_producer"`
	DelegationID        types.String `tfsdk:"delegation_id"`
	VersionID           types.String `tfsdk:"version_id"`
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
}

func NewPurposesDataSource() datasource.DataSource {
	return &purposesDataSource{}
}

func (d *purposesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_purposes"
}

func purposeNestedSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id":                     schema.StringAttribute{Computed: true},
		"eservice_id":            schema.StringAttribute{Computed: true},
		"consumer_id":            schema.StringAttribute{Computed: true},
		"title":                  schema.StringAttribute{Computed: true},
		"description":            schema.StringAttribute{Computed: true},
		"daily_calls":            schema.Int64Attribute{Computed: true},
		"state":                  schema.StringAttribute{Computed: true},
		"is_free_of_charge":      schema.BoolAttribute{Computed: true},
		"free_of_charge_reason":  schema.StringAttribute{Computed: true},
		"is_risk_analysis_valid": schema.BoolAttribute{Computed: true},
		"suspended_by_consumer":  schema.BoolAttribute{Computed: true},
		"suspended_by_producer":  schema.BoolAttribute{Computed: true},
		"delegation_id":          schema.StringAttribute{Computed: true},
		"version_id":             schema.StringAttribute{Computed: true},
		"created_at":             schema.StringAttribute{Computed: true},
		"updated_at":             schema.StringAttribute{Computed: true},
	}
}

func (d *purposesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists PDND purposes with optional filters.",
		Attributes: map[string]schema.Attribute{
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
			"title": schema.StringAttribute{
				Description: "Filter by purpose title",
				Optional:    true,
			},
			"consumer_ids": schema.ListAttribute{
				Description: "Filter by consumer UUIDs",
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(
						stringvalidator.LengthAtLeast(1),
					),
				},
			},
			"states": schema.ListAttribute{
				Description: "Filter by purpose states",
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(
						stringvalidator.OneOf("DRAFT", "ACTIVE", "SUSPENDED", "ARCHIVED", "WAITING_FOR_APPROVAL", "REJECTED"),
					),
				},
			},
			"purposes": schema.ListNestedAttribute{
				Description: "List of purposes",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: purposeNestedSchema(),
				},
			},
		},
	}
}

func (d *purposesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = pd.PurposesAPI
}

func (d *purposesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data purposesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := api.ListPurposesParams{
		Offset: 0,
		Limit:  50,
	}

	// Extract filter values
	if !data.EServiceIDs.IsNull() && !data.EServiceIDs.IsUnknown() {
		var ids []string
		resp.Diagnostics.Append(data.EServiceIDs.ElementsAs(ctx, &ids, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		params.EServiceIDs = parseUUIDs(ids)
	}

	if !data.Title.IsNull() && !data.Title.IsUnknown() {
		title := data.Title.ValueString()
		params.Title = &title
	}

	if !data.ConsumerIDs.IsNull() && !data.ConsumerIDs.IsUnknown() {
		var ids []string
		resp.Diagnostics.Append(data.ConsumerIDs.ElementsAs(ctx, &ids, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		params.ConsumerIDs = parseUUIDs(ids)
	}

	if !data.States.IsNull() && !data.States.IsUnknown() {
		var states []string
		resp.Diagnostics.Append(data.States.ElementsAs(ctx, &states, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		params.States = states
	}

	// Auto-paginate
	var allPurposes []api.Purpose
	for {
		page, err := d.client.ListPurposes(ctx, params)
		if err != nil {
			resp.Diagnostics.AddError("Error listing purposes", err.Error())
			return
		}

		allPurposes = append(allPurposes, page.Results...)

		if int32(len(allPurposes)) >= page.Pagination.TotalCount {
			break
		}
		params.Offset += params.Limit
	}

	// Convert to nested models
	data.Purposes = make([]purposeItemModel, len(allPurposes))
	for i, p := range allPurposes {
		data.Purposes[i] = purposeToItemModel(&p)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func purposeToItemModel(p *api.Purpose) purposeItemModel {
	m := purposeItemModel{
		ID:                  types.StringValue(p.ID.String()),
		EServiceID:          types.StringValue(p.EServiceID.String()),
		ConsumerID:          types.StringValue(p.ConsumerID.String()),
		Title:               types.StringValue(p.Title),
		Description:         types.StringValue(p.Description),
		IsFreeOfCharge:      types.BoolValue(p.IsFreeOfCharge),
		IsRiskAnalysisValid: types.BoolValue(p.IsRiskAnalysisValid),
		CreatedAt:           types.StringValue(p.CreatedAt.Format(time.RFC3339)),
	}

	// Derive state inline
	if p.CurrentVersion != nil {
		m.State = types.StringValue(p.CurrentVersion.State)
		m.DailyCalls = types.Int64Value(int64(p.CurrentVersion.DailyCalls))
		m.VersionID = types.StringValue(p.CurrentVersion.ID.String())
	} else if p.WaitingForApprovalVersion != nil {
		m.State = types.StringValue("WAITING_FOR_APPROVAL")
		m.DailyCalls = types.Int64Value(int64(p.WaitingForApprovalVersion.DailyCalls))
		m.VersionID = types.StringValue(p.WaitingForApprovalVersion.ID.String())
	} else {
		m.State = types.StringValue("DRAFT")
		m.DailyCalls = types.Int64Null()
		m.VersionID = types.StringNull()
	}

	// Optional fields
	if p.UpdatedAt != nil {
		m.UpdatedAt = types.StringValue(p.UpdatedAt.Format(time.RFC3339))
	} else {
		m.UpdatedAt = types.StringNull()
	}

	if p.FreeOfChargeReason != nil {
		m.FreeOfChargeReason = types.StringValue(*p.FreeOfChargeReason)
	} else {
		m.FreeOfChargeReason = types.StringNull()
	}

	if p.SuspendedByConsumer != nil {
		m.SuspendedByConsumer = types.BoolValue(*p.SuspendedByConsumer)
	} else {
		m.SuspendedByConsumer = types.BoolNull()
	}

	if p.SuspendedByProducer != nil {
		m.SuspendedByProducer = types.BoolValue(*p.SuspendedByProducer)
	} else {
		m.SuspendedByProducer = types.BoolNull()
	}

	if p.DelegationID != nil {
		m.DelegationID = types.StringValue(p.DelegationID.String())
	} else {
		m.DelegationID = types.StringNull()
	}

	return m
}
