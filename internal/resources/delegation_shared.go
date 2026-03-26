package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/lmammino/terraform-provider-pdnd/internal/client"
	"github.com/lmammino/terraform-provider-pdnd/internal/client/api"
	"github.com/lmammino/terraform-provider-pdnd/internal/models"
	"github.com/lmammino/terraform-provider-pdnd/internal/providerdata"
)

func delegationSchema(delegationType string) schema.Schema {
	return schema.Schema{
		Description: fmt.Sprintf("Manages a PDND %s delegation.", delegationType),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Delegation UUID",
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
			"delegate_id": schema.StringAttribute{
				Description: "Delegate UUID",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(uuidRegex, "must be a valid UUID"),
				},
			},
			"delegator_id": schema.StringAttribute{
				Description: "Delegator UUID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"state": schema.StringAttribute{
				Description: "Delegation state",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Creation timestamp",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"submitted_at": schema.StringAttribute{
				Description: "Submission timestamp",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "Last update timestamp",
				Computed:    true,
			},
			"activated_at": schema.StringAttribute{
				Description: "Activation timestamp",
				Computed:    true,
			},
			"rejected_at": schema.StringAttribute{
				Description: "Rejection timestamp",
				Computed:    true,
			},
			"revoked_at": schema.StringAttribute{
				Description: "Revocation timestamp",
				Computed:    true,
			},
			"rejection_reason": schema.StringAttribute{
				Description: "Reason for rejection",
				Computed:    true,
			},
		},
	}
}

func delegationConfigure(req resource.ConfigureRequest, resp *resource.ConfigureResponse) api.DelegationsAPI {
	if req.ProviderData == nil {
		return nil
	}

	pd, ok := req.ProviderData.(*providerdata.ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data",
			fmt.Sprintf("Expected *providerdata.ProviderData, got: %T", req.ProviderData),
		)
		return nil
	}

	return pd.DelegationsAPI
}

func delegationCreate(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse, c api.DelegationsAPI, delegationType string) {
	var plan models.DelegationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	seed := api.DelegationSeed{
		EServiceID: uuid.MustParse(plan.EServiceID.ValueString()),
		DelegateID: uuid.MustParse(plan.DelegateID.ValueString()),
	}

	delegation, err := c.CreateDelegation(ctx, delegationType, seed)
	if err != nil {
		resp.Diagnostics.AddError("Error creating delegation", err.Error())
		return
	}

	populateDelegationModel(&plan, delegation)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func delegationRead(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse, c api.DelegationsAPI, delegationType string) {
	var state models.DelegationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := uuid.MustParse(state.ID.ValueString())
	delegation, err := c.GetDelegation(ctx, delegationType, id)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading delegation", err.Error())
		return
	}

	populateDelegationModel(&state, delegation)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func delegationDelete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse, _ api.DelegationsAPI, _ string) {
	resp.Diagnostics.AddWarning(
		"Delegation Not Deleted on Platform",
		"The delegation has been removed from Terraform state but continues to exist on the PDND platform. Delegations cannot be deleted via the API.",
	)
}

func delegationImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, idPath, req, resp)
}

func populateDelegationModel(model *models.DelegationResourceModel, d *api.Delegation) {
	model.ID = types.StringValue(d.ID.String())
	model.EServiceID = types.StringValue(d.EServiceID.String())
	model.DelegateID = types.StringValue(d.DelegateID.String())
	model.DelegatorID = types.StringValue(d.DelegatorID.String())
	model.State = types.StringValue(d.State)
	model.CreatedAt = types.StringValue(d.CreatedAt.Format(time.RFC3339))
	model.SubmittedAt = types.StringValue(d.SubmittedAt.Format(time.RFC3339))

	if d.UpdatedAt != nil {
		model.UpdatedAt = types.StringValue(d.UpdatedAt.Format(time.RFC3339))
	} else {
		model.UpdatedAt = types.StringNull()
	}

	if d.ActivatedAt != nil {
		model.ActivatedAt = types.StringValue(d.ActivatedAt.Format(time.RFC3339))
	} else {
		model.ActivatedAt = types.StringNull()
	}

	if d.RejectedAt != nil {
		model.RejectedAt = types.StringValue(d.RejectedAt.Format(time.RFC3339))
	} else {
		model.RejectedAt = types.StringNull()
	}

	if d.RevokedAt != nil {
		model.RevokedAt = types.StringValue(d.RevokedAt.Format(time.RFC3339))
	} else {
		model.RevokedAt = types.StringNull()
	}

	if d.RejectionReason != nil {
		model.RejectionReason = types.StringValue(*d.RejectionReason)
	} else {
		model.RejectionReason = types.StringNull()
	}
}
