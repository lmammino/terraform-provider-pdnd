package resources

import (
	"context"
	"fmt"
	"regexp"

	"github.com/google/uuid"
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

var idPath = path.Root("id")

var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

var _ resource.Resource = &agreementResource{}
var _ resource.ResourceWithImportState = &agreementResource{}

type agreementResource struct {
	client api.AgreementsAPI
}

func NewAgreementResource() resource.Resource {
	return &agreementResource{}
}

func (r *agreementResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agreement"
}

func (r *agreementResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a PDND agreement lifecycle.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Agreement UUID",
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
			"descriptor_id": schema.StringAttribute{
				Description: "Descriptor UUID",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(uuidRegex, "must be a valid UUID"),
				},
			},
			"delegation_id": schema.StringAttribute{
				Description: "Delegation UUID",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(uuidRegex, "must be a valid UUID"),
				},
			},
			"desired_state": schema.StringAttribute{
				Description: "Desired agreement state: DRAFT, ACTIVE, or SUSPENDED",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("DRAFT", "ACTIVE", "SUSPENDED"),
				},
			},
			"consumer_notes": schema.StringAttribute{
				Description: "Consumer notes sent on agreement submission",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 1000),
				},
			},
			"allow_pending": schema.BoolAttribute{
				Description: "If true, allow the agreement to remain in PENDING state after submission",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"state": schema.StringAttribute{
				Description: "Observed server state of the agreement",
				Computed:    true,
			},
			"producer_id": schema.StringAttribute{
				Description: "Producer UUID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"consumer_id": schema.StringAttribute{
				Description: "Consumer UUID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
			"rejection_reason": schema.StringAttribute{
				Description: "Reason for rejection",
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
			"suspended_at": schema.StringAttribute{
				Description: "Suspension timestamp",
				Computed:    true,
			},
		},
	}
}

func (r *agreementResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = pd.AgreementsAPI
}

func (r *agreementResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.AgreementResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	desiredState := plan.DesiredState.ValueString()

	if desiredState == "SUSPENDED" {
		resp.Diagnostics.AddError(
			"Invalid Desired State",
			"Cannot create agreement directly in SUSPENDED state",
		)
		return
	}

	// Build seed
	eserviceID := uuid.MustParse(plan.EServiceID.ValueString())
	descriptorID := uuid.MustParse(plan.DescriptorID.ValueString())
	seed := api.AgreementSeed{
		EServiceID:   eserviceID,
		DescriptorID: descriptorID,
	}
	if !plan.DelegationID.IsNull() && !plan.DelegationID.IsUnknown() {
		delegID := uuid.MustParse(plan.DelegationID.ValueString())
		seed.DelegationID = &delegID
	}

	agreement, err := r.client.CreateAgreement(ctx, seed)
	if err != nil {
		resp.Diagnostics.AddError("Error creating agreement", err.Error())
		return
	}

	if desiredState == "DRAFT" {
		populateModelFromAgreement(&plan, agreement)
		plan.DesiredState = types.StringValue("DRAFT")
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	// desiredState == "ACTIVE"
	submission := api.AgreementSubmission{}
	if !plan.ConsumerNotes.IsNull() && !plan.ConsumerNotes.IsUnknown() {
		notes := plan.ConsumerNotes.ValueString()
		submission.ConsumerNotes = &notes
	}

	submitted, err := r.client.SubmitAgreement(ctx, agreement.ID, submission)
	if err != nil {
		// Clean up the draft
		_ = r.client.DeleteAgreement(ctx, agreement.ID)
		resp.Diagnostics.AddError("Error submitting agreement", err.Error())
		return
	}

	if submitted.State == "ACTIVE" {
		populateModelFromAgreement(&plan, submitted)
		plan.DesiredState = types.StringValue("ACTIVE")
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	if submitted.State == "PENDING" {
		if plan.AllowPending.ValueBool() {
			populateModelFromAgreement(&plan, submitted)
			plan.DesiredState = types.StringValue("ACTIVE")
			resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
			return
		}

		// Clean up
		_ = r.client.DeleteAgreement(ctx, submitted.ID)
		resp.Diagnostics.AddError(
			"Agreement Entered PENDING State",
			"Agreement entered PENDING state but allow_pending is false",
		)
		return
	}

	// Unexpected state after submission
	populateModelFromAgreement(&plan, submitted)
	plan.DesiredState = types.StringValue("ACTIVE")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *agreementResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.AgreementResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := uuid.MustParse(state.ID.ValueString())
	agreement, err := r.client.GetAgreement(ctx, id)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading agreement", err.Error())
		return
	}

	// Preserve desired_state and allow_pending from current state
	desiredState := state.DesiredState
	allowPending := state.AllowPending

	populateModelFromAgreement(&state, agreement)

	// If desired_state is null/unknown (e.g., after import), infer from observed state
	if desiredState.IsNull() || desiredState.IsUnknown() {
		inferred, err := inferDesiredState(agreement.State)
		if err != nil {
			resp.Diagnostics.AddError("Error importing agreement", err.Error())
			return
		}
		state.DesiredState = types.StringValue(inferred)
	} else {
		state.DesiredState = desiredState
	}

	if allowPending.IsNull() || allowPending.IsUnknown() {
		state.AllowPending = types.BoolValue(false)
	} else {
		state.AllowPending = allowPending
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *agreementResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan models.AgreementResourceModel
	var state models.AgreementResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := uuid.MustParse(state.ID.ValueString())
	currentState := state.State.ValueString()
	desiredState := plan.DesiredState.ValueString()
	descriptorChanged := plan.DescriptorID.ValueString() != state.DescriptorID.ValueString()

	// If descriptor changed and state is ACTIVE or SUSPENDED, do an upgrade
	if descriptorChanged && (currentState == "ACTIVE" || currentState == "SUSPENDED") {
		upgraded, err := r.client.UpgradeAgreement(ctx, id)
		if err != nil {
			resp.Diagnostics.AddError("Error upgrading agreement", err.Error())
			return
		}
		populateModelFromAgreement(&plan, upgraded)
		plan.DesiredState = types.StringValue(desiredState)
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	transitions, err := ComputeTransitions(currentState, desiredState, descriptorChanged)
	if err != nil {
		resp.Diagnostics.AddError("Error computing transitions", err.Error())
		return
	}

	for _, t := range transitions {
		switch t.Type {
		case TransitionSubmit:
			submission := api.AgreementSubmission{}
			if !plan.ConsumerNotes.IsNull() && !plan.ConsumerNotes.IsUnknown() {
				notes := plan.ConsumerNotes.ValueString()
				submission.ConsumerNotes = &notes
			}
			_, err = r.client.SubmitAgreement(ctx, id, submission)
		case TransitionSuspend:
			_, err = r.client.SuspendAgreement(ctx, id, nil)
		case TransitionUnsuspend:
			_, err = r.client.UnsuspendAgreement(ctx, id, nil)
		case TransitionUpgrade:
			_, err = r.client.UpgradeAgreement(ctx, id)
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
	agreement, err := r.client.GetAgreement(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading agreement after update", err.Error())
		return
	}

	populateModelFromAgreement(&plan, agreement)
	plan.DesiredState = types.StringValue(desiredState)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *agreementResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.AgreementResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	currentState := state.State.ValueString()
	id := uuid.MustParse(state.ID.ValueString())

	if canDelete(currentState) {
		err := r.client.DeleteAgreement(ctx, id)
		if err != nil {
			if client.IsNotFound(err) {
				return
			}
			resp.Diagnostics.AddError("Error deleting agreement", err.Error())
		}
		return
	}

	resp.Diagnostics.AddError(
		"Cannot Delete Agreement",
		fmt.Sprintf(
			"Cannot delete agreement %s in state %s. Only DRAFT, PENDING, and MISSING_CERTIFIED_ATTRIBUTES agreements can be deleted.",
			id, currentState,
		),
	)
}
