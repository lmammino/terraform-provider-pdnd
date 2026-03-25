package api

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	client "github.com/lmammino/terraform-provider-pdnd/internal/client"
	generated "github.com/lmammino/terraform-provider-pdnd/internal/client/generated"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// AgreementsAPI defines operations on PDND agreements.
// This interface is the boundary between Terraform resource code and the HTTP client.
type AgreementsAPI interface {
	CreateAgreement(ctx context.Context, seed AgreementSeed) (*Agreement, error)
	GetAgreement(ctx context.Context, id uuid.UUID) (*Agreement, error)
	ListAgreements(ctx context.Context, params ListAgreementsParams) (*AgreementsPage, error)
	DeleteAgreement(ctx context.Context, id uuid.UUID) error
	SubmitAgreement(ctx context.Context, id uuid.UUID, payload AgreementSubmission) (*Agreement, error)
	ApproveAgreement(ctx context.Context, id uuid.UUID, payload *DelegationRef) (*Agreement, error)
	RejectAgreement(ctx context.Context, id uuid.UUID, payload AgreementRejection) (*Agreement, error)
	SuspendAgreement(ctx context.Context, id uuid.UUID, payload *DelegationRef) (*Agreement, error)
	UnsuspendAgreement(ctx context.Context, id uuid.UUID, payload *DelegationRef) (*Agreement, error)
	UpgradeAgreement(ctx context.Context, id uuid.UUID) (*Agreement, error)
	CloneAgreement(ctx context.Context, id uuid.UUID) (*Agreement, error)
	ListAgreementPurposes(ctx context.Context, id uuid.UUID, params PaginationParams) (*PurposesPage, error)
}

// agreementsClient implements AgreementsAPI using the generated client.
type agreementsClient struct {
	client *generated.ClientWithResponses
}

// NewAgreementsClient creates a new AgreementsAPI backed by the generated client.
func NewAgreementsClient(c *generated.ClientWithResponses) AgreementsAPI {
	return &agreementsClient{client: c}
}

func (a *agreementsClient) CreateAgreement(ctx context.Context, seed AgreementSeed) (*Agreement, error) {
	body := agreementSeedToGenerated(seed)
	resp, err := a.client.CreateAgreementWithResponse(ctx, body)
	if err != nil {
		return nil, fmt.Errorf("create agreement: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON201 == nil {
		return nil, fmt.Errorf("create agreement: empty response body")
	}
	return agreementFromGenerated(resp.JSON201), nil
}

func (a *agreementsClient) GetAgreement(ctx context.Context, id uuid.UUID) (*Agreement, error) {
	resp, err := a.client.GetAgreementWithResponse(ctx, openapi_types.UUID(id))
	if err != nil {
		return nil, fmt.Errorf("get agreement: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("get agreement: empty response body")
	}
	return agreementFromGenerated(resp.JSON200), nil
}

func (a *agreementsClient) ListAgreements(ctx context.Context, params ListAgreementsParams) (*AgreementsPage, error) {
	genParams := &generated.GetAgreementsParams{
		Offset: params.Offset,
		Limit:  params.Limit,
	}

	if len(params.States) > 0 {
		states := make([]generated.AgreementState, len(params.States))
		for i, s := range params.States {
			states[i] = generated.AgreementState(s)
		}
		genParams.States = &states
	}
	if len(params.ProducerIDs) > 0 {
		ids := uuidsToOpenAPI(params.ProducerIDs)
		genParams.ProducerIds = &ids
	}
	if len(params.ConsumerIDs) > 0 {
		ids := uuidsToOpenAPI(params.ConsumerIDs)
		genParams.ConsumerIds = &ids
	}
	if len(params.DescriptorIDs) > 0 {
		ids := uuidsToOpenAPI(params.DescriptorIDs)
		genParams.DescriptorIds = &ids
	}
	if len(params.EServiceIDs) > 0 {
		ids := uuidsToOpenAPI(params.EServiceIDs)
		genParams.EserviceIds = &ids
	}

	resp, err := a.client.GetAgreementsWithResponse(ctx, genParams)
	if err != nil {
		return nil, fmt.Errorf("list agreements: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("list agreements: empty response body")
	}

	page := &AgreementsPage{
		Pagination: paginationFromGenerated(resp.JSON200.Pagination),
		Results:    make([]Agreement, len(resp.JSON200.Results)),
	}
	for i := range resp.JSON200.Results {
		page.Results[i] = *agreementFromGenerated(&resp.JSON200.Results[i])
	}
	return page, nil
}

func (a *agreementsClient) DeleteAgreement(ctx context.Context, id uuid.UUID) error {
	resp, err := a.client.DeleteAgreementWithResponse(ctx, openapi_types.UUID(id))
	if err != nil {
		return fmt.Errorf("delete agreement: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return err
	}
	return nil
}

func (a *agreementsClient) SubmitAgreement(ctx context.Context, id uuid.UUID, payload AgreementSubmission) (*Agreement, error) {
	body := agreementSubmissionToGenerated(payload)
	resp, err := a.client.SubmitAgreementWithResponse(ctx, openapi_types.UUID(id), body)
	if err != nil {
		return nil, fmt.Errorf("submit agreement: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("submit agreement: empty response body")
	}
	return agreementFromGenerated(resp.JSON200), nil
}

func (a *agreementsClient) ApproveAgreement(ctx context.Context, id uuid.UUID, payload *DelegationRef) (*Agreement, error) {
	body := delegationRefToGenerated(payload)
	resp, err := a.client.ApproveAgreementWithResponse(ctx, openapi_types.UUID(id), body)
	if err != nil {
		return nil, fmt.Errorf("approve agreement: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("approve agreement: empty response body")
	}
	return agreementFromGenerated(resp.JSON200), nil
}

func (a *agreementsClient) RejectAgreement(ctx context.Context, id uuid.UUID, payload AgreementRejection) (*Agreement, error) {
	body := agreementRejectionToGenerated(payload)
	resp, err := a.client.RejectAgreementWithResponse(ctx, openapi_types.UUID(id), body)
	if err != nil {
		return nil, fmt.Errorf("reject agreement: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("reject agreement: empty response body")
	}
	return agreementFromGenerated(resp.JSON200), nil
}

func (a *agreementsClient) SuspendAgreement(ctx context.Context, id uuid.UUID, payload *DelegationRef) (*Agreement, error) {
	body := delegationRefToGenerated(payload)
	resp, err := a.client.SuspendAgreementWithResponse(ctx, openapi_types.UUID(id), body)
	if err != nil {
		return nil, fmt.Errorf("suspend agreement: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("suspend agreement: empty response body")
	}
	return agreementFromGenerated(resp.JSON200), nil
}

func (a *agreementsClient) UnsuspendAgreement(ctx context.Context, id uuid.UUID, payload *DelegationRef) (*Agreement, error) {
	body := delegationRefToGenerated(payload)
	resp, err := a.client.UnsuspendAgreementWithResponse(ctx, openapi_types.UUID(id), body)
	if err != nil {
		return nil, fmt.Errorf("unsuspend agreement: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unsuspend agreement: empty response body")
	}
	return agreementFromGenerated(resp.JSON200), nil
}

func (a *agreementsClient) UpgradeAgreement(ctx context.Context, id uuid.UUID) (*Agreement, error) {
	resp, err := a.client.UpgradeAgreementWithResponse(ctx, openapi_types.UUID(id))
	if err != nil {
		return nil, fmt.Errorf("upgrade agreement: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("upgrade agreement: empty response body")
	}
	return agreementFromGenerated(resp.JSON200), nil
}

func (a *agreementsClient) CloneAgreement(ctx context.Context, id uuid.UUID) (*Agreement, error) {
	resp, err := a.client.CloneAgreementWithResponse(ctx, openapi_types.UUID(id))
	if err != nil {
		return nil, fmt.Errorf("clone agreement: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("clone agreement: empty response body")
	}
	return agreementFromGenerated(resp.JSON200), nil
}

func (a *agreementsClient) ListAgreementPurposes(ctx context.Context, id uuid.UUID, params PaginationParams) (*PurposesPage, error) {
	genParams := &generated.GetAgreementPurposesParams{
		Offset: params.Offset,
		Limit:  params.Limit,
	}
	resp, err := a.client.GetAgreementPurposesWithResponse(ctx, openapi_types.UUID(id), genParams)
	if err != nil {
		return nil, fmt.Errorf("list agreement purposes: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("list agreement purposes: empty response body")
	}

	page := &PurposesPage{
		Pagination: paginationFromGenerated(resp.JSON200.Pagination),
		Results:    make([]Purpose, len(resp.JSON200.Results)),
	}
	for i := range resp.JSON200.Results {
		page.Results[i] = purposeFromGenerated(&resp.JSON200.Results[i])
	}
	return page, nil
}
