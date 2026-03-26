package api

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	client "github.com/lmammino/terraform-provider-pdnd/internal/client"
	generated "github.com/lmammino/terraform-provider-pdnd/internal/client/generated"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// ClientsAPI defines operations on PDND clients.
type ClientsAPI interface {
	GetClient(ctx context.Context, id uuid.UUID) (*ClientInfo, error)
	ListClients(ctx context.Context, params ListClientsParams) (*ClientsPage, error)
	ListClientKeys(ctx context.Context, clientID uuid.UUID, offset, limit int32) (*ClientKeysPage, error)
	CreateClientKey(ctx context.Context, clientID uuid.UUID, seed ClientKeySeed) (*ClientKeyDetail, error)
	DeleteClientKey(ctx context.Context, clientID uuid.UUID, kid string) error
	AddClientPurpose(ctx context.Context, clientID, purposeID uuid.UUID) error
	RemoveClientPurpose(ctx context.Context, clientID, purposeID uuid.UUID) error
}

type clientsClient struct {
	client *generated.ClientWithResponses
}

// NewClientsClient creates a new ClientsAPI backed by the generated client.
func NewClientsClient(c *generated.ClientWithResponses) ClientsAPI {
	return &clientsClient{client: c}
}

func (cc *clientsClient) GetClient(ctx context.Context, id uuid.UUID) (*ClientInfo, error) {
	resp, err := cc.client.GetClientWithResponse(ctx, openapi_types.UUID(id))
	if err != nil {
		return nil, fmt.Errorf("get client: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("get client: empty response body")
	}
	return clientInfoFromGenerated(resp.JSON200)
}

func (cc *clientsClient) ListClients(ctx context.Context, params ListClientsParams) (*ClientsPage, error) {
	genParams := &generated.GetClientsParams{
		Offset: params.Offset,
		Limit:  params.Limit,
		Name:   params.Name,
	}
	if params.ConsumerID != nil {
		id := openapi_types.UUID(*params.ConsumerID)
		genParams.ConsumerId = &id
	}

	resp, err := cc.client.GetClientsWithResponse(ctx, genParams)
	if err != nil {
		return nil, fmt.Errorf("list clients: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("list clients: empty response body")
	}

	page := &ClientsPage{
		Pagination: paginationFromGenerated(resp.JSON200.Pagination),
		Results:    make([]ClientInfo, 0, len(resp.JSON200.Results)),
	}
	for i := range resp.JSON200.Results {
		info, err := clientInfoFromGenerated(&resp.JSON200.Results[i])
		if err != nil {
			continue // skip parse errors for partial clients
		}
		page.Results = append(page.Results, *info)
	}
	return page, nil
}

func (cc *clientsClient) ListClientKeys(ctx context.Context, clientID uuid.UUID, offset, limit int32) (*ClientKeysPage, error) {
	genParams := &generated.GetClientKeysParams{
		Offset: offset,
		Limit:  limit,
	}
	resp, err := cc.client.GetClientKeysWithResponse(ctx, openapi_types.UUID(clientID), genParams)
	if err != nil {
		return nil, fmt.Errorf("list client keys: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("list client keys: empty response body")
	}

	page := &ClientKeysPage{
		Pagination: paginationFromGenerated(resp.JSON200.Pagination),
		Results:    make([]ClientKey, len(resp.JSON200.Results)),
	}
	for i := range resp.JSON200.Results {
		page.Results[i] = clientKeyFromJWK(&resp.JSON200.Results[i])
	}
	return page, nil
}

func (cc *clientsClient) CreateClientKey(ctx context.Context, clientID uuid.UUID, seed ClientKeySeed) (*ClientKeyDetail, error) {
	body := clientKeySeedToGenerated(seed)
	resp, err := cc.client.CreateClientKeyWithResponse(ctx, openapi_types.UUID(clientID), body)
	if err != nil {
		return nil, fmt.Errorf("create client key: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("create client key: empty response body")
	}
	return clientKeyDetailFromGenerated(resp.JSON200), nil
}

func (cc *clientsClient) DeleteClientKey(ctx context.Context, clientID uuid.UUID, kid string) error {
	resp, err := cc.client.DeleteClientKeyByIdWithResponse(ctx, openapi_types.UUID(clientID), kid)
	if err != nil {
		return fmt.Errorf("delete client key: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return err
	}
	return nil
}

func (cc *clientsClient) AddClientPurpose(ctx context.Context, clientID, purposeID uuid.UUID) error {
	body := generated.ClientAddPurpose{
		PurposeId: openapi_types.UUID(purposeID),
	}
	resp, err := cc.client.AddClientPurposeWithResponse(ctx, openapi_types.UUID(clientID), body)
	if err != nil {
		return fmt.Errorf("add client purpose: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return err
	}
	return nil
}

func (cc *clientsClient) RemoveClientPurpose(ctx context.Context, clientID, purposeID uuid.UUID) error {
	resp, err := cc.client.RemoveClientPurposeWithResponse(ctx, openapi_types.UUID(clientID), openapi_types.UUID(purposeID))
	if err != nil {
		return fmt.Errorf("remove client purpose: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return err
	}
	return nil
}
