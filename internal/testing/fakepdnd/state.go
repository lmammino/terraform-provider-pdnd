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

// StoredEService represents an e-service in the fake server's state.
type StoredEService struct {
	ID                      uuid.UUID
	ProducerID              uuid.UUID
	Name                    string
	Description             string
	Technology              string // REST, SOAP
	Mode                    string // RECEIVE, DELIVER
	IsSignalHubEnabled      bool
	IsConsumerDelegable     bool
	IsClientAccessDelegable bool
	PersonalData            bool
}

// StoredDescriptor represents a descriptor in the fake server's state.
type StoredDescriptor struct {
	ID                      uuid.UUID
	EServiceID              uuid.UUID
	Version                 string
	State                   string // DRAFT, PUBLISHED, DEPRECATED, SUSPENDED, ARCHIVED, WAITING_FOR_APPROVAL
	AgreementApprovalPolicy string
	Audience                []string
	DailyCallsPerConsumer   int32
	DailyCallsTotal         int32
	VoucherLifespan         int32
	ServerUrls              []string
	Description             string
	CreatedAt               time.Time
	PublishedAt             *time.Time
	SuspendedAt             *time.Time
	DeprecatedAt            *time.Time
	ArchivedAt              *time.Time
}

// StoredCertifiedAttribute represents a certified attribute in the fake server's state.
type StoredCertifiedAttribute struct {
	ID          uuid.UUID
	Name        string
	Description string
	Code        string
	Origin      string
	CreatedAt   time.Time
}

// StoredDeclaredAttribute represents a declared attribute in the fake server's state.
type StoredDeclaredAttribute struct {
	ID          uuid.UUID
	Name        string
	Description string
	CreatedAt   time.Time
}

// StoredVerifiedAttribute represents a verified attribute in the fake server's state.
type StoredVerifiedAttribute struct {
	ID          uuid.UUID
	Name        string
	Description string
	CreatedAt   time.Time
}

// StoredDescriptorAttributeGroup represents an attribute group on a descriptor.
type StoredDescriptorAttributeGroup struct {
	Attributes []uuid.UUID
}

// StoredDelegation represents a consumer or producer delegation in the fake server's state.
type StoredDelegation struct {
	ID              uuid.UUID
	DelegatorID     uuid.UUID
	DelegateID      uuid.UUID
	EServiceID      uuid.UUID
	State           string // WAITING_FOR_APPROVAL, ACTIVE, REJECTED, REVOKED
	CreatedAt       time.Time
	SubmittedAt     time.Time
	UpdatedAt       *time.Time
	ActivatedAt     *time.Time
	RejectedAt      *time.Time
	RevokedAt       *time.Time
	RejectionReason *string
}

// StoredDocument represents a document or interface file in the fake server's state.
type StoredDocument struct {
	ID          uuid.UUID
	Name        string // filename
	PrettyName  string
	ContentType string
	Content     []byte // actual file bytes
	CreatedAt   time.Time
}

// StoredPurposeVersion represents a version of a purpose.
type StoredPurposeVersion struct {
	ID                uuid.UUID
	State             string
	DailyCalls        int32
	CreatedAt         time.Time
	UpdatedAt         *time.Time
	FirstActivationAt *time.Time
	SuspendedAt       *time.Time
	RejectionReason   *string
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
	Versions            []StoredPurposeVersion
}
