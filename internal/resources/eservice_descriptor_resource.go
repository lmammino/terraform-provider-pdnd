package resources

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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

var _ resource.Resource = &eserviceDescriptorResource{}
var _ resource.ResourceWithImportState = &eserviceDescriptorResource{}

type eserviceDescriptorResource struct {
	client api.EServicesAPI
}

func NewEServiceDescriptorResource() resource.Resource {
	return &eserviceDescriptorResource{}
}

func (r *eserviceDescriptorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_eservice_descriptor"
}

func (r *eserviceDescriptorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a PDND e-service descriptor lifecycle.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Descriptor UUID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"eservice_id": schema.StringAttribute{
				Description: "E-Service UUID",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(uuidRegex, "must be a valid UUID"),
				},
			},
			"version": schema.StringAttribute{
				Description: "Descriptor version",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"desired_state": schema.StringAttribute{
				Description: "Desired descriptor state: DRAFT, PUBLISHED, or SUSPENDED",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("DRAFT", "PUBLISHED", "SUSPENDED"),
				},
			},
			"state": schema.StringAttribute{
				Description: "Observed server state of the descriptor",
				Computed:    true,
			},
			"agreement_approval_policy": schema.StringAttribute{
				Description: "Agreement approval policy: AUTOMATIC or MANUAL",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("AUTOMATIC", "MANUAL"),
				},
			},
			"audience": schema.ListAttribute{
				Description: "Audience for the descriptor",
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"daily_calls_per_consumer": schema.Int64Attribute{
				Description: "Daily calls per consumer (1-1,000,000,000)",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 1_000_000_000),
				},
			},
			"daily_calls_total": schema.Int64Attribute{
				Description: "Total daily calls (1-1,000,000,000)",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 1_000_000_000),
				},
			},
			"voucher_lifespan": schema.Int64Attribute{
				Description: "Voucher lifespan in seconds (60-86,400)",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.Between(60, 86_400),
				},
			},
			"server_urls": schema.ListAttribute{
				Description: "Server URLs",
				Computed:    true,
				ElementType: types.StringType,
			},
			"description": schema.StringAttribute{
				Description: "Descriptor description",
				Optional:    true,
			},
			"allow_waiting_for_approval": schema.BoolAttribute{
				Description: "If true, allow the descriptor to remain in WAITING_FOR_APPROVAL state after publishing",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
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

func (r *eserviceDescriptorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *eserviceDescriptorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.EServiceDescriptorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	desiredState := plan.DesiredState.ValueString()

	if desiredState == "SUSPENDED" {
		resp.Diagnostics.AddError(
			"Invalid Desired State",
			"Cannot create descriptor directly in SUSPENDED state",
		)
		return
	}

	eserviceID := uuid.MustParse(plan.EServiceID.ValueString())

	// Build audience slice
	var audience []string
	resp.Diagnostics.Append(plan.Audience.ElementsAs(ctx, &audience, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build seed
	seed := api.DescriptorSeed{
		AgreementApprovalPolicy: plan.AgreementApprovalPolicy.ValueString(),
		Audience:                audience,
		DailyCallsPerConsumer:   int32(plan.DailyCallsPerConsumer.ValueInt64()),
		DailyCallsTotal:         int32(plan.DailyCallsTotal.ValueInt64()),
		VoucherLifespan:         int32(plan.VoucherLifespan.ValueInt64()),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		desc := plan.Description.ValueString()
		seed.Description = &desc
	}

	descriptor, err := r.client.CreateDescriptor(ctx, eserviceID, seed)
	if err != nil {
		resp.Diagnostics.AddError("Error creating descriptor", err.Error())
		return
	}

	if desiredState == "DRAFT" {
		populateDescriptorModel(ctx, &plan, descriptor)
		plan.DesiredState = types.StringValue("DRAFT")
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	// desiredState == "PUBLISHED"
	err = r.client.PublishDescriptor(ctx, eserviceID, descriptor.ID)
	if err != nil {
		// Clean up the draft
		_ = r.client.DeleteDraftDescriptor(ctx, eserviceID, descriptor.ID)
		resp.Diagnostics.AddError("Error publishing descriptor", err.Error())
		return
	}

	// Refresh state after publish (publish returns void)
	refreshed, err := r.client.GetDescriptor(ctx, eserviceID, descriptor.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading descriptor after publish", err.Error())
		return
	}

	if refreshed.State == "PUBLISHED" {
		populateDescriptorModel(ctx, &plan, refreshed)
		plan.DesiredState = types.StringValue("PUBLISHED")
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	if refreshed.State == "WAITING_FOR_APPROVAL" {
		if plan.AllowWaitingForApproval.ValueBool() {
			populateDescriptorModel(ctx, &plan, refreshed)
			plan.DesiredState = types.StringValue("PUBLISHED")
			resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
			return
		}

		resp.Diagnostics.AddError(
			"Descriptor Entered WAITING_FOR_APPROVAL State",
			"Descriptor entered WAITING_FOR_APPROVAL state but allow_waiting_for_approval is false. "+
				"The descriptor cannot be automatically deleted because it is no longer in DRAFT state. "+
				"Please set allow_waiting_for_approval = true or manually manage this descriptor.",
		)
		// Still save state so Terraform knows about the resource
		populateDescriptorModel(ctx, &plan, refreshed)
		plan.DesiredState = types.StringValue("PUBLISHED")
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	// Unexpected state after publish
	populateDescriptorModel(ctx, &plan, refreshed)
	plan.DesiredState = types.StringValue("PUBLISHED")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *eserviceDescriptorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.EServiceDescriptorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	eserviceID := uuid.MustParse(state.EServiceID.ValueString())
	descriptorID := uuid.MustParse(state.ID.ValueString())

	descriptor, err := r.client.GetDescriptor(ctx, eserviceID, descriptorID)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading descriptor", err.Error())
		return
	}

	// Preserve desired_state and allow_waiting_for_approval from current state
	desiredState := state.DesiredState
	allowWaiting := state.AllowWaitingForApproval

	populateDescriptorModel(ctx, &state, descriptor)

	// Preserve eservice_id (not returned by descriptor API)
	state.EServiceID = types.StringValue(eserviceID.String())

	// If desired_state is null/unknown (e.g., after import), infer from observed state
	if desiredState.IsNull() || desiredState.IsUnknown() {
		inferred, err := inferDescriptorDesiredState(descriptor.State)
		if err != nil {
			resp.Diagnostics.AddError("Error importing descriptor", err.Error())
			return
		}
		state.DesiredState = types.StringValue(inferred)
	} else {
		state.DesiredState = desiredState
	}

	if allowWaiting.IsNull() || allowWaiting.IsUnknown() {
		state.AllowWaitingForApproval = types.BoolValue(false)
	} else {
		state.AllowWaitingForApproval = allowWaiting
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *eserviceDescriptorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan models.EServiceDescriptorResourceModel
	var state models.EServiceDescriptorResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	eserviceID := uuid.MustParse(state.EServiceID.ValueString())
	descriptorID := uuid.MustParse(state.ID.ValueString())
	currentState := state.State.ValueString()
	desiredState := plan.DesiredState.ValueString()

	// If current state is DRAFT and config fields changed, update the draft descriptor
	if currentState == "DRAFT" {
		var audience []string
		resp.Diagnostics.Append(plan.Audience.ElementsAs(ctx, &audience, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		approvalPolicy := plan.AgreementApprovalPolicy.ValueString()
		dailyCallsPerConsumer := int32(plan.DailyCallsPerConsumer.ValueInt64())
		dailyCallsTotal := int32(plan.DailyCallsTotal.ValueInt64())
		voucherLifespan := int32(plan.VoucherLifespan.ValueInt64())

		draftUpdate := api.DescriptorDraftUpdate{
			AgreementApprovalPolicy: &approvalPolicy,
			Audience:                audience,
			DailyCallsPerConsumer:   &dailyCallsPerConsumer,
			DailyCallsTotal:         &dailyCallsTotal,
			VoucherLifespan:         &voucherLifespan,
		}
		if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
			desc := plan.Description.ValueString()
			draftUpdate.Description = &desc
		}

		_, err := r.client.UpdateDraftDescriptor(ctx, eserviceID, descriptorID, draftUpdate)
		if err != nil {
			resp.Diagnostics.AddError("Error updating draft descriptor", err.Error())
			return
		}
	}

	// If current state is PUBLISHED, SUSPENDED, or DEPRECATED and quotas changed, update quotas
	if currentState == "PUBLISHED" || currentState == "SUSPENDED" || currentState == "DEPRECATED" {
		quotasChanged := plan.DailyCallsPerConsumer.ValueInt64() != state.DailyCallsPerConsumer.ValueInt64() ||
			plan.DailyCallsTotal.ValueInt64() != state.DailyCallsTotal.ValueInt64() ||
			plan.VoucherLifespan.ValueInt64() != state.VoucherLifespan.ValueInt64()

		if quotasChanged {
			dailyCallsPerConsumer := int32(plan.DailyCallsPerConsumer.ValueInt64())
			dailyCallsTotal := int32(plan.DailyCallsTotal.ValueInt64())
			voucherLifespan := int32(plan.VoucherLifespan.ValueInt64())

			quotasUpdate := api.DescriptorQuotasUpdate{
				DailyCallsPerConsumer: &dailyCallsPerConsumer,
				DailyCallsTotal:       &dailyCallsTotal,
				VoucherLifespan:       &voucherLifespan,
			}

			_, err := r.client.UpdatePublishedDescriptorQuotas(ctx, eserviceID, descriptorID, quotasUpdate)
			if err != nil {
				resp.Diagnostics.AddError("Error updating descriptor quotas", err.Error())
				return
			}
		}
	}

	// Compute and execute state transitions
	transitions, err := ComputeDescriptorTransitions(currentState, desiredState)
	if err != nil {
		resp.Diagnostics.AddError("Error computing descriptor transitions", err.Error())
		return
	}

	for _, t := range transitions {
		switch t.Type {
		case DescriptorTransitionPublish:
			err = r.client.PublishDescriptor(ctx, eserviceID, descriptorID)
		case DescriptorTransitionSuspend:
			err = r.client.SuspendDescriptor(ctx, eserviceID, descriptorID)
		case DescriptorTransitionUnsuspend:
			err = r.client.UnsuspendDescriptor(ctx, eserviceID, descriptorID)
		}
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Error executing %s transition", t.Type),
				err.Error(),
			)
			return
		}
	}

	// Refresh state from API
	descriptor, err := r.client.GetDescriptor(ctx, eserviceID, descriptorID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading descriptor after update", err.Error())
		return
	}

	populateDescriptorModel(ctx, &plan, descriptor)
	plan.EServiceID = types.StringValue(eserviceID.String())
	plan.DesiredState = types.StringValue(desiredState)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *eserviceDescriptorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.EServiceDescriptorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	currentState := state.State.ValueString()
	eserviceID := uuid.MustParse(state.EServiceID.ValueString())
	descriptorID := uuid.MustParse(state.ID.ValueString())

	if canDeleteDescriptor(currentState) {
		err := r.client.DeleteDraftDescriptor(ctx, eserviceID, descriptorID)
		if err != nil {
			if client.IsNotFound(err) {
				return
			}
			resp.Diagnostics.AddError("Error deleting descriptor", err.Error())
		}
		return
	}

	resp.Diagnostics.AddError(
		"Cannot Delete Descriptor",
		fmt.Sprintf(
			"Cannot delete descriptor %s in state %s. Only DRAFT descriptors can be deleted.",
			descriptorID, currentState,
		),
	)
}
