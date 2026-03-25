package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/lmammino/terraform-provider-pdnd/internal/client/api"
	"github.com/lmammino/terraform-provider-pdnd/internal/providerdata"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
)

var _ datasource.DataSource = &eservicesDataSource{}

type eservicesDataSource struct {
	client api.EServicesAPI
}

type eservicesDataSourceModel struct {
	ProducerIDs types.List            `tfsdk:"producer_ids"`
	Name        types.String          `tfsdk:"name"`
	Technology  types.String          `tfsdk:"technology"`
	Mode        types.String          `tfsdk:"mode"`
	EServices   []eserviceNestedModel `tfsdk:"eservices"`
}

type eserviceNestedModel struct {
	ID                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	Description             types.String `tfsdk:"description"`
	Technology              types.String `tfsdk:"technology"`
	Mode                    types.String `tfsdk:"mode"`
	IsSignalHubEnabled      types.Bool   `tfsdk:"is_signal_hub_enabled"`
	IsConsumerDelegable     types.Bool   `tfsdk:"is_consumer_delegable"`
	IsClientAccessDelegable types.Bool   `tfsdk:"is_client_access_delegable"`
	PersonalData            types.Bool   `tfsdk:"personal_data"`
	ProducerID              types.String `tfsdk:"producer_id"`
	TemplateID              types.String `tfsdk:"template_id"`
}

func NewEServicesDataSource() datasource.DataSource {
	return &eservicesDataSource{}
}

func (d *eservicesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_eservices"
}

func eserviceNestedSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id":                        schema.StringAttribute{Computed: true},
		"name":                      schema.StringAttribute{Computed: true},
		"description":               schema.StringAttribute{Computed: true},
		"technology":                schema.StringAttribute{Computed: true},
		"mode":                      schema.StringAttribute{Computed: true},
		"is_signal_hub_enabled":     schema.BoolAttribute{Computed: true},
		"is_consumer_delegable":     schema.BoolAttribute{Computed: true},
		"is_client_access_delegable": schema.BoolAttribute{Computed: true},
		"personal_data":             schema.BoolAttribute{Computed: true},
		"producer_id":               schema.StringAttribute{Computed: true},
		"template_id":               schema.StringAttribute{Computed: true},
	}
}

func (d *eservicesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists PDND e-services with optional filters.",
		Attributes: map[string]schema.Attribute{
			"producer_ids": schema.ListAttribute{
				Description: "Filter by producer UUIDs",
				Optional:    true,
				ElementType: types.StringType,
			},
			"name": schema.StringAttribute{
				Description: "Filter by e-service name",
				Optional:    true,
			},
			"technology": schema.StringAttribute{
				Description: "Filter by technology (REST or SOAP)",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("REST", "SOAP"),
				},
			},
			"mode": schema.StringAttribute{
				Description: "Filter by mode (RECEIVE or DELIVER)",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("RECEIVE", "DELIVER"),
				},
			},
			"eservices": schema.ListNestedAttribute{
				Description: "List of e-services",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: eserviceNestedSchema(),
				},
			},
		},
	}
}

func (d *eservicesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *eservicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data eservicesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := api.ListEServicesParams{
		Offset: 0,
		Limit:  50,
	}

	// Extract filter values
	if !data.ProducerIDs.IsNull() && !data.ProducerIDs.IsUnknown() {
		var ids []string
		resp.Diagnostics.Append(data.ProducerIDs.ElementsAs(ctx, &ids, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		params.ProducerIDs = parseUUIDs(ids)
	}

	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		name := data.Name.ValueString()
		params.Name = &name
	}

	if !data.Technology.IsNull() && !data.Technology.IsUnknown() {
		tech := data.Technology.ValueString()
		params.Technology = &tech
	}

	if !data.Mode.IsNull() && !data.Mode.IsUnknown() {
		mode := data.Mode.ValueString()
		params.Mode = &mode
	}

	// Auto-paginate
	var allEServices []api.EService
	for {
		page, err := d.client.ListEServices(ctx, params)
		if err != nil {
			resp.Diagnostics.AddError("Error listing e-services", err.Error())
			return
		}

		allEServices = append(allEServices, page.Results...)

		if int32(len(allEServices)) >= page.Pagination.TotalCount {
			break
		}
		params.Offset += params.Limit
	}

	// Convert to nested models
	data.EServices = make([]eserviceNestedModel, len(allEServices))
	for i, es := range allEServices {
		data.EServices[i] = eserviceToNestedModel(&es)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func eserviceToNestedModel(es *api.EService) eserviceNestedModel {
	m := eserviceNestedModel{
		ID:          types.StringValue(es.ID.String()),
		Name:        types.StringValue(es.Name),
		Description: types.StringValue(es.Description),
		Technology:  types.StringValue(es.Technology),
		Mode:        types.StringValue(es.Mode),
		ProducerID:  types.StringValue(es.ProducerID.String()),
	}

	if es.IsSignalHubEnabled != nil {
		m.IsSignalHubEnabled = types.BoolValue(*es.IsSignalHubEnabled)
	} else {
		m.IsSignalHubEnabled = types.BoolNull()
	}
	if es.IsConsumerDelegable != nil {
		m.IsConsumerDelegable = types.BoolValue(*es.IsConsumerDelegable)
	} else {
		m.IsConsumerDelegable = types.BoolNull()
	}
	if es.IsClientAccessDelegable != nil {
		m.IsClientAccessDelegable = types.BoolValue(*es.IsClientAccessDelegable)
	} else {
		m.IsClientAccessDelegable = types.BoolNull()
	}
	if es.PersonalData != nil {
		m.PersonalData = types.BoolValue(*es.PersonalData)
	} else {
		m.PersonalData = types.BoolNull()
	}

	if es.TemplateID != nil {
		m.TemplateID = types.StringValue(es.TemplateID.String())
	} else {
		m.TemplateID = types.StringNull()
	}

	return m
}
