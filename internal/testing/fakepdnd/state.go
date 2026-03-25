package fakepdnd

import (
	"time"

	"github.com/google/uuid"
)

// StoredAgreement represents an agreement in the fake server's state.
type StoredAgreement struct {
	ID                  uuid.UUID
	EServiceID          uuid.UUID
	DescriptorID        uuid.UUID
	ProducerID          uuid.UUID
	ConsumerID          uuid.UUID
	DelegationID        *uuid.UUID
	State               string
	SuspendedByConsumer bool
	SuspendedByProducer bool
	SuspendedByPlatform bool
	ConsumerNotes       string
	RejectionReason     string
	CreatedAt           time.Time
	UpdatedAt           *time.Time
	SuspendedAt         *time.Time
}

// StoredPurpose represents a purpose in the fake server's state.
type StoredPurpose struct {
	ID                  uuid.UUID
	EServiceID          uuid.UUID
	ConsumerID          uuid.UUID
	SuspendedByConsumer bool
	SuspendedByProducer bool
	Title               string
	Description         string
	CreatedAt           time.Time
	UpdatedAt           *time.Time
	IsRiskAnalysisValid bool
	IsFreeOfCharge      bool
	FreeOfChargeReason  string
	DelegationID        *uuid.UUID
}
