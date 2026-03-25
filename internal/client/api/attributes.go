package api

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	client "github.com/lmammino/terraform-provider-pdnd/internal/client"
	generated "github.com/lmammino/terraform-provider-pdnd/internal/client/generated"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// AttributesAPI defines operations on PDND attributes.
// This interface is the boundary between Terraform data source code and the HTTP client.
type AttributesAPI interface {
	GetCertifiedAttribute(ctx context.Context, id uuid.UUID) (*CertifiedAttribute, error)
	ListCertifiedAttributes(ctx context.Context, params PaginationParams) (*CertifiedAttributesPage, error)
	GetDeclaredAttribute(ctx context.Context, id uuid.UUID) (*DeclaredAttribute, error)
	ListDeclaredAttributes(ctx context.Context, params PaginationParams) (*DeclaredAttributesPage, error)
	GetVerifiedAttribute(ctx context.Context, id uuid.UUID) (*VerifiedAttribute, error)
	ListVerifiedAttributes(ctx context.Context, params PaginationParams) (*VerifiedAttributesPage, error)
}

// attributesClient implements AttributesAPI using the generated client.
type attributesClient struct {
	client *generated.ClientWithResponses
}

// NewAttributesClient creates a new AttributesAPI backed by the generated client.
func NewAttributesClient(c *generated.ClientWithResponses) AttributesAPI {
	return &attributesClient{client: c}
}

func (a *attributesClient) GetCertifiedAttribute(ctx context.Context, id uuid.UUID) (*CertifiedAttribute, error) {
	resp, err := a.client.GetCertifiedAttributeWithResponse(ctx, openapi_types.UUID(id))
	if err != nil {
		return nil, fmt.Errorf("get certified attribute: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("get certified attribute: empty response body")
	}
	return certifiedAttributeFromGenerated(resp.JSON200), nil
}

func (a *attributesClient) ListCertifiedAttributes(ctx context.Context, params PaginationParams) (*CertifiedAttributesPage, error) {
	genParams := &generated.GetCertifiedAttributesParams{
		Offset: params.Offset,
		Limit:  params.Limit,
	}
	resp, err := a.client.GetCertifiedAttributesWithResponse(ctx, genParams)
	if err != nil {
		return nil, fmt.Errorf("list certified attributes: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("list certified attributes: empty response body")
	}

	page := &CertifiedAttributesPage{
		Pagination: paginationFromGenerated(resp.JSON200.Pagination),
		Results:    make([]CertifiedAttribute, len(resp.JSON200.Results)),
	}
	for i := range resp.JSON200.Results {
		page.Results[i] = *certifiedAttributeFromGenerated(&resp.JSON200.Results[i])
	}
	return page, nil
}

func (a *attributesClient) GetDeclaredAttribute(ctx context.Context, id uuid.UUID) (*DeclaredAttribute, error) {
	resp, err := a.client.GetDeclaredAttributeWithResponse(ctx, openapi_types.UUID(id))
	if err != nil {
		return nil, fmt.Errorf("get declared attribute: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("get declared attribute: empty response body")
	}
	return declaredAttributeFromGenerated(resp.JSON200), nil
}

func (a *attributesClient) ListDeclaredAttributes(ctx context.Context, params PaginationParams) (*DeclaredAttributesPage, error) {
	genParams := &generated.GetDeclaredAttributesParams{
		Offset: params.Offset,
		Limit:  params.Limit,
	}
	resp, err := a.client.GetDeclaredAttributesWithResponse(ctx, genParams)
	if err != nil {
		return nil, fmt.Errorf("list declared attributes: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("list declared attributes: empty response body")
	}

	page := &DeclaredAttributesPage{
		Pagination: paginationFromGenerated(resp.JSON200.Pagination),
		Results:    make([]DeclaredAttribute, len(resp.JSON200.Results)),
	}
	for i := range resp.JSON200.Results {
		page.Results[i] = *declaredAttributeFromGenerated(&resp.JSON200.Results[i])
	}
	return page, nil
}

func (a *attributesClient) GetVerifiedAttribute(ctx context.Context, id uuid.UUID) (*VerifiedAttribute, error) {
	resp, err := a.client.GetVerifiedAttributeWithResponse(ctx, openapi_types.UUID(id))
	if err != nil {
		return nil, fmt.Errorf("get verified attribute: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("get verified attribute: empty response body")
	}
	return verifiedAttributeFromGenerated(resp.JSON200), nil
}

func (a *attributesClient) ListVerifiedAttributes(ctx context.Context, params PaginationParams) (*VerifiedAttributesPage, error) {
	genParams := &generated.GetVerifiedAttributesParams{
		Offset: params.Offset,
		Limit:  params.Limit,
	}
	resp, err := a.client.GetVerifiedAttributesWithResponse(ctx, genParams)
	if err != nil {
		return nil, fmt.Errorf("list verified attributes: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("list verified attributes: empty response body")
	}

	page := &VerifiedAttributesPage{
		Pagination: paginationFromGenerated(resp.JSON200.Pagination),
		Results:    make([]VerifiedAttribute, len(resp.JSON200.Results)),
	}
	for i := range resp.JSON200.Results {
		page.Results[i] = *verifiedAttributeFromGenerated(&resp.JSON200.Results[i])
	}
	return page, nil
}
