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
	"github.com/lmammino/terraform-provider-pdnd/internal/client"
	"github.com/lmammino/terraform-provider-pdnd/internal/client/api"
	"github.com/lmammino/terraform-provider-pdnd/internal/models"
	"github.com/lmammino/terraform-provider-pdnd/internal/providerdata"
)

var _ resource.Resource = &descriptorDocumentResource{}
var _ resource.ResourceWithImportState = &descriptorDocumentResource{}

type descriptorDocumentResource struct {
	client api.DescriptorDocumentsAPI
}

func NewDescriptorDocumentResource() resource.Resource {
	return &descriptorDocumentResource{}
}

func (r *descriptorDocumentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_eservice_descriptor_document"
}

func (r *descriptorDocumentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a document on a PDND e-service descriptor.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite ID: eservice_id/descriptor_id/document_id",
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
				Description: "Human-readable name for the document",
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
				Description: "Document creation timestamp",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *descriptorDocumentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *descriptorDocumentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.DescriptorDocumentResourceModel
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

	doc, err := r.client.UploadDocument(ctx, eserviceID, descriptorID, fileName, fileContent, plan.PrettyName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error uploading document", err.Error())
		return
	}

	plan.DocumentID = types.StringValue(doc.ID.String())
	plan.Name = types.StringValue(doc.Name)
	plan.CreatedAt = types.StringValue(doc.CreatedAt.String())
	plan.FileHash = types.StringValue(computedHash)
	plan.ID = types.StringValue(fmt.Sprintf("%s/%s/%s", eserviceID, descriptorID, doc.ID))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *descriptorDocumentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.DescriptorDocumentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	eserviceID := uuid.MustParse(state.EServiceID.ValueString())
	descriptorID := uuid.MustParse(state.DescriptorID.ValueString())
	documentID := uuid.MustParse(state.DocumentID.ValueString())

	doc, err := r.client.GetDocumentByID(ctx, eserviceID, descriptorID, documentID)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading document", err.Error())
		return
	}

	// Update server-known fields, preserve file_path and file_hash from state
	state.DocumentID = types.StringValue(doc.ID.String())
	state.Name = types.StringValue(doc.Name)
	state.PrettyName = types.StringValue(doc.PrettyName)
	state.ContentType = types.StringValue(doc.ContentType)
	state.CreatedAt = types.StringValue(doc.CreatedAt.String())
	state.ID = types.StringValue(fmt.Sprintf("%s/%s/%s", eserviceID, descriptorID, doc.ID))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *descriptorDocumentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan models.DescriptorDocumentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// All user-facing attributes have RequiresReplace, so Update is only called
	// when file_hash changes from user-provided to computed or vice versa.
	// Recompute the hash if needed.
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

func (r *descriptorDocumentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.DescriptorDocumentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	eserviceID := uuid.MustParse(state.EServiceID.ValueString())
	descriptorID := uuid.MustParse(state.DescriptorID.ValueString())
	documentID := uuid.MustParse(state.DocumentID.ValueString())

	err := r.client.DeleteDocument(ctx, eserviceID, descriptorID, documentID)
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting document", err.Error())
	}
}

func (r *descriptorDocumentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	eserviceID, descriptorID, documentID, err := parseDocumentCompositeID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("eservice_id"), types.StringValue(eserviceID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("descriptor_id"), types.StringValue(descriptorID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("document_id"), types.StringValue(documentID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(eserviceID+"/"+descriptorID+"/"+documentID))...)
}
