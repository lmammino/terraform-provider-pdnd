package api

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	client "github.com/lmammino/terraform-provider-pdnd/internal/client"
	generated "github.com/lmammino/terraform-provider-pdnd/internal/client/generated"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// DescriptorAttributesAPI defines operations on descriptor attribute groups.
// The attrType parameter must be "certified", "declared", or "verified".
type DescriptorAttributesAPI interface {
	ListDescriptorAttributes(ctx context.Context, eserviceID, descriptorID uuid.UUID, attrType string, offset, limit int32) ([]DescriptorAttributeEntry, Pagination, error)
	CreateDescriptorAttributeGroup(ctx context.Context, eserviceID, descriptorID uuid.UUID, attrType string, attributeIDs []uuid.UUID) error
	DeleteDescriptorAttributeFromGroup(ctx context.Context, eserviceID, descriptorID uuid.UUID, attrType string, groupIndex int32, attributeID uuid.UUID) error
}

type descriptorAttributesClient struct {
	client *generated.ClientWithResponses
}

// NewDescriptorAttributesClient creates a new DescriptorAttributesAPI backed by the generated client.
func NewDescriptorAttributesClient(c *generated.ClientWithResponses) DescriptorAttributesAPI {
	return &descriptorAttributesClient{client: c}
}

func (d *descriptorAttributesClient) ListDescriptorAttributes(ctx context.Context, eserviceID, descriptorID uuid.UUID, attrType string, offset, limit int32) ([]DescriptorAttributeEntry, Pagination, error) {
	esID := openapi_types.UUID(eserviceID)
	descID := openapi_types.UUID(descriptorID)

	switch attrType {
	case "certified":
		params := &generated.GetEServiceDescriptorCertifiedAttributesParams{Offset: offset, Limit: limit}
		resp, err := d.client.GetEServiceDescriptorCertifiedAttributesWithResponse(ctx, esID, descID, params)
		if err != nil {
			return nil, Pagination{}, fmt.Errorf("list descriptor certified attributes: %w", err)
		}
		if err := client.CheckResponse(resp.HTTPResponse); err != nil {
			return nil, Pagination{}, err
		}
		if resp.JSON200 == nil {
			return nil, Pagination{}, fmt.Errorf("list descriptor certified attributes: empty response body")
		}
		entries := make([]DescriptorAttributeEntry, len(resp.JSON200.Results))
		for i, r := range resp.JSON200.Results {
			entries[i] = DescriptorAttributeEntry{
				AttributeID: uuid.UUID(r.Attribute.Id),
				GroupIndex:  r.GroupIndex,
			}
		}
		return entries, paginationFromGenerated(resp.JSON200.Pagination), nil

	case "declared":
		params := &generated.GetEServiceDescriptorDeclaredAttributesParams{Offset: offset, Limit: limit}
		resp, err := d.client.GetEServiceDescriptorDeclaredAttributesWithResponse(ctx, esID, descID, params)
		if err != nil {
			return nil, Pagination{}, fmt.Errorf("list descriptor declared attributes: %w", err)
		}
		if err := client.CheckResponse(resp.HTTPResponse); err != nil {
			return nil, Pagination{}, err
		}
		if resp.JSON200 == nil {
			return nil, Pagination{}, fmt.Errorf("list descriptor declared attributes: empty response body")
		}
		entries := make([]DescriptorAttributeEntry, len(resp.JSON200.Results))
		for i, r := range resp.JSON200.Results {
			entries[i] = DescriptorAttributeEntry{
				AttributeID: uuid.UUID(r.Attribute.Id),
				GroupIndex:  r.GroupIndex,
			}
		}
		return entries, paginationFromGenerated(resp.JSON200.Pagination), nil

	case "verified":
		params := &generated.GetEServiceDescriptorVerifiedAttributesParams{Offset: offset, Limit: limit}
		resp, err := d.client.GetEServiceDescriptorVerifiedAttributesWithResponse(ctx, esID, descID, params)
		if err != nil {
			return nil, Pagination{}, fmt.Errorf("list descriptor verified attributes: %w", err)
		}
		if err := client.CheckResponse(resp.HTTPResponse); err != nil {
			return nil, Pagination{}, err
		}
		if resp.JSON200 == nil {
			return nil, Pagination{}, fmt.Errorf("list descriptor verified attributes: empty response body")
		}
		entries := make([]DescriptorAttributeEntry, len(resp.JSON200.Results))
		for i, r := range resp.JSON200.Results {
			entries[i] = DescriptorAttributeEntry{
				AttributeID: uuid.UUID(r.Attribute.Id),
				GroupIndex:  r.GroupIndex,
			}
		}
		return entries, paginationFromGenerated(resp.JSON200.Pagination), nil

	default:
		return nil, Pagination{}, fmt.Errorf("unknown attribute type: %s", attrType)
	}
}

func (d *descriptorAttributesClient) CreateDescriptorAttributeGroup(ctx context.Context, eserviceID, descriptorID uuid.UUID, attrType string, attributeIDs []uuid.UUID) error {
	esID := openapi_types.UUID(eserviceID)
	descID := openapi_types.UUID(descriptorID)
	body := generated.EServiceDescriptorAttributesGroupSeed{
		AttributeIds: uuidsToOpenAPI(attributeIDs),
	}

	switch attrType {
	case "certified":
		resp, err := d.client.CreateEServiceDescriptorCertifiedAttributesGroupWithResponse(ctx, esID, descID, body)
		if err != nil {
			return fmt.Errorf("create descriptor certified attribute group: %w", err)
		}
		return client.CheckResponse(resp.HTTPResponse)

	case "declared":
		resp, err := d.client.CreateEServiceDescriptorDeclaredAttributesGroupWithResponse(ctx, esID, descID, body)
		if err != nil {
			return fmt.Errorf("create descriptor declared attribute group: %w", err)
		}
		return client.CheckResponse(resp.HTTPResponse)

	case "verified":
		resp, err := d.client.CreateEServiceDescriptorVerifiedAttributesGroupWithResponse(ctx, esID, descID, body)
		if err != nil {
			return fmt.Errorf("create descriptor verified attribute group: %w", err)
		}
		return client.CheckResponse(resp.HTTPResponse)

	default:
		return fmt.Errorf("unknown attribute type: %s", attrType)
	}
}

func (d *descriptorAttributesClient) DeleteDescriptorAttributeFromGroup(ctx context.Context, eserviceID, descriptorID uuid.UUID, attrType string, groupIndex int32, attributeID uuid.UUID) error {
	esID := openapi_types.UUID(eserviceID)
	descID := openapi_types.UUID(descriptorID)
	attrID := openapi_types.UUID(attributeID)

	switch attrType {
	case "certified":
		resp, err := d.client.DeleteEServiceDescriptorCertifiedAttributeFromGroupWithResponse(ctx, esID, descID, groupIndex, attrID)
		if err != nil {
			return fmt.Errorf("delete descriptor certified attribute from group: %w", err)
		}
		return client.CheckResponse(resp.HTTPResponse)

	case "declared":
		resp, err := d.client.DeleteEServiceDescriptorDeclaredAttributeFromGroupWithResponse(ctx, esID, descID, groupIndex, attrID)
		if err != nil {
			return fmt.Errorf("delete descriptor declared attribute from group: %w", err)
		}
		return client.CheckResponse(resp.HTTPResponse)

	case "verified":
		resp, err := d.client.DeleteEServiceDescriptorVerifiedAttributeFromGroupWithResponse(ctx, esID, descID, groupIndex, attrID)
		if err != nil {
			return fmt.Errorf("delete descriptor verified attribute from group: %w", err)
		}
		return client.CheckResponse(resp.HTTPResponse)

	default:
		return fmt.Errorf("unknown attribute type: %s", attrType)
	}
}
