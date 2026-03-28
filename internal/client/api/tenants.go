package api

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	client "github.com/lmammino/terraform-provider-pdnd/internal/client"
	generated "github.com/lmammino/terraform-provider-pdnd/internal/client/generated"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// TenantsAPI defines operations on PDND tenants.
type TenantsAPI interface {
	// Tenant reads
	GetTenant(ctx context.Context, id uuid.UUID) (*TenantInfo, error)
	ListTenants(ctx context.Context, params ListTenantsParams) (*TenantsPage, error)

	// Certified attributes
	ListTenantCertifiedAttributes(ctx context.Context, tenantID uuid.UUID, offset, limit int32) (*TenantCertifiedAttrsPage, error)
	AssignTenantCertifiedAttribute(ctx context.Context, tenantID, attributeID uuid.UUID) (*TenantCertifiedAttr, error)
	RevokeTenantCertifiedAttribute(ctx context.Context, tenantID, attributeID uuid.UUID) (*TenantCertifiedAttr, error)

	// Declared attributes
	ListTenantDeclaredAttributes(ctx context.Context, tenantID uuid.UUID, offset, limit int32) (*TenantDeclaredAttrsPage, error)
	AssignTenantDeclaredAttribute(ctx context.Context, tenantID, attributeID uuid.UUID, delegationID *uuid.UUID) (*TenantDeclaredAttr, error)
	RevokeTenantDeclaredAttribute(ctx context.Context, tenantID, attributeID uuid.UUID) (*TenantDeclaredAttr, error)

	// Verified attributes
	ListTenantVerifiedAttributes(ctx context.Context, tenantID uuid.UUID, offset, limit int32) (*TenantVerifiedAttrsPage, error)
	AssignTenantVerifiedAttribute(ctx context.Context, tenantID, attributeID, agreementID uuid.UUID, expirationDate *time.Time) (*TenantVerifiedAttr, error)
	RevokeTenantVerifiedAttribute(ctx context.Context, tenantID, attributeID, agreementID uuid.UUID) (*TenantVerifiedAttr, error)
}

type tenantsClient struct {
	client *generated.ClientWithResponses
}

// NewTenantsClient creates a new TenantsAPI backed by the generated client.
func NewTenantsClient(c *generated.ClientWithResponses) TenantsAPI {
	return &tenantsClient{client: c}
}

func (tc *tenantsClient) GetTenant(ctx context.Context, id uuid.UUID) (*TenantInfo, error) {
	resp, err := tc.client.GetTenantWithResponse(ctx, openapi_types.UUID(id))
	if err != nil {
		return nil, fmt.Errorf("get tenant: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("get tenant: empty response body")
	}
	return tenantInfoFromGenerated(resp.JSON200), nil
}

func (tc *tenantsClient) ListTenants(ctx context.Context, params ListTenantsParams) (*TenantsPage, error) {
	genParams := &generated.GetTenantsParams{
		Offset:  params.Offset,
		Limit:   params.Limit,
		IPACode: params.IPACode,
		TaxCode: params.TaxCode,
	}

	resp, err := tc.client.GetTenantsWithResponse(ctx, genParams)
	if err != nil {
		return nil, fmt.Errorf("list tenants: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("list tenants: empty response body")
	}

	page := &TenantsPage{
		Pagination: paginationFromGenerated(resp.JSON200.Pagination),
		Results:    make([]TenantInfo, 0, len(resp.JSON200.Results)),
	}
	for i := range resp.JSON200.Results {
		info := tenantInfoFromGenerated(&resp.JSON200.Results[i])
		page.Results = append(page.Results, *info)
	}
	return page, nil
}

func (tc *tenantsClient) ListTenantCertifiedAttributes(ctx context.Context, tenantID uuid.UUID, offset, limit int32) (*TenantCertifiedAttrsPage, error) {
	genParams := &generated.GetTenantCertifiedAttributesParams{
		Offset: offset,
		Limit:  limit,
	}
	resp, err := tc.client.GetTenantCertifiedAttributesWithResponse(ctx, openapi_types.UUID(tenantID), genParams)
	if err != nil {
		return nil, fmt.Errorf("list tenant certified attributes: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("list tenant certified attributes: empty response body")
	}

	page := &TenantCertifiedAttrsPage{
		Pagination: paginationFromGenerated(resp.JSON200.Pagination),
		Results:    make([]TenantCertifiedAttr, len(resp.JSON200.Results)),
	}
	for i := range resp.JSON200.Results {
		page.Results[i] = *tenantCertifiedAttrFromGenerated(&resp.JSON200.Results[i])
	}
	return page, nil
}

func (tc *tenantsClient) AssignTenantCertifiedAttribute(ctx context.Context, tenantID, attributeID uuid.UUID) (*TenantCertifiedAttr, error) {
	body := generated.TenantCertifiedAttributeSeed{
		Id: openapi_types.UUID(attributeID),
	}
	resp, err := tc.client.AssignTenantCertifiedAttributeWithResponse(ctx, openapi_types.UUID(tenantID), body)
	if err != nil {
		return nil, fmt.Errorf("assign tenant certified attribute: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("assign tenant certified attribute: empty response body")
	}
	return tenantCertifiedAttrFromGenerated(resp.JSON200), nil
}

func (tc *tenantsClient) RevokeTenantCertifiedAttribute(ctx context.Context, tenantID, attributeID uuid.UUID) (*TenantCertifiedAttr, error) {
	resp, err := tc.client.RevokeTenantCertifiedAttributeWithResponse(ctx, openapi_types.UUID(tenantID), openapi_types.UUID(attributeID))
	if err != nil {
		return nil, fmt.Errorf("revoke tenant certified attribute: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("revoke tenant certified attribute: empty response body")
	}
	return tenantCertifiedAttrFromGenerated(resp.JSON200), nil
}

func (tc *tenantsClient) ListTenantDeclaredAttributes(ctx context.Context, tenantID uuid.UUID, offset, limit int32) (*TenantDeclaredAttrsPage, error) {
	genParams := &generated.GetTenantDeclaredAttributesParams{
		Offset: offset,
		Limit:  limit,
	}
	resp, err := tc.client.GetTenantDeclaredAttributesWithResponse(ctx, openapi_types.UUID(tenantID), genParams)
	if err != nil {
		return nil, fmt.Errorf("list tenant declared attributes: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("list tenant declared attributes: empty response body")
	}

	page := &TenantDeclaredAttrsPage{
		Pagination: paginationFromGenerated(resp.JSON200.Pagination),
		Results:    make([]TenantDeclaredAttr, len(resp.JSON200.Results)),
	}
	for i := range resp.JSON200.Results {
		page.Results[i] = *tenantDeclaredAttrFromGenerated(&resp.JSON200.Results[i])
	}
	return page, nil
}

func (tc *tenantsClient) AssignTenantDeclaredAttribute(ctx context.Context, tenantID, attributeID uuid.UUID, delegationID *uuid.UUID) (*TenantDeclaredAttr, error) {
	body := generated.TenantDeclaredAttributeSeed{
		Id: openapi_types.UUID(attributeID),
	}
	if delegationID != nil {
		id := openapi_types.UUID(*delegationID)
		body.DelegationId = &id
	}
	resp, err := tc.client.AssignTenantDeclaredAttributeWithResponse(ctx, openapi_types.UUID(tenantID), body)
	if err != nil {
		return nil, fmt.Errorf("assign tenant declared attribute: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("assign tenant declared attribute: empty response body")
	}
	return tenantDeclaredAttrFromGenerated(resp.JSON200), nil
}

func (tc *tenantsClient) RevokeTenantDeclaredAttribute(ctx context.Context, tenantID, attributeID uuid.UUID) (*TenantDeclaredAttr, error) {
	resp, err := tc.client.RevokeTenantDeclaredAttributeWithResponse(ctx, openapi_types.UUID(tenantID), openapi_types.UUID(attributeID))
	if err != nil {
		return nil, fmt.Errorf("revoke tenant declared attribute: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("revoke tenant declared attribute: empty response body")
	}
	return tenantDeclaredAttrFromGenerated(resp.JSON200), nil
}

func (tc *tenantsClient) ListTenantVerifiedAttributes(ctx context.Context, tenantID uuid.UUID, offset, limit int32) (*TenantVerifiedAttrsPage, error) {
	genParams := &generated.GetTenantVerifiedAttributesParams{
		Offset: offset,
		Limit:  limit,
	}
	resp, err := tc.client.GetTenantVerifiedAttributesWithResponse(ctx, openapi_types.UUID(tenantID), genParams)
	if err != nil {
		return nil, fmt.Errorf("list tenant verified attributes: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("list tenant verified attributes: empty response body")
	}

	page := &TenantVerifiedAttrsPage{
		Pagination: paginationFromGenerated(resp.JSON200.Pagination),
		Results:    make([]TenantVerifiedAttr, len(resp.JSON200.Results)),
	}
	for i := range resp.JSON200.Results {
		page.Results[i] = *tenantVerifiedAttrFromGenerated(&resp.JSON200.Results[i])
	}
	return page, nil
}

func (tc *tenantsClient) AssignTenantVerifiedAttribute(ctx context.Context, tenantID, attributeID, agreementID uuid.UUID, expirationDate *time.Time) (*TenantVerifiedAttr, error) {
	body := generated.TenantVerifiedAttributeSeed{
		Id:             openapi_types.UUID(attributeID),
		AgreementId:    openapi_types.UUID(agreementID),
		ExpirationDate: expirationDate,
	}
	resp, err := tc.client.AssignTenantVerifiedAttributeWithResponse(ctx, openapi_types.UUID(tenantID), body)
	if err != nil {
		return nil, fmt.Errorf("assign tenant verified attribute: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("assign tenant verified attribute: empty response body")
	}
	return tenantVerifiedAttrFromGenerated(resp.JSON200), nil
}

func (tc *tenantsClient) RevokeTenantVerifiedAttribute(ctx context.Context, tenantID, attributeID, agreementID uuid.UUID) (*TenantVerifiedAttr, error) {
	params := &generated.RevokeTenantVerifiedAttributeParams{
		AgreementId: openapi_types.UUID(agreementID),
	}
	resp, err := tc.client.RevokeTenantVerifiedAttributeWithResponse(ctx, openapi_types.UUID(tenantID), openapi_types.UUID(attributeID), params)
	if err != nil {
		return nil, fmt.Errorf("revoke tenant verified attribute: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("revoke tenant verified attribute: empty response body")
	}
	return tenantVerifiedAttrFromGenerated(resp.JSON200), nil
}
