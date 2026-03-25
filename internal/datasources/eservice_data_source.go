package datasources

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/lmammino/terraform-provider-pdnd/internal/client/api"
	"github.com/lmammino/terraform-provider-pdnd/internal/models"
	"github.com/lmammino/terraform-provider-pdnd/internal/providerdata"
)

var _ datasource.DataSource = &eserviceDataSource{}

type eserviceDataSource struct {
	client api.EServicesAPI
}

func NewEServiceDataSource() datasource.DataSource {
	return &eserviceDataSource{}
}

func (d *eserviceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_eservice"
}

func (d *eserviceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single PDND e-service by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "E-Service UUID",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "E-Service name",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "E-Service description",
				Computed:    true,
			},
			"technology": schema.StringAttribute{
				Description: "E-Service technology (REST or SOAP)",
				Computed:    true,
			},
			"mode": schema.StringAttribute{
				Description: "E-Service mode (RECEIVE or DELIVER)",
				Computed:    true,
			},
			"is_signal_hub_enabled": schema.BoolAttribute{
				Description: "Whether SignalHub is enabled",
				Computed:    true,
			},
			"is_consumer_delegable": schema.BoolAttribute{
				Description: "Whether the e-service is consumer delegable",
				Computed:    true,
			},
			"is_client_access_delegable": schema.BoolAttribute{
				Description: "Whether the e-service is client access delegable",
				Computed:    true,
			},
			"personal_data": schema.BoolAttribute{
				Description: "Whether the e-service handles personal data",
				Computed:    true,
			},
			"producer_id": schema.StringAttribute{
				Description: "Producer UUID",
				Computed:    true,
			},
			"template_id": schema.StringAttribute{
				Description: "Template UUID",
				Computed:    true,
			},
		},
	}
}

func (d *eserviceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *eserviceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.EServiceDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Could not parse e-service ID as UUID: %s", err))
		return
	}

	eservice, err := d.client.GetEService(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading e-service", err.Error())
		return
	}

	populateEServiceDataSourceModel(&data, eservice)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func populateEServiceDataSourceModel(model *models.EServiceDataSourceModel, es *api.EService) {
	model.ID = types.StringValue(es.ID.String())
	model.Name = types.StringValue(es.Name)
	model.Description = types.StringValue(es.Description)
	model.Technology = types.StringValue(es.Technology)
	model.Mode = types.StringValue(es.Mode)
	model.ProducerID = types.StringValue(es.ProducerID.String())

	if es.IsSignalHubEnabled != nil {
		model.IsSignalHubEnabled = types.BoolValue(*es.IsSignalHubEnabled)
	} else {
		model.IsSignalHubEnabled = types.BoolNull()
	}
	if es.IsConsumerDelegable != nil {
		model.IsConsumerDelegable = types.BoolValue(*es.IsConsumerDelegable)
	} else {
		model.IsConsumerDelegable = types.BoolNull()
	}
	if es.IsClientAccessDelegable != nil {
		model.IsClientAccessDelegable = types.BoolValue(*es.IsClientAccessDelegable)
	} else {
		model.IsClientAccessDelegable = types.BoolNull()
	}
	if es.PersonalData != nil {
		model.PersonalData = types.BoolValue(*es.PersonalData)
	} else {
		model.PersonalData = types.BoolNull()
	}

	if es.TemplateID != nil {
		model.TemplateID = types.StringValue(es.TemplateID.String())
	} else {
		model.TemplateID = types.StringNull()
	}
}
