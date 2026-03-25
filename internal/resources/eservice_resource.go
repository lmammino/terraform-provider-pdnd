package resources

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/lmammino/terraform-provider-pdnd/internal/client"
	"github.com/lmammino/terraform-provider-pdnd/internal/client/api"
	"github.com/lmammino/terraform-provider-pdnd/internal/models"
	"github.com/lmammino/terraform-provider-pdnd/internal/providerdata"
)

var _ resource.Resource = &eserviceResource{}
var _ resource.ResourceWithImportState = &eserviceResource{}

type eserviceResource struct {
	client api.EServicesAPI
}

func NewEServiceResource() resource.Resource {
	return &eserviceResource{}
}

func (r *eserviceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_eservice"
}

func (r *eserviceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a PDND e-service.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "E-Service UUID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "E-Service name (5-60 characters)",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(5, 60),
				},
			},
			"description": schema.StringAttribute{
				Description: "E-Service description (10-250 characters)",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(10, 250),
				},
			},
			"technology": schema.StringAttribute{
				Description: "E-Service technology: REST or SOAP",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("REST", "SOAP"),
				},
			},
			"mode": schema.StringAttribute{
				Description: "E-Service mode: RECEIVE or DELIVER",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("RECEIVE", "DELIVER"),
				},
			},
			"is_signal_hub_enabled": schema.BoolAttribute{
				Description: "Whether the signal hub is enabled",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"is_consumer_delegable": schema.BoolAttribute{
				Description: "Whether consumers can delegate",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"is_client_access_delegable": schema.BoolAttribute{
				Description: "Whether client access can be delegated",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"personal_data": schema.BoolAttribute{
				Description: "Whether the e-service handles personal data",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"producer_id": schema.StringAttribute{
				Description: "Producer UUID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"template_id": schema.StringAttribute{
				Description: "Template UUID",
				Computed:    true,
			},
			"initial_descriptor_id": schema.StringAttribute{
				Description: "Initial descriptor UUID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"initial_descriptor_agreement_approval_policy": schema.StringAttribute{
				Description: "Agreement approval policy for the initial descriptor: AUTOMATIC or MANUAL",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("AUTOMATIC", "MANUAL"),
				},
			},
			"initial_descriptor_audience": schema.ListAttribute{
				Description: "Audience for the initial descriptor",
				Required:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listUseStateForUnknown{},
				},
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"initial_descriptor_daily_calls_per_consumer": schema.Int64Attribute{
				Description: "Daily calls per consumer for the initial descriptor (1-1,000,000,000)",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64UseStateForUnknown{},
				},
				Validators: []validator.Int64{
					int64validator.Between(1, 1_000_000_000),
				},
			},
			"initial_descriptor_daily_calls_total": schema.Int64Attribute{
				Description: "Total daily calls for the initial descriptor (1-1,000,000,000)",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64UseStateForUnknown{},
				},
				Validators: []validator.Int64{
					int64validator.Between(1, 1_000_000_000),
				},
			},
			"initial_descriptor_voucher_lifespan": schema.Int64Attribute{
				Description: "Voucher lifespan in seconds for the initial descriptor (60-86,400)",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64UseStateForUnknown{},
				},
				Validators: []validator.Int64{
					int64validator.Between(60, 86_400),
				},
			},
			"initial_descriptor_description": schema.StringAttribute{
				Description: "Description for the initial descriptor",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *eserviceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	pd, ok := req.ProviderData.(*providerdata.ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data",
			fmt.Sprintf("Expected *providerdata.ProviderData, got: %T", req.ProviderData),
		)
		return
	}

	r.client = pd.EServicesAPI
}

func (r *eserviceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.EServiceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build audience slice
	var audience []string
	resp.Diagnostics.Append(plan.InitialDescriptorAudience.ElementsAs(ctx, &audience, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build descriptor seed
	descriptorSeed := api.DescriptorSeedForCreation{
		AgreementApprovalPolicy: plan.InitialDescriptorAgreementApprovalPolicy.ValueString(),
		Audience:                audience,
		DailyCallsPerConsumer:   int32(plan.InitialDescriptorDailyCallsPerConsumer.ValueInt64()),
		DailyCallsTotal:         int32(plan.InitialDescriptorDailyCallsTotal.ValueInt64()),
		VoucherLifespan:         int32(plan.InitialDescriptorVoucherLifespan.ValueInt64()),
	}
	if !plan.InitialDescriptorDescription.IsNull() && !plan.InitialDescriptorDescription.IsUnknown() {
		desc := plan.InitialDescriptorDescription.ValueString()
		descriptorSeed.Description = &desc
	}

	// Build e-service seed
	seed := api.EServiceSeed{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Technology:  plan.Technology.ValueString(),
		Mode:        plan.Mode.ValueString(),
		Descriptor:  descriptorSeed,
	}

	if !plan.IsSignalHubEnabled.IsNull() && !plan.IsSignalHubEnabled.IsUnknown() {
		v := plan.IsSignalHubEnabled.ValueBool()
		seed.IsSignalHubEnabled = &v
	}
	if !plan.IsConsumerDelegable.IsNull() && !plan.IsConsumerDelegable.IsUnknown() {
		v := plan.IsConsumerDelegable.ValueBool()
		seed.IsConsumerDelegable = &v
	}
	if !plan.IsClientAccessDelegable.IsNull() && !plan.IsClientAccessDelegable.IsUnknown() {
		v := plan.IsClientAccessDelegable.ValueBool()
		seed.IsClientAccessDelegable = &v
	}
	if !plan.PersonalData.IsNull() && !plan.PersonalData.IsUnknown() {
		v := plan.PersonalData.ValueBool()
		seed.PersonalData = &v
	}

	eservice, err := r.client.CreateEService(ctx, seed)
	if err != nil {
		resp.Diagnostics.AddError("Error creating e-service", err.Error())
		return
	}

	// List descriptors to get the initial descriptor ID
	descriptorsPage, err := r.client.ListDescriptors(ctx, eservice.ID, api.ListDescriptorsParams{
		Offset: 0,
		Limit:  1,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error listing descriptors after e-service creation", err.Error())
		return
	}

	initialDescriptor := findInitialDescriptor(descriptorsPage.Results)
	if initialDescriptor == nil {
		resp.Diagnostics.AddError(
			"Initial Descriptor Not Found",
			"No descriptor was found after creating the e-service",
		)
		return
	}

	populateEServiceModel(&plan, eservice, initialDescriptor)

	// Resolve initial_descriptor_description if still unknown (optional field not provided by user).
	if plan.InitialDescriptorDescription.IsUnknown() {
		if initialDescriptor != nil && initialDescriptor.Description != nil && *initialDescriptor.Description != "" {
			plan.InitialDescriptorDescription = types.StringValue(*initialDescriptor.Description)
		} else {
			plan.InitialDescriptorDescription = types.StringNull()
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *eserviceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.EServiceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := uuid.MustParse(state.ID.ValueString())
	eservice, err := r.client.GetEService(ctx, id)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading e-service", err.Error())
		return
	}

	// List all descriptors to find initial
	descriptorsPage, err := r.client.ListDescriptors(ctx, id, api.ListDescriptorsParams{
		Offset: 0,
		Limit:  50,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error listing descriptors", err.Error())
		return
	}

	initialDescriptor := findInitialDescriptor(descriptorsPage.Results)

	// Preserve initial_descriptor_* fields from current state (creation-only)
	prevApprovalPolicy := state.InitialDescriptorAgreementApprovalPolicy
	prevAudience := state.InitialDescriptorAudience
	prevDailyCallsPerConsumer := state.InitialDescriptorDailyCallsPerConsumer
	prevDailyCallsTotal := state.InitialDescriptorDailyCallsTotal
	prevVoucherLifespan := state.InitialDescriptorVoucherLifespan
	prevDescriptorDescription := state.InitialDescriptorDescription

	// Populate e-service fields from API
	state.ID = types.StringValue(eservice.ID.String())
	state.Name = types.StringValue(eservice.Name)
	state.Description = types.StringValue(eservice.Description)
	state.Technology = types.StringValue(eservice.Technology)
	state.Mode = types.StringValue(eservice.Mode)
	state.ProducerID = types.StringValue(eservice.ProducerID.String())

	if eservice.IsSignalHubEnabled != nil {
		state.IsSignalHubEnabled = types.BoolValue(*eservice.IsSignalHubEnabled)
	} else {
		state.IsSignalHubEnabled = types.BoolValue(false)
	}
	if eservice.IsConsumerDelegable != nil {
		state.IsConsumerDelegable = types.BoolValue(*eservice.IsConsumerDelegable)
	} else {
		state.IsConsumerDelegable = types.BoolValue(false)
	}
	if eservice.IsClientAccessDelegable != nil {
		state.IsClientAccessDelegable = types.BoolValue(*eservice.IsClientAccessDelegable)
	} else {
		state.IsClientAccessDelegable = types.BoolValue(false)
	}
	if eservice.PersonalData != nil {
		state.PersonalData = types.BoolValue(*eservice.PersonalData)
	} else {
		state.PersonalData = types.BoolValue(false)
	}

	if eservice.TemplateID != nil {
		state.TemplateID = types.StringValue(eservice.TemplateID.String())
	} else {
		state.TemplateID = types.StringNull()
	}

	// Set initial descriptor ID
	if initialDescriptor != nil {
		state.InitialDescriptorID = types.StringValue(initialDescriptor.ID.String())
	}

	// On import (initial_descriptor_* fields are null): populate from actual descriptor
	if prevApprovalPolicy.IsNull() || prevApprovalPolicy.IsUnknown() {
		if initialDescriptor != nil {
			populateInitialDescriptorFromAPI(&state, initialDescriptor)
		}
	} else {
		// Preserve creation-only values
		state.InitialDescriptorAgreementApprovalPolicy = prevApprovalPolicy
		state.InitialDescriptorAudience = prevAudience
		state.InitialDescriptorDailyCallsPerConsumer = prevDailyCallsPerConsumer
		state.InitialDescriptorDailyCallsTotal = prevDailyCallsTotal
		state.InitialDescriptorVoucherLifespan = prevVoucherLifespan
		state.InitialDescriptorDescription = prevDescriptorDescription
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *eserviceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan models.EServiceResourceModel
	var state models.EServiceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := uuid.MustParse(state.ID.ValueString())

	// List descriptors to check if any are non-DRAFT
	descriptorsPage, err := r.client.ListDescriptors(ctx, id, api.ListDescriptorsParams{
		Offset: 0,
		Limit:  50,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error listing descriptors", err.Error())
		return
	}

	if isEServiceDraft(descriptorsPage.Results) {
		// All descriptors are DRAFT: use UpdateDraftEService
		name := plan.Name.ValueString()
		desc := plan.Description.ValueString()
		tech := plan.Technology.ValueString()
		mode := plan.Mode.ValueString()

		signalHub := plan.IsSignalHubEnabled.ValueBool()
		consumerDelegable := plan.IsConsumerDelegable.ValueBool()
		clientAccessDelegable := plan.IsClientAccessDelegable.ValueBool()
		personalData := plan.PersonalData.ValueBool()

		draftUpdate := api.EServiceDraftUpdate{
			Name:                    &name,
			Description:             &desc,
			Technology:              &tech,
			Mode:                    &mode,
			IsSignalHubEnabled:      &signalHub,
			IsConsumerDelegable:     &consumerDelegable,
			IsClientAccessDelegable: &clientAccessDelegable,
			PersonalData:            &personalData,
		}

		_, err := r.client.UpdateDraftEService(ctx, id, draftUpdate)
		if err != nil {
			resp.Diagnostics.AddError("Error updating draft e-service", err.Error())
			return
		}
	} else {
		// Published path: call per-field endpoints as needed
		if plan.Name.ValueString() != state.Name.ValueString() {
			_, err := r.client.UpdatePublishedEServiceName(ctx, id, plan.Name.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Error updating e-service name", err.Error())
				return
			}
		}

		if plan.Description.ValueString() != state.Description.ValueString() {
			_, err := r.client.UpdatePublishedEServiceDescription(ctx, id, plan.Description.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Error updating e-service description", err.Error())
				return
			}
		}

		delegationChanged := plan.IsConsumerDelegable.ValueBool() != state.IsConsumerDelegable.ValueBool() ||
			plan.IsClientAccessDelegable.ValueBool() != state.IsClientAccessDelegable.ValueBool()
		if delegationChanged {
			consumerDelegable := plan.IsConsumerDelegable.ValueBool()
			clientAccessDelegable := plan.IsClientAccessDelegable.ValueBool()
			_, err := r.client.UpdatePublishedEServiceDelegation(ctx, id, api.EServiceDelegationUpdate{
				IsConsumerDelegable:     &consumerDelegable,
				IsClientAccessDelegable: &clientAccessDelegable,
			})
			if err != nil {
				resp.Diagnostics.AddError("Error updating e-service delegation", err.Error())
				return
			}
		}

		if plan.IsSignalHubEnabled.ValueBool() != state.IsSignalHubEnabled.ValueBool() {
			_, err := r.client.UpdatePublishedEServiceSignalHub(ctx, id, plan.IsSignalHubEnabled.ValueBool())
			if err != nil {
				resp.Diagnostics.AddError("Error updating e-service signal hub", err.Error())
				return
			}
		}
	}

	// Refresh from API
	eservice, err := r.client.GetEService(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading e-service after update", err.Error())
		return
	}

	// Re-list descriptors to find initial
	descriptorsPage, err = r.client.ListDescriptors(ctx, id, api.ListDescriptorsParams{
		Offset: 0,
		Limit:  50,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error listing descriptors after update", err.Error())
		return
	}

	initialDescriptor := findInitialDescriptor(descriptorsPage.Results)
	populateEServiceModel(&plan, eservice, initialDescriptor)

	// Preserve initial_descriptor_* from plan (creation-only, UseStateForUnknown)
	plan.InitialDescriptorAgreementApprovalPolicy = state.InitialDescriptorAgreementApprovalPolicy
	plan.InitialDescriptorAudience = state.InitialDescriptorAudience
	plan.InitialDescriptorDailyCallsPerConsumer = state.InitialDescriptorDailyCallsPerConsumer
	plan.InitialDescriptorDailyCallsTotal = state.InitialDescriptorDailyCallsTotal
	plan.InitialDescriptorVoucherLifespan = state.InitialDescriptorVoucherLifespan
	plan.InitialDescriptorDescription = state.InitialDescriptorDescription

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *eserviceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.EServiceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := uuid.MustParse(state.ID.ValueString())

	err := r.client.DeleteEService(ctx, id)
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		if client.IsConflict(err) {
			resp.Diagnostics.AddError(
				"Cannot Delete E-Service",
				fmt.Sprintf(
					"Cannot delete e-service %s: it has non-draft descriptors. Archive or delete all descriptors first.",
					id,
				),
			)
			return
		}
		resp.Diagnostics.AddError("Error deleting e-service", err.Error())
	}
}

// listUseStateForUnknown is a plan modifier for List attributes that preserves state for unknown values.
type listUseStateForUnknown struct{}

func (m listUseStateForUnknown) Description(_ context.Context) string {
	return "Use state value for unknown."
}

func (m listUseStateForUnknown) MarkdownDescription(_ context.Context) string {
	return "Use state value for unknown."
}

func (m listUseStateForUnknown) PlanModifyList(_ context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

// int64UseStateForUnknown is a plan modifier for Int64 attributes that preserves state for unknown values.
type int64UseStateForUnknown struct{}

func (m int64UseStateForUnknown) Description(_ context.Context) string {
	return "Use state value for unknown."
}

func (m int64UseStateForUnknown) MarkdownDescription(_ context.Context) string {
	return "Use state value for unknown."
}

func (m int64UseStateForUnknown) PlanModifyInt64(_ context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

// Ensure the eservice resource uses idPath from agreement_resource.go.
var _ = path.Root("id") // idPath is defined in agreement_resource.go
