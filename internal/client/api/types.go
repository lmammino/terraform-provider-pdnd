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

// Delegation represents a consumer or producer delegation.
type Delegation struct {
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

// DelegationSeed contains fields for creating a new delegation.
type DelegationSeed struct {
	EServiceID uuid.UUID
	DelegateID uuid.UUID
}

// DelegationRejection contains fields for rejecting a delegation.
type DelegationRejection struct {
	RejectionReason string
}

// DelegationsPage is a paginated list of delegations.
type DelegationsPage struct {
	Results    []Delegation
	Pagination Pagination
}

// ListDelegationsParams contains filter parameters for listing delegations.
type ListDelegationsParams struct {
	States       []string
	DelegatorIDs []uuid.UUID
	DelegateIDs  []uuid.UUID
	EServiceIDs  []uuid.UUID
	Offset       int32
	Limit        int32
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

// CertifiedAttribute represents a PDND certified attribute.
type CertifiedAttribute struct {
	ID          uuid.UUID
	Name        string
	Description string
	Code        string
	Origin      string
	CreatedAt   time.Time
}

// DeclaredAttribute represents a PDND declared attribute.
type DeclaredAttribute struct {
	ID          uuid.UUID
	Name        string
	Description string
	CreatedAt   time.Time
}

// VerifiedAttribute represents a PDND verified attribute.
type VerifiedAttribute struct {
	ID          uuid.UUID
	Name        string
	Description string
	CreatedAt   time.Time
}

// CertifiedAttributesPage is a paginated list of certified attributes.
type CertifiedAttributesPage struct {
	Results    []CertifiedAttribute
	Pagination Pagination
}

// DeclaredAttributesPage is a paginated list of declared attributes.
type DeclaredAttributesPage struct {
	Results    []DeclaredAttribute
	Pagination Pagination
}

// VerifiedAttributesPage is a paginated list of verified attributes.
type VerifiedAttributesPage struct {
	Results    []VerifiedAttribute
	Pagination Pagination
}

// DescriptorAttributeEntry represents an attribute assigned to a descriptor group.
type DescriptorAttributeEntry struct {
	AttributeID uuid.UUID
	GroupIndex  int32
}

// DescriptorDocument represents a file (document or interface) on a descriptor.
type DescriptorDocument struct {
	ID          uuid.UUID
	Name        string
	PrettyName  string
	ContentType string
	CreatedAt   time.Time
}

// PurposeVersion represents a version of a PDND purpose.
type PurposeVersion struct {
	ID                uuid.UUID
	State             string // DRAFT, ACTIVE, SUSPENDED, ARCHIVED, WAITING_FOR_APPROVAL, REJECTED
	DailyCalls        int32
	CreatedAt         time.Time
	UpdatedAt         *time.Time
	FirstActivationAt *time.Time
	SuspendedAt       *time.Time
	RejectionReason   *string
}

// Purpose represents a PDND purpose.
type Purpose struct {
	ID                        uuid.UUID
	EServiceID                uuid.UUID
	ConsumerID                uuid.UUID
	SuspendedByConsumer       *bool
	SuspendedByProducer       *bool
	Title                     string
	Description               string
	CreatedAt                 time.Time
	UpdatedAt                 *time.Time
	IsRiskAnalysisValid       bool
	IsFreeOfCharge            bool
	FreeOfChargeReason        *string
	DelegationID              *uuid.UUID
	CurrentVersion            *PurposeVersion
	WaitingForApprovalVersion *PurposeVersion
	RejectedVersion           *PurposeVersion
	PurposeTemplateID         *uuid.UUID
}

// PurposeSeed contains fields for creating a new purpose.
type PurposeSeed struct {
	EServiceID         uuid.UUID
	Title              string
	Description        string
	DailyCalls         int32
	IsFreeOfCharge     bool
	FreeOfChargeReason *string
	DelegationID       *uuid.UUID
}

// PurposeDraftUpdate contains fields for updating a draft purpose.
type PurposeDraftUpdate struct {
	Title              *string
	Description        *string
	DailyCalls         *int32
	IsFreeOfCharge     *bool
	FreeOfChargeReason *string
}

// ListPurposesParams contains filter parameters for listing purposes.
type ListPurposesParams struct {
	EServiceIDs []uuid.UUID
	Title       *string
	ConsumerIDs []uuid.UUID
	States      []string
	Offset      int32
	Limit       int32
}

// PurposeVersionSeed contains fields for creating a new purpose version.
type PurposeVersionSeed struct {
	DailyCalls int32
}

// ClientInfo represents a PDND client (full visibility).
type ClientInfo struct {
	ID          uuid.UUID
	ConsumerID  uuid.UUID
	Name        string
	Description *string
	CreatedAt   time.Time
}

// ClientKey represents a JWK public key on a client.
type ClientKey struct {
	Kid string
	Kty string
	Alg *string
	Use *string
}

// ClientKeyDetail represents a key with its associated client ID (from create response).
type ClientKeyDetail struct {
	ClientID uuid.UUID
	Key      ClientKey
}

// ClientKeySeed contains fields for creating a new client key.
type ClientKeySeed struct {
	Key  string // Base64 PEM public key
	Use  string // SIG or ENC
	Alg  string
	Name string // 5-60 chars
}

// ClientsPage is a paginated list of clients.
type ClientsPage struct {
	Results    []ClientInfo
	Pagination Pagination
}

// ClientKeysPage is a paginated list of client keys.
type ClientKeysPage struct {
	Results    []ClientKey
	Pagination Pagination
}

// ListClientsParams contains filter parameters for listing clients.
type ListClientsParams struct {
	Name       *string
	ConsumerID *uuid.UUID
	Offset     int32
	Limit      int32
}

// TenantInfo represents a PDND tenant.
type TenantInfo struct {
	ID          uuid.UUID
	Name        string
	Kind        *string
	ExternalID  *TenantExternalID
	CreatedAt   time.Time
	UpdatedAt   *time.Time
	OnboardedAt *time.Time
	SubUnitType *string
}

// TenantExternalID represents a tenant's external identity.
type TenantExternalID struct {
	Origin string
	Value  string
}

// TenantsPage is a paginated list of tenants.
type TenantsPage struct {
	Results    []TenantInfo
	Pagination Pagination
}

// ListTenantsParams contains filter parameters for listing tenants.
type ListTenantsParams struct {
	IPACode *string
	TaxCode *string
	Offset  int32
	Limit   int32
}

// TenantCertifiedAttr represents a certified attribute assigned to a tenant.
type TenantCertifiedAttr struct {
	ID         uuid.UUID
	AssignedAt time.Time
	RevokedAt  *time.Time
}

// TenantDeclaredAttr represents a declared attribute assigned to a tenant.
type TenantDeclaredAttr struct {
	ID           uuid.UUID
	AssignedAt   time.Time
	RevokedAt    *time.Time
	DelegationID *uuid.UUID
}

// TenantVerifiedAttr represents a verified attribute assigned to a tenant.
type TenantVerifiedAttr struct {
	ID         uuid.UUID
	AssignedAt time.Time
}

// TenantCertifiedAttrsPage is a paginated list.
type TenantCertifiedAttrsPage struct {
	Results    []TenantCertifiedAttr
	Pagination Pagination
}

// TenantDeclaredAttrsPage is a paginated list.
type TenantDeclaredAttrsPage struct {
	Results    []TenantDeclaredAttr
	Pagination Pagination
}

// TenantVerifiedAttrsPage is a paginated list.
type TenantVerifiedAttrsPage struct {
	Results    []TenantVerifiedAttr
	Pagination Pagination
}
