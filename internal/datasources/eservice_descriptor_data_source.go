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

var _ datasource.DataSource = &eserviceDescriptorDataSource{}

type eserviceDescriptorDataSource struct {
	client api.EServicesAPI
}

type eserviceDescriptorDataSourceModel struct {
	EServiceID              types.String `tfsdk:"eservice_id"`
	ID                      types.String `tfsdk:"id"`
	Version                 types.String `tfsdk:"version"`
	State                   types.String `tfsdk:"state"`
	AgreementApprovalPolicy types.String `tfsdk:"agreement_approval_policy"`
	Audience                types.List   `tfsdk:"audience"`
	DailyCallsPerConsumer   types.Int64  `tfsdk:"daily_calls_per_consumer"`
	DailyCallsTotal         types.Int64  `tfsdk:"daily_calls_total"`
	VoucherLifespan         types.Int64  `tfsdk:"voucher_lifespan"`
	ServerUrls              types.List   `tfsdk:"server_urls"`
	Description             types.String `tfsdk:"description"`
	PublishedAt             types.String `tfsdk:"published_at"`
	SuspendedAt             types.String `tfsdk:"suspended_at"`
	DeprecatedAt            types.String `tfsdk:"deprecated_at"`
	ArchivedAt              types.String `tfsdk:"archived_at"`
}

func NewEServiceDescriptorDataSource() datasource.DataSource {
	return &eserviceDescriptorDataSource{}
}

func (d *eserviceDescriptorDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_eservice_descriptor"
}

func (d *eserviceDescriptorDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single PDND e-service descriptor by e-service ID and descriptor ID.",
		Attributes: map[string]schema.Attribute{
			"eservice_id": schema.StringAttribute{
				Description: "E-Service UUID",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "Descriptor UUID",
				Required:    true,
			},
			"version": schema.StringAttribute{
				Description: "Descriptor version",
				Computed:    true,
			},
			"state": schema.StringAttribute{
				Description: "Descriptor state",
				Computed:    true,
			},
			"agreement_approval_policy": schema.StringAttribute{
				Description: "Agreement approval policy",
				Computed:    true,
			},
			"audience": schema.ListAttribute{
				Description: "Audience list",
				Computed:    true,
				ElementType: types.StringType,
			},
			"daily_calls_per_consumer": schema.Int64Attribute{
				Description: "Daily calls per consumer",
				Computed:    true,
			},
			"daily_calls_total": schema.Int64Attribute{
				Description: "Total daily calls",
				Computed:    true,
			},
			"voucher_lifespan": schema.Int64Attribute{
				Description: "Voucher lifespan in seconds",
				Computed:    true,
			},
			"server_urls": schema.ListAttribute{
				Description: "Server URLs",
				Computed:    true,
				ElementType: types.StringType,
			},
			"description": schema.StringAttribute{
				Description: "Descriptor description",
				Computed:    true,
			},
			"published_at": schema.StringAttribute{
				Description: "Publication timestamp",
				Computed:    true,
			},
			"suspended_at": schema.StringAttribute{
				Description: "Suspension timestamp",
				Computed:    true,
			},
			"deprecated_at": schema.StringAttribute{
				Description: "Deprecation timestamp",
				Computed:    true,
			},
			"archived_at": schema.StringAttribute{
				Description: "Archival timestamp",
				Computed:    true,
			},
		},
	}
}

func (d *eserviceDescriptorDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = pd.EServicesAPI
}

func (d *eserviceDescriptorDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data eserviceDescriptorDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	eserviceID, err := uuid.Parse(data.EServiceID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid E-Service ID", fmt.Sprintf("Could not parse eservice_id as UUID: %s", err))
		return
	}

	descriptorID, err := uuid.Parse(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Descriptor ID", fmt.Sprintf("Could not parse id as UUID: %s", err))
		return
	}

	descriptor, err := d.client.GetDescriptor(ctx, eserviceID, descriptorID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading descriptor", err.Error())
		return
	}

	populateDescriptorDataSourceModel(ctx, &data, descriptor)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func populateDescriptorDataSourceModel(ctx context.Context, model *eserviceDescriptorDataSourceModel, d *api.Descriptor) {
	model.ID = types.StringValue(d.ID.String())
	model.Version = types.StringValue(d.Version)
	model.State = types.StringValue(d.State)
	model.AgreementApprovalPolicy = types.StringValue(d.AgreementApprovalPolicy)
	model.DailyCallsPerConsumer = types.Int64Value(int64(d.DailyCallsPerConsumer))
	model.DailyCallsTotal = types.Int64Value(int64(d.DailyCallsTotal))
	model.VoucherLifespan = types.Int64Value(int64(d.VoucherLifespan))

	// Audience
	model.Audience, _ = types.ListValueFrom(ctx, types.StringType, d.Audience)

	// ServerUrls
	if d.ServerUrls != nil {
		model.ServerUrls, _ = types.ListValueFrom(ctx, types.StringType, d.ServerUrls)
	} else {
		model.ServerUrls, _ = types.ListValueFrom(ctx, types.StringType, []string{})
	}

	// Description
	if d.Description != nil {
		model.Description = types.StringValue(*d.Description)
	} else {
		model.Description = types.StringNull()
	}

	// Timestamps
	if d.PublishedAt != nil {
		model.PublishedAt = types.StringValue(d.PublishedAt.Format(time.RFC3339))
	} else {
		model.PublishedAt = types.StringNull()
	}

	if d.SuspendedAt != nil {
		model.SuspendedAt = types.StringValue(d.SuspendedAt.Format(time.RFC3339))
	} else {
		model.SuspendedAt = types.StringNull()
	}

	if d.DeprecatedAt != nil {
		model.DeprecatedAt = types.StringValue(d.DeprecatedAt.Format(time.RFC3339))
	} else {
		model.DeprecatedAt = types.StringNull()
	}

	if d.ArchivedAt != nil {
		model.ArchivedAt = types.StringValue(d.ArchivedAt.Format(time.RFC3339))
	} else {
		model.ArchivedAt = types.StringNull()
	}
}
