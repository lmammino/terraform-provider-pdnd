package resources

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
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

var _ resource.Resource = &purposeResource{}
var _ resource.ResourceWithImportState = &purposeResource{}

type purposeResource struct {
	client api.PurposesAPI
}

func NewPurposeResource() resource.Resource {
	return &purposeResource{}
}

func (r *purposeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_purpose"
}

func (r *purposeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a PDND purpose lifecycle.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Purpose UUID",
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
			"title": schema.StringAttribute{
				Description: "Purpose title (5-60 characters)",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(5, 60),
				},
			},
			"description": schema.StringAttribute{
				Description: "Purpose description (10-250 characters)",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(10, 250),
				},
			},
			"daily_calls": schema.Int64Attribute{
				Description: "Maximum daily API calls (1-1,000,000,000)",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 1_000_000_000),
				},
			},
			"is_free_of_charge": schema.BoolAttribute{
				Description: "Whether the service is free of charge",
				Required:    true,
			},
			"free_of_charge_reason": schema.StringAttribute{
				Description: "Reason for free of charge, if applicable",
				Optional:    true,
			},
			"delegation_id": schema.StringAttribute{
				Description: "Delegation UUID for delegated consumers",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(uuidRegex, "must be a valid UUID"),
				},
			},
			"desired_state": schema.StringAttribute{
				Description: "Desired purpose state: DRAFT, ACTIVE, or SUSPENDED",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("DRAFT", "ACTIVE", "SUSPENDED"),
				},
			},
			"allow_waiting_for_approval": schema.BoolAttribute{
				Description: "If true, allow the purpose to remain in WAITING_FOR_APPROVAL state after activation",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"state": schema.StringAttribute{
				Description: "Observed server state of the purpose",
				Computed:    true,
			},
			"consumer_id": schema.StringAttribute{
				Description: "Consumer tenant UUID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"is_risk_analysis_valid": schema.BoolAttribute{
				Description: "Whether the risk analysis is valid",
				Computed:    true,
			},
			"suspended_by_consumer": schema.BoolAttribute{
				Description: "Whether suspended by consumer",
				Computed:    true,
			},
			"suspended_by_producer": schema.BoolAttribute{
				Description: "Whether suspended by producer",
				Computed:    true,
			},
			"version_id": schema.StringAttribute{
				Description: "Current version UUID",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Creation timestamp",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "Last update timestamp",
				Computed:    true,
			},
			"first_activation_at": schema.StringAttribute{
				Description: "First activation timestamp",
				Computed:    true,
			},
			"suspended_at": schema.StringAttribute{
				Description: "Suspension timestamp",
				Computed:    true,
			},
			"rejection_reason": schema.StringAttribute{
				Description: "Rejection reason, if rejected",
				Computed:    true,
			},
		},
	}
}

func (r *purposeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = pd.PurposesAPI
}

func (r *purposeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.PurposeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	desiredState := plan.DesiredState.ValueString()

	if desiredState == "SUSPENDED" {
		resp.Diagnostics.AddError(
			"Invalid Desired State",
			"Cannot create purpose directly in SUSPENDED state",
		)
		return
	}

	eserviceID := uuid.MustParse(plan.EServiceID.ValueString())

	seed := api.PurposeSeed{
		EServiceID:     eserviceID,
		Title:          plan.Title.ValueString(),
		Description:    plan.Description.ValueString(),
		DailyCalls:     int32(plan.DailyCalls.ValueInt64()),
		IsFreeOfCharge: plan.IsFreeOfCharge.ValueBool(),
	}
	if !plan.FreeOfChargeReason.IsNull() && !plan.FreeOfChargeReason.IsUnknown() {
		reason := plan.FreeOfChargeReason.ValueString()
		seed.FreeOfChargeReason = &reason
	}
	if !plan.DelegationID.IsNull() && !plan.DelegationID.IsUnknown() {
		id := uuid.MustParse(plan.DelegationID.ValueString())
		seed.DelegationID = &id
	}

	purpose, err := r.client.CreatePurpose(ctx, seed)
	if err != nil {
		resp.Diagnostics.AddError("Error creating purpose", err.Error())
		return
	}

	if desiredState == "DRAFT" {
		populateModelFromPurpose(&plan, purpose)
		plan.DesiredState = types.StringValue("DRAFT")
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	// desiredState == "ACTIVE"
	activated, err := r.client.ActivatePurpose(ctx, purpose.ID, nil)
	if err != nil {
		// Clean up the draft
		_ = r.client.DeletePurpose(ctx, purpose.ID)
		resp.Diagnostics.AddError("Error activating purpose", err.Error())
		return
	}

	activatedState := derivePurposeState(activated)

	if activatedState == "ACTIVE" {
		populateModelFromPurpose(&plan, activated)
		plan.DesiredState = types.StringValue("ACTIVE")
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	if activatedState == "WAITING_FOR_APPROVAL" {
		if plan.AllowWaitingForApproval.ValueBool() {
			populateModelFromPurpose(&plan, activated)
			plan.DesiredState = types.StringValue("ACTIVE")
			resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
			return
		}

		resp.Diagnostics.AddError(
			"Purpose Entered WAITING_FOR_APPROVAL State",
			"Purpose entered WAITING_FOR_APPROVAL state but allow_waiting_for_approval is false. "+
				"Set allow_waiting_for_approval = true or reduce daily_calls below the threshold.",
		)
		// Still save state so Terraform knows about the resource
		populateModelFromPurpose(&plan, activated)
		plan.DesiredState = types.StringValue("ACTIVE")
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	// Unexpected state
	populateModelFromPurpose(&plan, activated)
	plan.DesiredState = types.StringValue("ACTIVE")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *purposeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.PurposeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	purposeID := uuid.MustParse(state.ID.ValueString())

	purpose, err := r.client.GetPurpose(ctx, purposeID)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading purpose", err.Error())
		return
	}

	desiredState := state.DesiredState
	allowWaiting := state.AllowWaitingForApproval

	populateModelFromPurpose(&state, purpose)

	// If desired_state is null/unknown (e.g., after import), infer from observed state
	if desiredState.IsNull() || desiredState.IsUnknown() {
		inferred, err := inferPurposeDesiredState(derivePurposeState(purpose))
		if err != nil {
			resp.Diagnostics.AddError("Error importing purpose", err.Error())
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

func (r *purposeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan models.PurposeResourceModel
	var state models.PurposeResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	purposeID := uuid.MustParse(state.ID.ValueString())
	currentState := state.State.ValueString()
	desiredState := plan.DesiredState.ValueString()

	// If current state is DRAFT, update fields via PATCH
	if currentState == "DRAFT" {
		title := plan.Title.ValueString()
		description := plan.Description.ValueString()
		dailyCalls := int32(plan.DailyCalls.ValueInt64())
		isFreeOfCharge := plan.IsFreeOfCharge.ValueBool()

		draftUpdate := api.PurposeDraftUpdate{
			Title:          &title,
			Description:    &description,
			DailyCalls:     &dailyCalls,
			IsFreeOfCharge: &isFreeOfCharge,
		}
		if !plan.FreeOfChargeReason.IsNull() && !plan.FreeOfChargeReason.IsUnknown() {
			reason := plan.FreeOfChargeReason.ValueString()
			draftUpdate.FreeOfChargeReason = &reason
		}

		_, err := r.client.UpdateDraftPurpose(ctx, purposeID, draftUpdate)
		if err != nil {
			resp.Diagnostics.AddError("Error updating draft purpose", err.Error())
			return
		}
	}

	// If current state is ACTIVE or SUSPENDED
	if currentState == "ACTIVE" || currentState == "SUSPENDED" {
		// Check for immutable field changes
		if plan.Title.ValueString() != state.Title.ValueString() ||
			plan.Description.ValueString() != state.Description.ValueString() ||
			plan.IsFreeOfCharge.ValueBool() != state.IsFreeOfCharge.ValueBool() ||
			plan.FreeOfChargeReason.ValueString() != state.FreeOfChargeReason.ValueString() {
			resp.Diagnostics.AddError(
				"Cannot Update Non-Draft Purpose Fields",
				"title, description, is_free_of_charge, and free_of_charge_reason cannot be changed when the purpose is not in DRAFT state. "+
					"Only daily_calls and desired_state can be modified.",
			)
			return
		}

		// Handle daily_calls change via new version
		if plan.DailyCalls.ValueInt64() != state.DailyCalls.ValueInt64() {
			_, err := r.client.CreatePurposeVersion(ctx, purposeID, api.PurposeVersionSeed{
				DailyCalls: int32(plan.DailyCalls.ValueInt64()),
			})
			if err != nil {
				resp.Diagnostics.AddError("Error creating purpose version", err.Error())
				return
			}
		}
	}

	// Compute and execute state transitions
	transitions, err := ComputePurposeTransitions(currentState, desiredState)
	if err != nil {
		resp.Diagnostics.AddError("Error computing purpose transitions", err.Error())
		return
	}

	for _, t := range transitions {
		switch t.Type {
		case PurposeTransitionActivate:
			_, err = r.client.ActivatePurpose(ctx, purposeID, nil)
		case PurposeTransitionSuspend:
			_, err = r.client.SuspendPurpose(ctx, purposeID, nil)
		case PurposeTransitionUnsuspend:
			_, err = r.client.UnsuspendPurpose(ctx, purposeID, nil)
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
	purpose, err := r.client.GetPurpose(ctx, purposeID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading purpose after update", err.Error())
		return
	}

	populateModelFromPurpose(&plan, purpose)
	plan.DesiredState = types.StringValue(desiredState)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *purposeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.PurposeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	currentState := state.State.ValueString()
	purposeID := uuid.MustParse(state.ID.ValueString())

	if canDeletePurpose(currentState) {
		err := r.client.DeletePurpose(ctx, purposeID)
		if err != nil {
			if client.IsNotFound(err) {
				return
			}
			resp.Diagnostics.AddError("Error deleting purpose", err.Error())
		}
		return
	}

	resp.Diagnostics.AddError(
		"Cannot Delete Purpose",
		fmt.Sprintf(
			"Cannot delete purpose %s in state %s. Only DRAFT and WAITING_FOR_APPROVAL purposes can be deleted.",
			purposeID, currentState,
		),
	)
}
