package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	client "github.com/lmammino/terraform-provider-pdnd/internal/client"
	generated "github.com/lmammino/terraform-provider-pdnd/internal/client/generated"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// PurposesAPI defines operations on PDND purposes.
type PurposesAPI interface {
	CreatePurpose(ctx context.Context, seed PurposeSeed) (*Purpose, error)
	GetPurpose(ctx context.Context, id uuid.UUID) (*Purpose, error)
	DeletePurpose(ctx context.Context, id uuid.UUID) error
	UpdateDraftPurpose(ctx context.Context, id uuid.UUID, update PurposeDraftUpdate) (*Purpose, error)
	ActivatePurpose(ctx context.Context, id uuid.UUID, delegationRef *DelegationRef) (*Purpose, error)
	ApprovePurpose(ctx context.Context, id uuid.UUID, delegationRef *DelegationRef) (*Purpose, error)
	SuspendPurpose(ctx context.Context, id uuid.UUID, delegationRef *DelegationRef) (*Purpose, error)
	UnsuspendPurpose(ctx context.Context, id uuid.UUID, delegationRef *DelegationRef) (*Purpose, error)
	ArchivePurpose(ctx context.Context, id uuid.UUID) (*Purpose, error)
	CreatePurposeVersion(ctx context.Context, id uuid.UUID, seed PurposeVersionSeed) (*PurposeVersion, error)
}

type purposesClient struct {
	client *generated.ClientWithResponses
}

// NewPurposesClient creates a new PurposesAPI backed by the generated client.
func NewPurposesClient(c *generated.ClientWithResponses) PurposesAPI {
	return &purposesClient{client: c}
}

func (p *purposesClient) CreatePurpose(ctx context.Context, seed PurposeSeed) (*Purpose, error) {
	body := purposeSeedToGenerated(seed)
	resp, err := p.client.CreatePurposeWithResponse(ctx, body)
	if err != nil {
		return nil, fmt.Errorf("create purpose: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON201 == nil {
		return nil, fmt.Errorf("create purpose: empty response body")
	}
	result := purposeFromGenerated(resp.JSON201)
	return &result, nil
}

func (p *purposesClient) GetPurpose(ctx context.Context, id uuid.UUID) (*Purpose, error) {
	resp, err := p.client.GetPurposeWithResponse(ctx, openapi_types.UUID(id))
	if err != nil {
		return nil, fmt.Errorf("get purpose: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("get purpose: empty response body")
	}
	result := purposeFromGenerated(resp.JSON200)
	return &result, nil
}

func (p *purposesClient) DeletePurpose(ctx context.Context, id uuid.UUID) error {
	resp, err := p.client.DeletePurposeWithResponse(ctx, openapi_types.UUID(id))
	if err != nil {
		return fmt.Errorf("delete purpose: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return err
	}
	return nil
}

func (p *purposesClient) UpdateDraftPurpose(ctx context.Context, id uuid.UUID, update PurposeDraftUpdate) (*Purpose, error) {
	// Use WithBody variant to bypass the union type issue.
	// Marshal PurposeDraftUpdateSeed directly.
	bodyMap := generated.PurposeDraftUpdateSeed{
		Title:              update.Title,
		Description:        update.Description,
		DailyCalls:         update.DailyCalls,
		IsFreeOfCharge:     update.IsFreeOfCharge,
		FreeOfChargeReason: update.FreeOfChargeReason,
	}
	jsonBytes, err := json.Marshal(bodyMap)
	if err != nil {
		return nil, fmt.Errorf("update draft purpose: marshal body: %w", err)
	}

	resp, err := p.client.UpdateDraftPurposeWithBodyWithResponse(ctx, openapi_types.UUID(id), "application/merge-patch+json", bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("update draft purpose: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("update draft purpose: empty response body")
	}
	result := purposeFromGenerated(resp.JSON200)
	return &result, nil
}

func (p *purposesClient) ActivatePurpose(ctx context.Context, id uuid.UUID, delegationRef *DelegationRef) (*Purpose, error) {
	body := delegationRefToGenerated(delegationRef)
	resp, err := p.client.ActivateDraftPurposeWithResponse(ctx, openapi_types.UUID(id), body)
	if err != nil {
		return nil, fmt.Errorf("activate purpose: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("activate purpose: empty response body")
	}
	result := purposeFromGenerated(resp.JSON200)
	return &result, nil
}

func (p *purposesClient) ApprovePurpose(ctx context.Context, id uuid.UUID, delegationRef *DelegationRef) (*Purpose, error) {
	body := delegationRefToGenerated(delegationRef)
	resp, err := p.client.ApprovePurposeWithResponse(ctx, openapi_types.UUID(id), body)
	if err != nil {
		return nil, fmt.Errorf("approve purpose: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("approve purpose: empty response body")
	}
	result := purposeFromGenerated(resp.JSON200)
	return &result, nil
}

func (p *purposesClient) SuspendPurpose(ctx context.Context, id uuid.UUID, delegationRef *DelegationRef) (*Purpose, error) {
	body := delegationRefToGenerated(delegationRef)
	resp, err := p.client.SuspendPurposeWithResponse(ctx, openapi_types.UUID(id), body)
	if err != nil {
		return nil, fmt.Errorf("suspend purpose: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("suspend purpose: empty response body")
	}
	result := purposeFromGenerated(resp.JSON200)
	return &result, nil
}

func (p *purposesClient) UnsuspendPurpose(ctx context.Context, id uuid.UUID, delegationRef *DelegationRef) (*Purpose, error) {
	body := delegationRefToGenerated(delegationRef)
	resp, err := p.client.UnsuspendPurposeWithResponse(ctx, openapi_types.UUID(id), body)
	if err != nil {
		return nil, fmt.Errorf("unsuspend purpose: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unsuspend purpose: empty response body")
	}
	result := purposeFromGenerated(resp.JSON200)
	return &result, nil
}

func (p *purposesClient) ArchivePurpose(ctx context.Context, id uuid.UUID) (*Purpose, error) {
	resp, err := p.client.ArchivePurposeWithResponse(ctx, openapi_types.UUID(id))
	if err != nil {
		return nil, fmt.Errorf("archive purpose: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("archive purpose: empty response body")
	}
	result := purposeFromGenerated(resp.JSON200)
	return &result, nil
}

func (p *purposesClient) CreatePurposeVersion(ctx context.Context, id uuid.UUID, seed PurposeVersionSeed) (*PurposeVersion, error) {
	body := purposeVersionSeedToGenerated(seed)
	resp, err := p.client.CreatePurposeVersionWithResponse(ctx, openapi_types.UUID(id), body)
	if err != nil {
		return nil, fmt.Errorf("create purpose version: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON201 == nil {
		return nil, fmt.Errorf("create purpose version: empty response body")
	}
	return purposeVersionFromGenerated(resp.JSON201), nil
}
