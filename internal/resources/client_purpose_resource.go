package resources

import (
	"context"
	"fmt"
	"strings"

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

var _ resource.Resource = &clientPurposeResource{}
var _ resource.ResourceWithImportState = &clientPurposeResource{}

type clientPurposeResource struct {
	client api.ClientsAPI
}

func NewClientPurposeResource() resource.Resource {
	return &clientPurposeResource{}
}

func (r *clientPurposeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_client_purpose"
}

func (r *clientPurposeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a purpose association on a PDND client.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite ID: client_id/purpose_id",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"client_id": schema.StringAttribute{
				Description: "Client UUID",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(uuidRegex, "must be a valid UUID"),
				},
			},
			"purpose_id": schema.StringAttribute{
				Description: "Purpose UUID",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(uuidRegex, "must be a valid UUID"),
				},
			},
		},
	}
}

func (r *clientPurposeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = pd.ClientsAPI
}

func (r *clientPurposeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.ClientPurposeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clientID := uuid.MustParse(plan.ClientID.ValueString())
	purposeID := uuid.MustParse(plan.PurposeID.ValueString())

	err := r.client.AddClientPurpose(ctx, clientID, purposeID)
	if err != nil {
		resp.Diagnostics.AddError("Error adding client purpose", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", clientID, purposeID))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *clientPurposeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.ClientPurposeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Trust the state. The link is a simple association — if deleted externally,
	// the next apply will re-create the link.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *clientPurposeResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// No-op: all fields RequiresReplace
}

func (r *clientPurposeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.ClientPurposeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clientID := uuid.MustParse(state.ClientID.ValueString())
	purposeID := uuid.MustParse(state.PurposeID.ValueString())

	err := r.client.RemoveClientPurpose(ctx, clientID, purposeID)
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error removing client purpose", err.Error())
	}
}

func (r *clientPurposeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	clientID, purposeID, err := parseClientPurposeCompositeID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("client_id"), types.StringValue(clientID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("purpose_id"), types.StringValue(purposeID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(clientID+"/"+purposeID))...)
}

// parseClientPurposeCompositeID parses a composite import ID of the form "client_id/purpose_id".
func parseClientPurposeCompositeID(id string) (string, string, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("expected import ID format: client_id/purpose_id, got: %s", id)
	}

	clientID := parts[0]
	purposeID := parts[1]

	if _, err := uuid.Parse(clientID); err != nil {
		return "", "", fmt.Errorf("invalid client_id UUID: %s", clientID)
	}
	if _, err := uuid.Parse(purposeID); err != nil {
		return "", "", fmt.Errorf("invalid purpose_id UUID: %s", purposeID)
	}

	return clientID, purposeID, nil
}
