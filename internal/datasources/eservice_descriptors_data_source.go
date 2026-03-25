package datasources

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/lmammino/terraform-provider-pdnd/internal/client/api"
	"github.com/lmammino/terraform-provider-pdnd/internal/providerdata"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
)

var _ datasource.DataSource = &eserviceDescriptorsDataSource{}

type eserviceDescriptorsDataSource struct {
	client api.EServicesAPI
}

type eserviceDescriptorsDataSourceModel struct {
	EServiceID  types.String              `tfsdk:"eservice_id"`
	State       types.String              `tfsdk:"state"`
	Descriptors []descriptorNestedModel   `tfsdk:"descriptors"`
}

type descriptorNestedModel struct {
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

func NewEServiceDescriptorsDataSource() datasource.DataSource {
	return &eserviceDescriptorsDataSource{}
}

func (d *eserviceDescriptorsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_eservice_descriptors"
}

func descriptorNestedSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id":                        schema.StringAttribute{Computed: true},
		"version":                   schema.StringAttribute{Computed: true},
		"state":                     schema.StringAttribute{Computed: true},
		"agreement_approval_policy": schema.StringAttribute{Computed: true},
		"audience": schema.ListAttribute{
			Computed:    true,
			ElementType: types.StringType,
		},
		"daily_calls_per_consumer": schema.Int64Attribute{Computed: true},
		"daily_calls_total":        schema.Int64Attribute{Computed: true},
		"voucher_lifespan":         schema.Int64Attribute{Computed: true},
		"server_urls": schema.ListAttribute{
			Computed:    true,
			ElementType: types.StringType,
		},
		"description":  schema.StringAttribute{Computed: true},
		"published_at": schema.StringAttribute{Computed: true},
		"suspended_at": schema.StringAttribute{Computed: true},
		"deprecated_at": schema.StringAttribute{Computed: true},
		"archived_at":  schema.StringAttribute{Computed: true},
	}
}

func (d *eserviceDescriptorsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists descriptors for a PDND e-service with optional state filter.",
		Attributes: map[string]schema.Attribute{
			"eservice_id": schema.StringAttribute{
				Description: "E-Service UUID",
				Required:    true,
			},
			"state": schema.StringAttribute{
				Description: "Filter by descriptor state",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("DRAFT", "PUBLISHED", "DEPRECATED", "SUSPENDED", "ARCHIVED", "WAITING_FOR_APPROVAL"),
				},
			},
			"descriptors": schema.ListNestedAttribute{
				Description: "List of descriptors",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: descriptorNestedSchema(),
				},
			},
		},
	}
}

func (d *eserviceDescriptorsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *eserviceDescriptorsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data eserviceDescriptorsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	eserviceID, err := uuid.Parse(data.EServiceID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid E-Service ID", fmt.Sprintf("Could not parse eservice_id as UUID: %s", err))
		return
	}

	params := api.ListDescriptorsParams{
		Offset: 0,
		Limit:  50,
	}

	if !data.State.IsNull() && !data.State.IsUnknown() {
		state := data.State.ValueString()
		params.State = &state
	}

	// Auto-paginate
	var allDescriptors []api.Descriptor
	for {
		page, err := d.client.ListDescriptors(ctx, eserviceID, params)
		if err != nil {
			resp.Diagnostics.AddError("Error listing descriptors", err.Error())
			return
		}

		allDescriptors = append(allDescriptors, page.Results...)

		if int32(len(allDescriptors)) >= page.Pagination.TotalCount {
			break
		}
		params.Offset += params.Limit
	}

	// Convert to nested models
	data.Descriptors = make([]descriptorNestedModel, len(allDescriptors))
	for i, desc := range allDescriptors {
		data.Descriptors[i] = descriptorToNestedModel(ctx, &desc)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func descriptorToNestedModel(ctx context.Context, d *api.Descriptor) descriptorNestedModel {
	m := descriptorNestedModel{
		ID:                      types.StringValue(d.ID.String()),
		Version:                 types.StringValue(d.Version),
		State:                   types.StringValue(d.State),
		AgreementApprovalPolicy: types.StringValue(d.AgreementApprovalPolicy),
		DailyCallsPerConsumer:   types.Int64Value(int64(d.DailyCallsPerConsumer)),
		DailyCallsTotal:         types.Int64Value(int64(d.DailyCallsTotal)),
		VoucherLifespan:         types.Int64Value(int64(d.VoucherLifespan)),
	}

	// Audience
	m.Audience, _ = types.ListValueFrom(ctx, types.StringType, d.Audience)

	// ServerUrls
	if d.ServerUrls != nil {
		m.ServerUrls, _ = types.ListValueFrom(ctx, types.StringType, d.ServerUrls)
	} else {
		m.ServerUrls, _ = types.ListValueFrom(ctx, types.StringType, []string{})
	}

	// Description
	if d.Description != nil {
		m.Description = types.StringValue(*d.Description)
	} else {
		m.Description = types.StringNull()
	}

	// Timestamps
	if d.PublishedAt != nil {
		m.PublishedAt = types.StringValue(d.PublishedAt.Format(time.RFC3339))
	} else {
		m.PublishedAt = types.StringNull()
	}

	if d.SuspendedAt != nil {
		m.SuspendedAt = types.StringValue(d.SuspendedAt.Format(time.RFC3339))
	} else {
		m.SuspendedAt = types.StringNull()
	}

	if d.DeprecatedAt != nil {
		m.DeprecatedAt = types.StringValue(d.DeprecatedAt.Format(time.RFC3339))
	} else {
		m.DeprecatedAt = types.StringNull()
	}

	if d.ArchivedAt != nil {
		m.ArchivedAt = types.StringValue(d.ArchivedAt.Format(time.RFC3339))
	} else {
		m.ArchivedAt = types.StringNull()
	}

	return m
}
