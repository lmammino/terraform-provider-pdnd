package resources

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/lmammino/terraform-provider-pdnd/internal/client/api"
	"github.com/lmammino/terraform-provider-pdnd/internal/models"
	"github.com/lmammino/terraform-provider-pdnd/internal/providerdata"
)

var _ resource.Resource = &descriptorInterfaceResource{}
var _ resource.ResourceWithImportState = &descriptorInterfaceResource{}

type descriptorInterfaceResource struct {
	client api.DescriptorDocumentsAPI
}

func NewDescriptorInterfaceResource() resource.Resource {
	return &descriptorInterfaceResource{}
}

func (r *descriptorInterfaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_eservice_descriptor_interface"
}

func (r *descriptorInterfaceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the interface on a PDND e-service descriptor.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite ID: eservice_id/descriptor_id",
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(uuidRegex, "must be a valid UUID"),
				},
			},
			"pretty_name": schema.StringAttribute{
				Description: "Human-readable name for the interface",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"file_path": schema.StringAttribute{
				Description: "Path to the file to upload",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"content_type": schema.StringAttribute{
				Description: "MIME content type of the file",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"file_hash": schema.StringAttribute{
				Description: "SHA256 hash of the file content for change detection",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"document_id": schema.StringAttribute{
				Description: "Server-assigned document UUID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Filename derived from file_path",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "Interface creation timestamp",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *descriptorInterfaceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = pd.DescriptorDocumentsAPI
}

func (r *descriptorInterfaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.DescriptorInterfaceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	eserviceID := uuid.MustParse(plan.EServiceID.ValueString())
	descriptorID := uuid.MustParse(plan.DescriptorID.ValueString())
	filePath := plan.FilePath.ValueString()

	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		resp.Diagnostics.AddError("Error reading file", fmt.Sprintf("Could not read file %s: %s", filePath, err))
		return
	}

	fileName := filepath.Base(filePath)
	hashBytes := sha256.Sum256(fileContent)
	computedHash := hex.EncodeToString(hashBytes[:])

	doc, err := r.client.UploadInterface(ctx, eserviceID, descriptorID, fileName, fileContent, plan.PrettyName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error uploading interface", err.Error())
		return
	}

	plan.DocumentID = types.StringValue(doc.ID.String())
	plan.Name = types.StringValue(doc.Name)
	plan.CreatedAt = types.StringValue(doc.CreatedAt.String())
	plan.FileHash = types.StringValue(computedHash)
	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", eserviceID, descriptorID))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *descriptorInterfaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.DescriptorInterfaceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	eserviceID := uuid.MustParse(state.EServiceID.ValueString())
	descriptorID := uuid.MustParse(state.DescriptorID.ValueString())

	exists, err := r.client.CheckInterfaceExists(ctx, eserviceID, descriptorID)
	if err != nil {
		resp.Diagnostics.AddError("Error checking interface", err.Error())
		return
	}

	if !exists {
		resp.State.RemoveResource(ctx)
		return
	}

	// Keep all state fields as-is (they're all immutable)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *descriptorInterfaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan models.DescriptorInterfaceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// All user-facing attributes have RequiresReplace, so Update is only called
	// when file_hash changes from user-provided to computed or vice versa.
	if plan.FileHash.IsUnknown() {
		filePath := plan.FilePath.ValueString()
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			resp.Diagnostics.AddError("Error reading file", fmt.Sprintf("Could not read file %s: %s", filePath, err))
			return
		}
		hashBytes := sha256.Sum256(fileContent)
		plan.FileHash = types.StringValue(hex.EncodeToString(hashBytes[:]))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *descriptorInterfaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.DescriptorInterfaceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	eserviceID := uuid.MustParse(state.EServiceID.ValueString())
	descriptorID := uuid.MustParse(state.DescriptorID.ValueString())

	err := r.client.DeleteInterface(ctx, eserviceID, descriptorID)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting interface", err.Error())
	}
}

func (r *descriptorInterfaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	eserviceID, descriptorID, err := parseDescriptorCompositeID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("eservice_id"), types.StringValue(eserviceID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("descriptor_id"), types.StringValue(descriptorID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(eserviceID+"/"+descriptorID))...)
}
