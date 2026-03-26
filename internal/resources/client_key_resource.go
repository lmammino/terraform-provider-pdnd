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

var _ resource.Resource = &clientKeyResource{}
var _ resource.ResourceWithImportState = &clientKeyResource{}

type clientKeyResource struct {
	client api.ClientsAPI
}

func NewClientKeyResource() resource.Resource {
	return &clientKeyResource{}
}

func (r *clientKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_client_key"
}

func (r *clientKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a public key on a PDND client.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite ID: client_id/kid",
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
			"key": schema.StringAttribute{
				Description: "Base64 PEM public key",
				Required:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"use": schema.StringAttribute{
				Description: "Key usage: SIG or ENC",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("SIG", "ENC"),
				},
			},
			"alg": schema.StringAttribute{
				Description: "Key algorithm",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Key name (5-60 characters)",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(5, 60),
				},
			},
			"kid": schema.StringAttribute{
				Description: "Server-assigned key ID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"kty": schema.StringAttribute{
				Description: "Key type (RSA, EC, etc.)",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *clientKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *clientKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.ClientKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clientID := uuid.MustParse(plan.ClientID.ValueString())

	seed := api.ClientKeySeed{
		Key:  plan.Key.ValueString(),
		Use:  plan.Use.ValueString(),
		Alg:  plan.Alg.ValueString(),
		Name: plan.Name.ValueString(),
	}

	detail, err := r.client.CreateClientKey(ctx, clientID, seed)
	if err != nil {
		resp.Diagnostics.AddError("Error creating client key", err.Error())
		return
	}

	plan.Kid = types.StringValue(detail.Key.Kid)
	plan.Kty = types.StringValue(detail.Key.Kty)
	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", clientID, detail.Key.Kid))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *clientKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.ClientKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clientID := uuid.MustParse(state.ClientID.ValueString())
	kid := state.Kid.ValueString()

	// Paginate through keys to find the one matching kid
	found := false
	var offset int32
	const limit int32 = 50

	for {
		page, err := r.client.ListClientKeys(ctx, clientID, offset, limit)
		if err != nil {
			if client.IsNotFound(err) {
				resp.State.RemoveResource(ctx)
				return
			}
			resp.Diagnostics.AddError("Error reading client keys", err.Error())
			return
		}

		for _, k := range page.Results {
			if k.Kid == kid {
				found = true
				state.Kty = types.StringValue(k.Kty)
				break
			}
		}

		if found || int32(len(page.Results)) < limit {
			break
		}
		offset += limit
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	// Preserve key, use, alg, name from state (not returned by list)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *clientKeyResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// No-op: all user fields have RequiresReplace
}

func (r *clientKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.ClientKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clientID := uuid.MustParse(state.ClientID.ValueString())
	kid := state.Kid.ValueString()

	err := r.client.DeleteClientKey(ctx, clientID, kid)
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting client key", err.Error())
	}
}

func (r *clientKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	clientID, kid, err := parseClientKeyCompositeID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("client_id"), types.StringValue(clientID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("kid"), types.StringValue(kid))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(clientID+"/"+kid))...)
}

// parseClientKeyCompositeID parses a composite import ID of the form "client_id/kid".
func parseClientKeyCompositeID(id string) (string, string, error) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 || parts[1] == "" {
		return "", "", fmt.Errorf("expected import ID format: client_id/kid, got: %s", id)
	}

	clientID := parts[0]
	kid := parts[1]

	if _, err := uuid.Parse(clientID); err != nil {
		return "", "", fmt.Errorf("invalid client_id UUID: %s", clientID)
	}

	return clientID, kid, nil
}
