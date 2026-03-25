package api

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	client "github.com/lmammino/terraform-provider-pdnd/internal/client"
	generated "github.com/lmammino/terraform-provider-pdnd/internal/client/generated"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// EServicesAPI defines operations on PDND e-services and their descriptors.
// This interface is the boundary between Terraform resource code and the HTTP client.
type EServicesAPI interface {
	// E-Service CRUD
	CreateEService(ctx context.Context, seed EServiceSeed) (*EService, error)
	GetEService(ctx context.Context, id uuid.UUID) (*EService, error)
	ListEServices(ctx context.Context, params ListEServicesParams) (*EServicesPage, error)
	DeleteEService(ctx context.Context, id uuid.UUID) error
	UpdateDraftEService(ctx context.Context, id uuid.UUID, seed EServiceDraftUpdate) (*EService, error)

	// Published e-service updates (per-field endpoints)
	UpdatePublishedEServiceName(ctx context.Context, id uuid.UUID, name string) (*EService, error)
	UpdatePublishedEServiceDescription(ctx context.Context, id uuid.UUID, description string) (*EService, error)
	UpdatePublishedEServiceDelegation(ctx context.Context, id uuid.UUID, seed EServiceDelegationUpdate) (*EService, error)
	UpdatePublishedEServiceSignalHub(ctx context.Context, id uuid.UUID, enabled bool) (*EService, error)

	// Descriptor CRUD
	CreateDescriptor(ctx context.Context, eserviceID uuid.UUID, seed DescriptorSeed) (*Descriptor, error)
	GetDescriptor(ctx context.Context, eserviceID, descriptorID uuid.UUID) (*Descriptor, error)
	ListDescriptors(ctx context.Context, eserviceID uuid.UUID, params ListDescriptorsParams) (*DescriptorsPage, error)
	DeleteDraftDescriptor(ctx context.Context, eserviceID, descriptorID uuid.UUID) error
	UpdateDraftDescriptor(ctx context.Context, eserviceID, descriptorID uuid.UUID, seed DescriptorDraftUpdate) (*Descriptor, error)
	UpdatePublishedDescriptorQuotas(ctx context.Context, eserviceID, descriptorID uuid.UUID, seed DescriptorQuotasUpdate) (*Descriptor, error)

	// Descriptor state transitions
	PublishDescriptor(ctx context.Context, eserviceID, descriptorID uuid.UUID) error
	SuspendDescriptor(ctx context.Context, eserviceID, descriptorID uuid.UUID) error
	UnsuspendDescriptor(ctx context.Context, eserviceID, descriptorID uuid.UUID) error
	ApproveDelegatedDescriptor(ctx context.Context, eserviceID, descriptorID uuid.UUID) error
}

// eservicesClient implements EServicesAPI using the generated client.
type eservicesClient struct {
	client *generated.ClientWithResponses
}

// NewEServicesClient creates a new EServicesAPI backed by the generated client.
func NewEServicesClient(c *generated.ClientWithResponses) EServicesAPI {
	return &eservicesClient{client: c}
}

func (e *eservicesClient) CreateEService(ctx context.Context, seed EServiceSeed) (*EService, error) {
	body := eserviceSeedToGenerated(seed)
	resp, err := e.client.CreateEServiceWithResponse(ctx, body)
	if err != nil {
		return nil, fmt.Errorf("create eservice: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON201 == nil {
		return nil, fmt.Errorf("create eservice: empty response body")
	}
	return eserviceFromGenerated(resp.JSON201), nil
}

func (e *eservicesClient) GetEService(ctx context.Context, id uuid.UUID) (*EService, error) {
	resp, err := e.client.GetEServiceWithResponse(ctx, openapi_types.UUID(id))
	if err != nil {
		return nil, fmt.Errorf("get eservice: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("get eservice: empty response body")
	}
	return eserviceFromGenerated(resp.JSON200), nil
}

func (e *eservicesClient) ListEServices(ctx context.Context, params ListEServicesParams) (*EServicesPage, error) {
	genParams := &generated.GetEServicesParams{
		Offset: params.Offset,
		Limit:  params.Limit,
	}

	if len(params.ProducerIDs) > 0 {
		ids := uuidsToOpenAPI(params.ProducerIDs)
		genParams.ProducerIds = &ids
	}
	if params.Name != nil {
		genParams.Name = params.Name
	}
	if params.Technology != nil {
		t := generated.EServiceTechnology(*params.Technology)
		genParams.Technology = &t
	}
	if params.Mode != nil {
		m := generated.EServiceMode(*params.Mode)
		genParams.Mode = &m
	}

	resp, err := e.client.GetEServicesWithResponse(ctx, genParams)
	if err != nil {
		return nil, fmt.Errorf("list eservices: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("list eservices: empty response body")
	}

	page := &EServicesPage{
		Pagination: paginationFromGenerated(resp.JSON200.Pagination),
		Results:    make([]EService, len(resp.JSON200.Results)),
	}
	for i := range resp.JSON200.Results {
		page.Results[i] = *eserviceFromGenerated(&resp.JSON200.Results[i])
	}
	return page, nil
}

func (e *eservicesClient) DeleteEService(ctx context.Context, id uuid.UUID) error {
	resp, err := e.client.DeleteEServiceWithResponse(ctx, openapi_types.UUID(id))
	if err != nil {
		return fmt.Errorf("delete eservice: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return err
	}
	return nil
}

func (e *eservicesClient) UpdateDraftEService(ctx context.Context, id uuid.UUID, seed EServiceDraftUpdate) (*EService, error) {
	body := eserviceDraftUpdateToGenerated(seed)
	resp, err := e.client.UpdateDraftEServiceWithApplicationMergePatchPlusJSONBodyWithResponse(ctx, openapi_types.UUID(id), body)
	if err != nil {
		return nil, fmt.Errorf("update draft eservice: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("update draft eservice: empty response body")
	}
	return eserviceFromGenerated(resp.JSON200), nil
}

func (e *eservicesClient) UpdatePublishedEServiceName(ctx context.Context, id uuid.UUID, name string) (*EService, error) {
	body := generated.EServiceNameUpdateSeed{Name: name}
	resp, err := e.client.UpdatePublishedEServiceNameWithApplicationMergePatchPlusJSONBodyWithResponse(ctx, openapi_types.UUID(id), body)
	if err != nil {
		return nil, fmt.Errorf("update published eservice name: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("update published eservice name: empty response body")
	}
	return eserviceFromGenerated(resp.JSON200), nil
}

func (e *eservicesClient) UpdatePublishedEServiceDescription(ctx context.Context, id uuid.UUID, description string) (*EService, error) {
	body := generated.EServiceDescriptionUpdateSeed{Description: description}
	resp, err := e.client.UpdatePublishedEServiceDescriptionWithApplicationMergePatchPlusJSONBodyWithResponse(ctx, openapi_types.UUID(id), body)
	if err != nil {
		return nil, fmt.Errorf("update published eservice description: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("update published eservice description: empty response body")
	}
	return eserviceFromGenerated(resp.JSON200), nil
}

func (e *eservicesClient) UpdatePublishedEServiceDelegation(ctx context.Context, id uuid.UUID, seed EServiceDelegationUpdate) (*EService, error) {
	body := generated.EServiceDelegationUpdateSeed{
		IsConsumerDelegable:     seed.IsConsumerDelegable,
		IsClientAccessDelegable: seed.IsClientAccessDelegable,
	}
	resp, err := e.client.UpdatePublishedEServiceDelegationWithApplicationMergePatchPlusJSONBodyWithResponse(ctx, openapi_types.UUID(id), body)
	if err != nil {
		return nil, fmt.Errorf("update published eservice delegation: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("update published eservice delegation: empty response body")
	}
	return eserviceFromGenerated(resp.JSON200), nil
}

func (e *eservicesClient) UpdatePublishedEServiceSignalHub(ctx context.Context, id uuid.UUID, enabled bool) (*EService, error) {
	body := generated.EServiceSignalHubUpdateSeed{IsSignalHubEnabled: enabled}
	resp, err := e.client.UpdatePublishedEServiceSignalHubWithApplicationMergePatchPlusJSONBodyWithResponse(ctx, openapi_types.UUID(id), body)
	if err != nil {
		return nil, fmt.Errorf("update published eservice signal hub: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("update published eservice signal hub: empty response body")
	}
	return eserviceFromGenerated(resp.JSON200), nil
}

func (e *eservicesClient) CreateDescriptor(ctx context.Context, eserviceID uuid.UUID, seed DescriptorSeed) (*Descriptor, error) {
	body := descriptorSeedToGenerated(seed)
	resp, err := e.client.CreateDescriptorWithResponse(ctx, openapi_types.UUID(eserviceID), body)
	if err != nil {
		return nil, fmt.Errorf("create descriptor: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON201 == nil {
		return nil, fmt.Errorf("create descriptor: empty response body")
	}
	return descriptorFromGenerated(resp.JSON201), nil
}

func (e *eservicesClient) GetDescriptor(ctx context.Context, eserviceID, descriptorID uuid.UUID) (*Descriptor, error) {
	resp, err := e.client.GetEServiceDescriptorWithResponse(ctx, openapi_types.UUID(eserviceID), openapi_types.UUID(descriptorID))
	if err != nil {
		return nil, fmt.Errorf("get descriptor: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("get descriptor: empty response body")
	}
	return descriptorFromGenerated(resp.JSON200), nil
}

func (e *eservicesClient) ListDescriptors(ctx context.Context, eserviceID uuid.UUID, params ListDescriptorsParams) (*DescriptorsPage, error) {
	genParams := &generated.GetEServiceDescriptorsParams{
		Offset: params.Offset,
		Limit:  params.Limit,
	}
	if params.State != nil {
		s := generated.EServiceDescriptorState(*params.State)
		genParams.State = &s
	}

	resp, err := e.client.GetEServiceDescriptorsWithResponse(ctx, openapi_types.UUID(eserviceID), genParams)
	if err != nil {
		return nil, fmt.Errorf("list descriptors: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("list descriptors: empty response body")
	}

	page := &DescriptorsPage{
		Pagination: paginationFromGenerated(resp.JSON200.Pagination),
		Results:    make([]Descriptor, len(resp.JSON200.Results)),
	}
	for i := range resp.JSON200.Results {
		page.Results[i] = *descriptorFromGenerated(&resp.JSON200.Results[i])
	}
	return page, nil
}

func (e *eservicesClient) DeleteDraftDescriptor(ctx context.Context, eserviceID, descriptorID uuid.UUID) error {
	resp, err := e.client.DeleteDraftEServiceDescriptorWithResponse(ctx, openapi_types.UUID(eserviceID), openapi_types.UUID(descriptorID))
	if err != nil {
		return fmt.Errorf("delete draft descriptor: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return err
	}
	return nil
}

func (e *eservicesClient) UpdateDraftDescriptor(ctx context.Context, eserviceID, descriptorID uuid.UUID, seed DescriptorDraftUpdate) (*Descriptor, error) {
	body := descriptorDraftUpdateToGenerated(seed)
	resp, err := e.client.UpdateDraftEServiceDescriptorWithApplicationMergePatchPlusJSONBodyWithResponse(ctx, openapi_types.UUID(eserviceID), openapi_types.UUID(descriptorID), body)
	if err != nil {
		return nil, fmt.Errorf("update draft descriptor: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("update draft descriptor: empty response body")
	}
	return descriptorFromGenerated(resp.JSON200), nil
}

func (e *eservicesClient) UpdatePublishedDescriptorQuotas(ctx context.Context, eserviceID, descriptorID uuid.UUID, seed DescriptorQuotasUpdate) (*Descriptor, error) {
	body := descriptorQuotasUpdateToGenerated(seed)
	resp, err := e.client.UpdatePublishedEServiceDescriptorQuotasWithApplicationMergePatchPlusJSONBodyWithResponse(ctx, openapi_types.UUID(eserviceID), openapi_types.UUID(descriptorID), body)
	if err != nil {
		return nil, fmt.Errorf("update published descriptor quotas: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("update published descriptor quotas: empty response body")
	}
	return descriptorFromGenerated(resp.JSON200), nil
}

func (e *eservicesClient) PublishDescriptor(ctx context.Context, eserviceID, descriptorID uuid.UUID) error {
	resp, err := e.client.PublishDescriptorWithResponse(ctx, openapi_types.UUID(eserviceID), openapi_types.UUID(descriptorID))
	if err != nil {
		return fmt.Errorf("publish descriptor: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return err
	}
	return nil
}

func (e *eservicesClient) SuspendDescriptor(ctx context.Context, eserviceID, descriptorID uuid.UUID) error {
	resp, err := e.client.SuspendDescriptorWithResponse(ctx, openapi_types.UUID(eserviceID), openapi_types.UUID(descriptorID))
	if err != nil {
		return fmt.Errorf("suspend descriptor: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return err
	}
	return nil
}

func (e *eservicesClient) UnsuspendDescriptor(ctx context.Context, eserviceID, descriptorID uuid.UUID) error {
	resp, err := e.client.UnsuspendDescriptorWithResponse(ctx, openapi_types.UUID(eserviceID), openapi_types.UUID(descriptorID))
	if err != nil {
		return fmt.Errorf("unsuspend descriptor: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return err
	}
	return nil
}

func (e *eservicesClient) ApproveDelegatedDescriptor(ctx context.Context, eserviceID, descriptorID uuid.UUID) error {
	resp, err := e.client.ApproveDelegatedEServiceDescriptorWithResponse(ctx, openapi_types.UUID(eserviceID), openapi_types.UUID(descriptorID))
	if err != nil {
		return fmt.Errorf("approve delegated descriptor: %w", err)
	}
	if err := client.CheckResponse(resp.HTTPResponse); err != nil {
		return err
	}
	return nil
}
