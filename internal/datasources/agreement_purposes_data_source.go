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

var _ datasource.DataSource = &agreementPurposesDataSource{}

type agreementPurposesDataSource struct {
	client api.AgreementsAPI
}

type agreementPurposesDataSourceModel struct {
	AgreementID types.String         `tfsdk:"agreement_id"`
	Purposes    []purposeNestedModel `tfsdk:"purposes"`
}

type purposeNestedModel struct {
	ID                  types.String `tfsdk:"id"`
	EServiceID          types.String `tfsdk:"eservice_id"`
	ConsumerID          types.String `tfsdk:"consumer_id"`
	Title               types.String `tfsdk:"title"`
	Description         types.String `tfsdk:"description"`
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
	IsRiskAnalysisValid types.Bool   `tfsdk:"is_risk_analysis_valid"`
	IsFreeOfCharge      types.Bool   `tfsdk:"is_free_of_charge"`
	FreeOfChargeReason  types.String `tfsdk:"free_of_charge_reason"`
	SuspendedByConsumer types.Bool   `tfsdk:"suspended_by_consumer"`
	SuspendedByProducer types.Bool   `tfsdk:"suspended_by_producer"`
	DelegationID        types.String `tfsdk:"delegation_id"`
}

func NewAgreementPurposesDataSource() datasource.DataSource {
	return &agreementPurposesDataSource{}
}

func (d *agreementPurposesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agreement_purposes"
}

func (d *agreementPurposesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists purposes associated with a PDND agreement.",
		Attributes: map[string]schema.Attribute{
			"agreement_id": schema.StringAttribute{
				Description: "Agreement UUID",
				Required:    true,
			},
			"purposes": schema.ListNestedAttribute{
				Description: "List of purposes",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                    schema.StringAttribute{Computed: true},
						"eservice_id":           schema.StringAttribute{Computed: true},
						"consumer_id":           schema.StringAttribute{Computed: true},
						"title":                 schema.StringAttribute{Computed: true},
						"description":           schema.StringAttribute{Computed: true},
						"created_at":            schema.StringAttribute{Computed: true},
						"updated_at":            schema.StringAttribute{Computed: true},
						"is_risk_analysis_valid": schema.BoolAttribute{Computed: true},
						"is_free_of_charge":     schema.BoolAttribute{Computed: true},
						"free_of_charge_reason": schema.StringAttribute{Computed: true},
						"suspended_by_consumer": schema.BoolAttribute{Computed: true},
						"suspended_by_producer": schema.BoolAttribute{Computed: true},
						"delegation_id":         schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *agreementPurposesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *agreementPurposesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data agreementPurposesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	agreementID, err := uuid.Parse(data.AgreementID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Agreement ID", fmt.Sprintf("Could not parse agreement_id as UUID: %s", err))
		return
	}

	params := api.PaginationParams{
		Offset: 0,
		Limit:  50,
	}

	// Auto-paginate
	var allPurposes []api.Purpose
	for {
		page, err := d.client.ListAgreementPurposes(ctx, agreementID, params)
		if err != nil {
			resp.Diagnostics.AddError("Error listing agreement purposes", err.Error())
			return
		}

		allPurposes = append(allPurposes, page.Results...)

		if int32(len(allPurposes)) >= page.Pagination.TotalCount {
			break
		}
		params.Offset += params.Limit
	}

	// Convert to nested models
	data.Purposes = make([]purposeNestedModel, len(allPurposes))
	for i, p := range allPurposes {
		data.Purposes[i] = purposeToNestedModel(&p)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func purposeToNestedModel(p *api.Purpose) purposeNestedModel {
	m := purposeNestedModel{
		ID:                  types.StringValue(p.ID.String()),
		EServiceID:          types.StringValue(p.EServiceID.String()),
		ConsumerID:          types.StringValue(p.ConsumerID.String()),
		Title:               types.StringValue(p.Title),
		Description:         types.StringValue(p.Description),
		CreatedAt:           types.StringValue(p.CreatedAt.Format(time.RFC3339)),
		IsRiskAnalysisValid: types.BoolValue(p.IsRiskAnalysisValid),
		IsFreeOfCharge:      types.BoolValue(p.IsFreeOfCharge),
	}

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
