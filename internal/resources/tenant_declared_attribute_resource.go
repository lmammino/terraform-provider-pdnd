package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
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

var _ resource.Resource = &tenantDeclaredAttributeResource{}
var _ resource.ResourceWithImportState = &tenantDeclaredAttributeResource{}

type tenantDeclaredAttributeResource struct {
	client api.TenantsAPI
}

func NewTenantDeclaredAttributeResource() resource.Resource {
	return &tenantDeclaredAttributeResource{}
}

func (r *tenantDeclaredAttributeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant_declared_attribute"
}

func (r *tenantDeclaredAttributeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a declared attribute assignment on a PDND tenant.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite ID: tenant_id/attribute_id",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tenant_id": schema.StringAttribute{
				Description: "Tenant UUID",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(uuidRegex, "must be a valid UUID"),
				},
			},
			"attribute_id": schema.StringAttribute{
				Description: "Attribute UUID",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(uuidRegex, "must be a valid UUID"),
				},
			},
			"delegation_id": schema.StringAttribute{
				Description: "Delegation UUID (optional)",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(uuidRegex, "must be a valid UUID"),
				},
			},
			"assigned_at": schema.StringAttribute{
				Description: "Assignment timestamp",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"revoked_at": schema.StringAttribute{
				Description: "Revocation timestamp",
				Computed:    true,
			},
		},
	}
}

func (r *tenantDeclaredAttributeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = pd.TenantsAPI
}

func (r *tenantDeclaredAttributeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.TenantDeclaredAttrResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantID := uuid.MustParse(plan.TenantID.ValueString())
	attributeID := uuid.MustParse(plan.AttributeID.ValueString())

	var delegationID *uuid.UUID
	if !plan.DelegationID.IsNull() && !plan.DelegationID.IsUnknown() {
		id := uuid.MustParse(plan.DelegationID.ValueString())
		delegationID = &id
	}

	result, err := r.client.AssignTenantDeclaredAttribute(ctx, tenantID, attributeID, delegationID)
	if err != nil {
		resp.Diagnostics.AddError("Error assigning tenant declared attribute", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", tenantID, attributeID))
	plan.AssignedAt = types.StringValue(result.AssignedAt.Format(time.RFC3339))
	if result.RevokedAt != nil {
		plan.RevokedAt = types.StringValue(result.RevokedAt.Format(time.RFC3339))
	} else {
		plan.RevokedAt = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *tenantDeclaredAttributeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.TenantDeclaredAttrResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantID := uuid.MustParse(state.TenantID.ValueString())
	attributeID := uuid.MustParse(state.AttributeID.ValueString())

	// Paginate to find the matching attribute
	var found *api.TenantDeclaredAttr
	var offset int32
	const limit int32 = 50

	for {
		page, err := r.client.ListTenantDeclaredAttributes(ctx, tenantID, offset, limit)
		if err != nil {
			if client.IsNotFound(err) {
				resp.State.RemoveResource(ctx)
				return
			}
			resp.Diagnostics.AddError("Error reading tenant declared attributes", err.Error())
			return
		}

		for i := range page.Results {
			if page.Results[i].ID == attributeID {
				found = &page.Results[i]
				break
			}
		}

		if found != nil || int32(len(page.Results)) < limit {
			break
		}
		offset += limit
	}

	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// If revoked externally, treat as deleted
	if found.RevokedAt != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.AssignedAt = types.StringValue(found.AssignedAt.Format(time.RFC3339))
	state.RevokedAt = types.StringNull()
	if found.DelegationID != nil {
		state.DelegationID = types.StringValue(found.DelegationID.String())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *tenantDeclaredAttributeResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// No-op: all fields RequiresReplace
}

func (r *tenantDeclaredAttributeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.TenantDeclaredAttrResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantID := uuid.MustParse(state.TenantID.ValueString())
	attributeID := uuid.MustParse(state.AttributeID.ValueString())

	_, err := r.client.RevokeTenantDeclaredAttribute(ctx, tenantID, attributeID)
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error revoking tenant declared attribute", err.Error())
	}
}

func (r *tenantDeclaredAttributeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tenantID, attributeID, err := parseTenantAttributeCompositeID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tenant_id"), types.StringValue(tenantID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("attribute_id"), types.StringValue(attributeID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(tenantID+"/"+attributeID))...)
}
