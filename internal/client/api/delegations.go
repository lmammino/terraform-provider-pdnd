package api

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	client "github.com/lmammino/terraform-provider-pdnd/internal/client"
	generated "github.com/lmammino/terraform-provider-pdnd/internal/client/generated"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// DelegationsAPI defines operations on PDND delegations.
// The delegationType parameter must be "consumer" or "producer".
type DelegationsAPI interface {
	CreateDelegation(ctx context.Context, delegationType string, seed DelegationSeed) (*Delegation, error)
	GetDelegation(ctx context.Context, delegationType string, id uuid.UUID) (*Delegation, error)
	ListDelegations(ctx context.Context, delegationType string, params ListDelegationsParams) (*DelegationsPage, error)
	AcceptDelegation(ctx context.Context, delegationType string, id uuid.UUID) (*Delegation, error)
	RejectDelegation(ctx context.Context, delegationType string, id uuid.UUID, rejection DelegationRejection) (*Delegation, error)
}

type delegationsClient struct {
	client *generated.ClientWithResponses
}

// NewDelegationsClient creates a new DelegationsAPI backed by the generated client.
func NewDelegationsClient(c *generated.ClientWithResponses) DelegationsAPI {
	return &delegationsClient{client: c}
}

func (d *delegationsClient) CreateDelegation(ctx context.Context, delegationType string, seed DelegationSeed) (*Delegation, error) {
	body := delegationSeedToGenerated(seed)

	switch delegationType {
	case "consumer":
		resp, err := d.client.CreateConsumerDelegationWithResponse(ctx, body)
		if err != nil {
			return nil, fmt.Errorf("create consumer delegation: %w", err)
		}
		if err := client.CheckResponse(resp.HTTPResponse); err != nil {
			return nil, err
		}
		if resp.JSON201 == nil {
			return nil, fmt.Errorf("create consumer delegation: empty response body")
		}
		return consumerDelegationFromGenerated(resp.JSON201), nil

	case "producer":
		resp, err := d.client.CreateProducerDelegationWithResponse(ctx, body)
		if err != nil {
			return nil, fmt.Errorf("create producer delegation: %w", err)
		}
		if err := client.CheckResponse(resp.HTTPResponse); err != nil {
			return nil, err
		}
		if resp.JSON201 == nil {
			return nil, fmt.Errorf("create producer delegation: empty response body")
		}
		return producerDelegationFromGenerated(resp.JSON201), nil

	default:
		return nil, fmt.Errorf("unknown delegation type: %s", delegationType)
	}
}

func (d *delegationsClient) GetDelegation(ctx context.Context, delegationType string, id uuid.UUID) (*Delegation, error) {
	delegationID := openapi_types.UUID(id)

	switch delegationType {
	case "consumer":
		resp, err := d.client.GetConsumerDelegationWithResponse(ctx, delegationID)
		if err != nil {
			return nil, fmt.Errorf("get consumer delegation: %w", err)
		}
		if err := client.CheckResponse(resp.HTTPResponse); err != nil {
			return nil, err
		}
		if resp.JSON200 == nil {
			return nil, fmt.Errorf("get consumer delegation: empty response body")
		}
		return consumerDelegationFromGenerated(resp.JSON200), nil

	case "producer":
		resp, err := d.client.GetProducerDelegationWithResponse(ctx, delegationID)
		if err != nil {
			return nil, fmt.Errorf("get producer delegation: %w", err)
		}
		if err := client.CheckResponse(resp.HTTPResponse); err != nil {
			return nil, err
		}
		if resp.JSON200 == nil {
			return nil, fmt.Errorf("get producer delegation: empty response body")
		}
		return producerDelegationFromGenerated(resp.JSON200), nil

	default:
		return nil, fmt.Errorf("unknown delegation type: %s", delegationType)
	}
}

func (d *delegationsClient) ListDelegations(ctx context.Context, delegationType string, params ListDelegationsParams) (*DelegationsPage, error) {
	switch delegationType {
	case "consumer":
		genParams := &generated.GetConsumerDelegationsParams{
			Offset: params.Offset,
			Limit:  params.Limit,
		}
		if len(params.States) > 0 {
			states := make([]generated.DelegationState, len(params.States))
			for i, s := range params.States {
				states[i] = generated.DelegationState(s)
			}
			genParams.States = &states
		}
		if len(params.DelegatorIDs) > 0 {
			ids := uuidsToOpenAPI(params.DelegatorIDs)
			genParams.DelegatorIds = &ids
		}
		if len(params.DelegateIDs) > 0 {
			ids := uuidsToOpenAPI(params.DelegateIDs)
			genParams.DelegateIds = &ids
		}
		if len(params.EServiceIDs) > 0 {
			ids := uuidsToOpenAPI(params.EServiceIDs)
			genParams.EserviceIds = &ids
		}

		resp, err := d.client.GetConsumerDelegationsWithResponse(ctx, genParams)
		if err != nil {
			return nil, fmt.Errorf("list consumer delegations: %w", err)
		}
		if err := client.CheckResponse(resp.HTTPResponse); err != nil {
			return nil, err
		}
		if resp.JSON200 == nil {
			return nil, fmt.Errorf("list consumer delegations: empty response body")
		}

		page := &DelegationsPage{
			Pagination: paginationFromGenerated(resp.JSON200.Pagination),
			Results:    make([]Delegation, len(resp.JSON200.Results)),
		}
		for i := range resp.JSON200.Results {
			page.Results[i] = *consumerDelegationFromGenerated(&resp.JSON200.Results[i])
		}
		return page, nil

	case "producer":
		genParams := &generated.GetProducerDelegationsParams{
			Offset: params.Offset,
			Limit:  params.Limit,
		}
		if len(params.States) > 0 {
			states := make([]generated.DelegationState, len(params.States))
			for i, s := range params.States {
				states[i] = generated.DelegationState(s)
			}
			genParams.States = &states
		}
		if len(params.DelegatorIDs) > 0 {
			ids := uuidsToOpenAPI(params.DelegatorIDs)
			genParams.DelegatorIds = &ids
		}
		if len(params.DelegateIDs) > 0 {
			ids := uuidsToOpenAPI(params.DelegateIDs)
			genParams.DelegateIds = &ids
		}
		if len(params.EServiceIDs) > 0 {
			ids := uuidsToOpenAPI(params.EServiceIDs)
			genParams.EserviceIds = &ids
		}

		resp, err := d.client.GetProducerDelegationsWithResponse(ctx, genParams)
		if err != nil {
			return nil, fmt.Errorf("list producer delegations: %w", err)
		}
		if err := client.CheckResponse(resp.HTTPResponse); err != nil {
			return nil, err
		}
		if resp.JSON200 == nil {
			return nil, fmt.Errorf("list producer delegations: empty response body")
		}

		page := &DelegationsPage{
			Pagination: paginationFromGenerated(resp.JSON200.Pagination),
			Results:    make([]Delegation, len(resp.JSON200.Results)),
		}
		for i := range resp.JSON200.Results {
			page.Results[i] = *producerDelegationFromGenerated(&resp.JSON200.Results[i])
		}
		return page, nil

	default:
		return nil, fmt.Errorf("unknown delegation type: %s", delegationType)
	}
}

func (d *delegationsClient) AcceptDelegation(ctx context.Context, delegationType string, id uuid.UUID) (*Delegation, error) {
	delegationID := openapi_types.UUID(id)

	switch delegationType {
	case "consumer":
		resp, err := d.client.AcceptConsumerDelegationWithResponse(ctx, delegationID)
		if err != nil {
			return nil, fmt.Errorf("accept consumer delegation: %w", err)
		}
		if err := client.CheckResponse(resp.HTTPResponse); err != nil {
			return nil, err
		}
		if resp.JSON200 == nil {
			return nil, fmt.Errorf("accept consumer delegation: empty response body")
		}
		return consumerDelegationFromGenerated(resp.JSON200), nil

	case "producer":
		resp, err := d.client.AcceptProducerDelegationWithResponse(ctx, delegationID)
		if err != nil {
			return nil, fmt.Errorf("accept producer delegation: %w", err)
		}
		if err := client.CheckResponse(resp.HTTPResponse); err != nil {
			return nil, err
		}
		if resp.JSON200 == nil {
			return nil, fmt.Errorf("accept producer delegation: empty response body")
		}
		return producerDelegationFromGenerated(resp.JSON200), nil

	default:
		return nil, fmt.Errorf("unknown delegation type: %s", delegationType)
	}
}

func (d *delegationsClient) RejectDelegation(ctx context.Context, delegationType string, id uuid.UUID, rejection DelegationRejection) (*Delegation, error) {
	delegationID := openapi_types.UUID(id)
	body := delegationRejectionToGenerated(rejection)

	switch delegationType {
	case "consumer":
		resp, err := d.client.RejectConsumerDelegationWithResponse(ctx, delegationID, body)
		if err != nil {
			return nil, fmt.Errorf("reject consumer delegation: %w", err)
		}
		if err := client.CheckResponse(resp.HTTPResponse); err != nil {
			return nil, err
		}
		if resp.JSON200 == nil {
			return nil, fmt.Errorf("reject consumer delegation: empty response body")
		}
		return consumerDelegationFromGenerated(resp.JSON200), nil

	case "producer":
		resp, err := d.client.RejectProducerDelegationWithResponse(ctx, delegationID, body)
		if err != nil {
			return nil, fmt.Errorf("reject producer delegation: %w", err)
		}
		if err := client.CheckResponse(resp.HTTPResponse); err != nil {
			return nil, err
		}
		if resp.JSON200 == nil {
			return nil, fmt.Errorf("reject producer delegation: empty response body")
		}
		return producerDelegationFromGenerated(resp.JSON200), nil

	default:
		return nil, fmt.Errorf("unknown delegation type: %s", delegationType)
	}
}
