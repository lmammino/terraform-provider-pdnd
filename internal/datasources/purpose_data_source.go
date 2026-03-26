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

var _ datasource.DataSource = &purposeDataSource{}

type purposeDataSource struct {
	client api.PurposesAPI
}

func NewPurposeDataSource() datasource.DataSource {
	return &purposeDataSource{}
}

func (d *purposeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_purpose"
}

func (d *purposeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single PDND purpose by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Purpose UUID",
				Required:    true,
			},
			"eservice_id": schema.StringAttribute{
				Description: "E-Service UUID",
				Computed:    true,
			},
			"consumer_id": schema.StringAttribute{
				Description: "Consumer UUID",
				Computed:    true,
			},
			"title": schema.StringAttribute{
				Description: "Purpose title",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Purpose description",
				Computed:    true,
			},
			"daily_calls": schema.Int64Attribute{
				Description: "Daily calls limit",
				Computed:    true,
			},
			"state": schema.StringAttribute{
				Description: "Purpose state",
				Computed:    true,
			},
			"is_free_of_charge": schema.BoolAttribute{
				Description: "Whether the purpose is free of charge",
				Computed:    true,
			},
			"free_of_charge_reason": schema.StringAttribute{
				Description: "Reason for being free of charge",
				Computed:    true,
			},
			"is_risk_analysis_valid": schema.BoolAttribute{
				Description: "Whether the risk analysis is valid",
				Computed:    true,
			},
			"suspended_by_consumer": schema.BoolAttribute{
				Description: "Whether suspended by the consumer",
				Computed:    true,
			},
			"suspended_by_producer": schema.BoolAttribute{
				Description: "Whether suspended by the producer",
				Computed:    true,
			},
			"delegation_id": schema.StringAttribute{
				Description: "Delegation UUID",
				Computed:    true,
			},
			"version_id": schema.StringAttribute{
				Description: "Current version UUID",
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
		},
	}
}

func (d *purposeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *purposeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.PurposeDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Could not parse purpose ID as UUID: %s", err))
		return
	}

	purpose, err := d.client.GetPurpose(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading purpose", err.Error())
		return
	}

	populatePurposeDataSourceModel(&data, purpose)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func populatePurposeDataSourceModel(model *models.PurposeDataSourceModel, p *api.Purpose) {
	model.ID = types.StringValue(p.ID.String())
	model.EServiceID = types.StringValue(p.EServiceID.String())
	model.ConsumerID = types.StringValue(p.ConsumerID.String())
	model.Title = types.StringValue(p.Title)
	model.Description = types.StringValue(p.Description)
	model.IsFreeOfCharge = types.BoolValue(p.IsFreeOfCharge)
	model.IsRiskAnalysisValid = types.BoolValue(p.IsRiskAnalysisValid)
	model.CreatedAt = types.StringValue(p.CreatedAt.Format(time.RFC3339))

	// Derive state inline (avoid cross-package dependency on resources)
	if p.CurrentVersion != nil {
		model.State = types.StringValue(p.CurrentVersion.State)
		model.DailyCalls = types.Int64Value(int64(p.CurrentVersion.DailyCalls))
		model.VersionID = types.StringValue(p.CurrentVersion.ID.String())
	} else if p.WaitingForApprovalVersion != nil {
		model.State = types.StringValue("WAITING_FOR_APPROVAL")
		model.DailyCalls = types.Int64Value(int64(p.WaitingForApprovalVersion.DailyCalls))
		model.VersionID = types.StringValue(p.WaitingForApprovalVersion.ID.String())
	} else {
		model.State = types.StringValue("DRAFT")
		model.DailyCalls = types.Int64Null()
		model.VersionID = types.StringNull()
	}

	// Optional fields
	if p.UpdatedAt != nil {
		model.UpdatedAt = types.StringValue(p.UpdatedAt.Format(time.RFC3339))
	} else {
		model.UpdatedAt = types.StringNull()
	}

	if p.FreeOfChargeReason != nil {
		model.FreeOfChargeReason = types.StringValue(*p.FreeOfChargeReason)
	} else {
		model.FreeOfChargeReason = types.StringNull()
	}

	if p.SuspendedByConsumer != nil {
		model.SuspendedByConsumer = types.BoolValue(*p.SuspendedByConsumer)
	} else {
		model.SuspendedByConsumer = types.BoolNull()
	}

	if p.SuspendedByProducer != nil {
		model.SuspendedByProducer = types.BoolValue(*p.SuspendedByProducer)
	} else {
		model.SuspendedByProducer = types.BoolNull()
	}

	if p.DelegationID != nil {
		model.DelegationID = types.StringValue(p.DelegationID.String())
	} else {
		model.DelegationID = types.StringNull()
	}
}
