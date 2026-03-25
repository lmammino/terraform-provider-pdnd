package resources

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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

// descriptorAttributeGroupObjectType is the attr.Type for a group nested attribute.
var descriptorAttributeGroupObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"attribute_ids": types.ListType{ElemType: types.StringType},
	},
}

func descriptorAttributesSchema(attrType string) schema.Schema {
	return schema.Schema{
		Description: fmt.Sprintf("Manages %s attribute groups on a PDND e-service descriptor.", attrType),
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
		},
		Blocks: map[string]schema.Block{
			"group": schema.ListNestedBlock{
				Description: "Attribute groups. Attributes within a group are OR'd; groups are AND'd.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"attribute_ids": schema.ListAttribute{
							Description: "List of attribute UUIDs in this group",
							Required:    true,
							ElementType: types.StringType,
							Validators: []validator.List{
								listvalidator.SizeAtLeast(1),
								listvalidator.ValueStringsAre(
									stringvalidator.RegexMatches(uuidRegex, "must be a valid UUID"),
								),
							},
						},
					},
				},
			},
		},
	}
}

func descriptorAttributesConfigure(req resource.ConfigureRequest, resp *resource.ConfigureResponse) api.DescriptorAttributesAPI {
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

	return pd.DescriptorAttributesAPI
}

func descriptorAttributesCreate(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse, client api.DescriptorAttributesAPI, attrType string) {
	var plan models.DescriptorAttributesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	eserviceID := uuid.MustParse(plan.EServiceID.ValueString())
	descriptorID := uuid.MustParse(plan.DescriptorID.ValueString())

	groups, diags := extractGroups(ctx, plan.Groups)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, group := range groups {
		err := client.CreateDescriptorAttributeGroup(ctx, eserviceID, descriptorID, attrType, group)
		if err != nil {
			resp.Diagnostics.AddError("Error creating attribute group", err.Error())
			return
		}
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", eserviceID, descriptorID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func descriptorAttributesRead(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse, client api.DescriptorAttributesAPI, attrType string) {
	var state models.DescriptorAttributesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	eserviceID := uuid.MustParse(state.EServiceID.ValueString())
	descriptorID := uuid.MustParse(state.DescriptorID.ValueString())

	entries, err := fetchAllDescriptorAttributes(ctx, client, eserviceID, descriptorID, attrType)
	if err != nil {
		resp.Diagnostics.AddError("Error reading descriptor attributes", err.Error())
		return
	}

	groups := reconstructGroups(entries)

	groupList, diags := groupsToTerraformList(groups)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Groups = groupList
	state.ID = types.StringValue(fmt.Sprintf("%s/%s", eserviceID, descriptorID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func descriptorAttributesUpdate(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse, client api.DescriptorAttributesAPI, attrType string) {
	var plan models.DescriptorAttributesResourceModel
	var state models.DescriptorAttributesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	eserviceID := uuid.MustParse(state.EServiceID.ValueString())
	descriptorID := uuid.MustParse(state.DescriptorID.ValueString())

	// Delete all existing attributes
	existing, err := fetchAllDescriptorAttributes(ctx, client, eserviceID, descriptorID, attrType)
	if err != nil {
		resp.Diagnostics.AddError("Error reading existing attributes", err.Error())
		return
	}

	for _, entry := range existing {
		err := client.DeleteDescriptorAttributeFromGroup(ctx, eserviceID, descriptorID, attrType, entry.GroupIndex, entry.AttributeID)
		if err != nil {
			resp.Diagnostics.AddError("Error deleting attribute from group", err.Error())
			return
		}
	}

	// Recreate all groups from plan
	groups, diags := extractGroups(ctx, plan.Groups)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, group := range groups {
		err := client.CreateDescriptorAttributeGroup(ctx, eserviceID, descriptorID, attrType, group)
		if err != nil {
			resp.Diagnostics.AddError("Error creating attribute group", err.Error())
			return
		}
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", eserviceID, descriptorID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func descriptorAttributesDelete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse, client api.DescriptorAttributesAPI, attrType string) {
	var state models.DescriptorAttributesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	eserviceID := uuid.MustParse(state.EServiceID.ValueString())
	descriptorID := uuid.MustParse(state.DescriptorID.ValueString())

	existing, err := fetchAllDescriptorAttributes(ctx, client, eserviceID, descriptorID, attrType)
	if err != nil {
		resp.Diagnostics.AddError("Error reading existing attributes", err.Error())
		return
	}

	for _, entry := range existing {
		err := client.DeleteDescriptorAttributeFromGroup(ctx, eserviceID, descriptorID, attrType, entry.GroupIndex, entry.AttributeID)
		if err != nil {
			resp.Diagnostics.AddError("Error deleting attribute from group", err.Error())
			return
		}
	}
}

func descriptorAttributesImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	eserviceID, descriptorID, err := parseDescriptorCompositeID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("eservice_id"), types.StringValue(eserviceID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("descriptor_id"), types.StringValue(descriptorID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(eserviceID+"/"+descriptorID))...)
}

// fetchAllDescriptorAttributes fetches all attribute entries with pagination.
func fetchAllDescriptorAttributes(ctx context.Context, client api.DescriptorAttributesAPI, eserviceID, descriptorID uuid.UUID, attrType string) ([]api.DescriptorAttributeEntry, error) {
	var all []api.DescriptorAttributeEntry
	var offset int32
	const limit int32 = 50

	for {
		entries, pagination, err := client.ListDescriptorAttributes(ctx, eserviceID, descriptorID, attrType, offset, limit)
		if err != nil {
			return nil, err
		}
		all = append(all, entries...)
		if int32(len(all)) >= pagination.TotalCount {
			break
		}
		offset += limit
	}

	return all, nil
}

// reconstructGroups converts a flat list of entries into ordered groups.
func reconstructGroups(entries []api.DescriptorAttributeEntry) [][]uuid.UUID {
	if len(entries) == 0 {
		return nil
	}

	grouped := make(map[int32][]uuid.UUID)
	for _, e := range entries {
		grouped[e.GroupIndex] = append(grouped[e.GroupIndex], e.AttributeID)
	}

	// Sort by group index
	indices := make([]int32, 0, len(grouped))
	for idx := range grouped {
		indices = append(indices, idx)
	}
	sort.Slice(indices, func(i, j int) bool { return indices[i] < indices[j] })

	result := make([][]uuid.UUID, len(indices))
	for i, idx := range indices {
		result[i] = grouped[idx]
	}
	return result
}

// extractGroups reads groups from the Terraform plan list.
func extractGroups(ctx context.Context, groupsList types.List) ([][]uuid.UUID, diag.Diagnostics) {
	var groupModels []models.DescriptorAttributeGroupModel
	diags := groupsList.ElementsAs(ctx, &groupModels, false)
	if diags.HasError() {
		return nil, diags
	}

	groups := make([][]uuid.UUID, len(groupModels))
	for i, gm := range groupModels {
		var ids []string
		d := gm.AttributeIDs.ElementsAs(ctx, &ids, false)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		uuids := make([]uuid.UUID, len(ids))
		for j, idStr := range ids {
			uuids[j] = uuid.MustParse(idStr)
		}
		groups[i] = uuids
	}

	return groups, diags
}

// groupsToTerraformList converts groups from [][]uuid.UUID to a types.List for state.
func groupsToTerraformList(groups [][]uuid.UUID) (types.List, diag.Diagnostics) {
	if len(groups) == 0 {
		return types.ListValueMust(descriptorAttributeGroupObjectType, []attr.Value{}), nil
	}

	groupValues := make([]attr.Value, len(groups))
	for i, group := range groups {
		idStrings := make([]attr.Value, len(group))
		for j, id := range group {
			idStrings[j] = types.StringValue(id.String())
		}

		idList, diags := types.ListValue(types.StringType, idStrings)
		if diags.HasError() {
			return types.ListNull(descriptorAttributeGroupObjectType), diags
		}

		obj, diags := types.ObjectValue(
			descriptorAttributeGroupObjectType.AttrTypes,
			map[string]attr.Value{
				"attribute_ids": idList,
			},
		)
		if diags.HasError() {
			return types.ListNull(descriptorAttributeGroupObjectType), diags
		}

		groupValues[i] = obj
	}

	return types.ListValue(descriptorAttributeGroupObjectType, groupValues)
}
