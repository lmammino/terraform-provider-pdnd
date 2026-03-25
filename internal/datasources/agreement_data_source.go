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

var _ datasource.DataSource = &agreementDataSource{}

type agreementDataSource struct {
	client api.AgreementsAPI
}

func NewAgreementDataSource() datasource.DataSource {
	return &agreementDataSource{}
}

func (d *agreementDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agreement"
}

func (d *agreementDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single PDND agreement by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Agreement UUID",
				Required:    true,
			},
			"eservice_id": schema.StringAttribute{
				Description: "E-Service UUID",
				Computed:    true,
			},
			"descriptor_id": schema.StringAttribute{
				Description: "Descriptor UUID",
				Computed:    true,
			},
			"producer_id": schema.StringAttribute{
				Description: "Producer UUID",
				Computed:    true,
			},
			"consumer_id": schema.StringAttribute{
				Description: "Consumer UUID",
				Computed:    true,
			},
			"delegation_id": schema.StringAttribute{
				Description: "Delegation UUID",
				Computed:    true,
			},
			"state": schema.StringAttribute{
				Description: "Agreement state",
				Computed:    true,
			},
			"suspended_by_consumer": schema.BoolAttribute{
				Description: "Whether the agreement is suspended by the consumer",
				Computed:    true,
			},
			"suspended_by_producer": schema.BoolAttribute{
				Description: "Whether the agreement is suspended by the producer",
				Computed:    true,
			},
			"suspended_by_platform": schema.BoolAttribute{
				Description: "Whether the agreement is suspended by the platform",
				Computed:    true,
			},
			"consumer_notes": schema.StringAttribute{
				Description: "Consumer notes",
				Computed:    true,
			},
			"rejection_reason": schema.StringAttribute{
				Description: "Reason for rejection",
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
			"suspended_at": schema.StringAttribute{
				Description: "Suspension timestamp",
				Computed:    true,
			},
		},
	}
}

func (d *agreementDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *agreementDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.AgreementDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Could not parse agreement ID as UUID: %s", err))
		return
	}

	agreement, err := d.client.GetAgreement(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading agreement", err.Error())
		return
	}

	populateDataSourceModel(&data, agreement)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func populateDataSourceModel(model *models.AgreementDataSourceModel, a *api.Agreement) {
	model.ID = types.StringValue(a.ID.String())
	model.EServiceID = types.StringValue(a.EServiceID.String())
	model.DescriptorID = types.StringValue(a.DescriptorID.String())
	model.ProducerID = types.StringValue(a.ProducerID.String())
	model.ConsumerID = types.StringValue(a.ConsumerID.String())
	model.State = types.StringValue(a.State)
	model.CreatedAt = types.StringValue(a.CreatedAt.Format(time.RFC3339))

	if a.DelegationID != nil {
		model.DelegationID = types.StringValue(a.DelegationID.String())
	} else {
		model.DelegationID = types.StringNull()
	}

	if a.SuspendedByConsumer != nil {
		model.SuspendedByConsumer = types.BoolValue(*a.SuspendedByConsumer)
	} else {
		model.SuspendedByConsumer = types.BoolNull()
	}

	if a.SuspendedByProducer != nil {
		model.SuspendedByProducer = types.BoolValue(*a.SuspendedByProducer)
	} else {
		model.SuspendedByProducer = types.BoolNull()
	}

	if a.SuspendedByPlatform != nil {
		model.SuspendedByPlatform = types.BoolValue(*a.SuspendedByPlatform)
	} else {
		model.SuspendedByPlatform = types.BoolNull()
	}

	if a.ConsumerNotes != nil {
		model.ConsumerNotes = types.StringValue(*a.ConsumerNotes)
	} else {
		model.ConsumerNotes = types.StringNull()
	}

	if a.RejectionReason != nil {
		model.RejectionReason = types.StringValue(*a.RejectionReason)
	} else {
		model.RejectionReason = types.StringNull()
	}

	if a.UpdatedAt != nil {
		model.UpdatedAt = types.StringValue(a.UpdatedAt.Format(time.RFC3339))
	} else {
		model.UpdatedAt = types.StringNull()
	}

	if a.SuspendedAt != nil {
		model.SuspendedAt = types.StringValue(a.SuspendedAt.Format(time.RFC3339))
	} else {
		model.SuspendedAt = types.StringNull()
	}
}
