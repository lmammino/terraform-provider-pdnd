package api

import (
	"time"

	"github.com/google/uuid"
)

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
