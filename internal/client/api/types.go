package api

import (
	"time"

	"github.com/google/uuid"
)

// EService represents a PDND e-service.
type EService struct {
	ID                      uuid.UUID
	ProducerID              uuid.UUID
	Name                    string
	Description             string
	Technology              string // "REST" or "SOAP"
	Mode                    string // "RECEIVE" or "DELIVER"
	IsSignalHubEnabled      *bool
	IsConsumerDelegable     *bool
	IsClientAccessDelegable *bool
	PersonalData            *bool
	TemplateID              *uuid.UUID
}

// EServiceSeed contains fields for creating a new e-service.
type EServiceSeed struct {
	Name                    string
	Description             string
	Technology              string
	Mode                    string
	IsSignalHubEnabled      *bool
	IsConsumerDelegable     *bool
	IsClientAccessDelegable *bool
	PersonalData            *bool
	Descriptor              DescriptorSeedForCreation
}

// DescriptorSeedForCreation contains the descriptor fields required when creating an e-service.
type DescriptorSeedForCreation struct {
	AgreementApprovalPolicy string
	Audience                []string
	DailyCallsPerConsumer   int32
	DailyCallsTotal         int32
	VoucherLifespan         int32
	Description             *string
}

// EServiceDraftUpdate contains fields for updating a draft e-service.
type EServiceDraftUpdate struct {
	Name                    *string
	Description             *string
	Technology              *string
	Mode                    *string
	IsSignalHubEnabled      *bool
	IsConsumerDelegable     *bool
	IsClientAccessDelegable *bool
	PersonalData            *bool
}

// EServiceDelegationUpdate contains fields for updating delegation settings on a published e-service.
type EServiceDelegationUpdate struct {
	IsConsumerDelegable     *bool
	IsClientAccessDelegable *bool
}

// Descriptor represents a PDND e-service descriptor.
type Descriptor struct {
	ID                      uuid.UUID
	Version                 string
	State                   string // DRAFT, PUBLISHED, DEPRECATED, SUSPENDED, ARCHIVED, WAITING_FOR_APPROVAL
	AgreementApprovalPolicy string
	Audience                []string
	DailyCallsPerConsumer   int32
	DailyCallsTotal         int32
	VoucherLifespan         int32
	ServerUrls              []string
	Description             *string
	PublishedAt             *time.Time
	SuspendedAt             *time.Time
	DeprecatedAt            *time.Time
	ArchivedAt              *time.Time
}

// DescriptorSeed contains fields for creating a new descriptor.
type DescriptorSeed struct {
	AgreementApprovalPolicy string
	Audience                []string
	DailyCallsPerConsumer   int32
	DailyCallsTotal         int32
	VoucherLifespan         int32
	Description             *string
}

// DescriptorDraftUpdate contains fields for updating a draft descriptor.
type DescriptorDraftUpdate struct {
	AgreementApprovalPolicy *string
	Audience                []string // nil means no change
	DailyCallsPerConsumer   *int32
	DailyCallsTotal         *int32
	VoucherLifespan         *int32
	Description             *string
}

// DescriptorQuotasUpdate contains fields for updating quotas on a published descriptor.
type DescriptorQuotasUpdate struct {
	DailyCallsPerConsumer *int32
	DailyCallsTotal       *int32
	VoucherLifespan       *int32
}

// ListEServicesParams contains filter parameters for listing e-services.
type ListEServicesParams struct {
	ProducerIDs []uuid.UUID
	Name        *string
	Technology  *string
	Mode        *string
	Offset      int32
	Limit       int32
}

// EServicesPage is a paginated list of e-services.
type EServicesPage struct {
	Results    []EService
	Pagination Pagination
}

// ListDescriptorsParams contains filter parameters for listing descriptors.
type ListDescriptorsParams struct {
	State  *string
	Offset int32
	Limit  int32
}

// DescriptorsPage is a paginated list of descriptors.
type DescriptorsPage struct {
	Results    []Descriptor
	Pagination Pagination
}

// Agreement represents a PDND agreement with all fields from the API.
type Agreement struct {
	ID                  uuid.UUID
	EServiceID          uuid.UUID
	DescriptorID        uuid.UUID
	ProducerID          uuid.UUID
	ConsumerID          uuid.UUID
	DelegationID        *uuid.UUID
	State               string // One of: DRAFT, ACTIVE, ARCHIVED, PENDING, SUSPENDED, MISSING_CERTIFIED_ATTRIBUTES, REJECTED
	SuspendedByConsumer *bool
	SuspendedByProducer *bool
	SuspendedByPlatform *bool
	ConsumerNotes       *string
	RejectionReason     *string
	CreatedAt           time.Time
	UpdatedAt           *time.Time
	SuspendedAt         *time.Time
}

// AgreementSeed contains fields for creating a new agreement.
// Maps to POST /agreements request body.
type AgreementSeed struct {
	EServiceID   uuid.UUID  // required
	DescriptorID uuid.UUID  // required
	DelegationID *uuid.UUID // optional
}

// AgreementSubmission contains fields for submitting a draft agreement.
// Maps to POST /agreements/{id}/submit request body.
type AgreementSubmission struct {
	ConsumerNotes *string // optional, max 1000 chars
}

// AgreementRejection contains fields for rejecting a pending agreement.
// Maps to POST /agreements/{id}/reject request body.
type AgreementRejection struct {
	Reason string // required, 20-1000 chars
}

// DelegationRef identifies a delegation for delegate operations.
// Used by approve, suspend, unsuspend endpoints.
type DelegationRef struct {
	DelegationID uuid.UUID // required
}

// ListAgreementsParams contains filter parameters for listing agreements.
type ListAgreementsParams struct {
	States        []string    // AgreementState values to filter by
	ProducerIDs   []uuid.UUID
	ConsumerIDs   []uuid.UUID
	DescriptorIDs []uuid.UUID
	EServiceIDs   []uuid.UUID
	Offset        int32 // required, min 0
	Limit         int32 // required, 1-50
}

// PaginationParams for paginated list endpoints.
type PaginationParams struct {
	Offset int32
	Limit  int32
}

// AgreementsPage is a paginated list of agreements.
type AgreementsPage struct {
	Results    []Agreement
	Pagination Pagination
}

// PurposesPage is a paginated list of purposes.
type PurposesPage struct {
	Results    []Purpose
	Pagination Pagination
}

// Pagination contains pagination metadata from list responses.
type Pagination struct {
	Offset     int32
	Limit      int32
	TotalCount int32
}

// Purpose represents a PDND purpose associated with an agreement.
type Purpose struct {
	ID                  uuid.UUID
	EServiceID          uuid.UUID
	ConsumerID          uuid.UUID
	SuspendedByConsumer *bool
	SuspendedByProducer *bool
	Title               string
	Description         string
	CreatedAt           time.Time
	UpdatedAt           *time.Time
	IsRiskAnalysisValid bool
	IsFreeOfCharge      bool
	FreeOfChargeReason  *string
	DelegationID        *uuid.UUID
}
