package fakepdnd

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"

	"github.com/google/uuid"
)

// FakeServer is a deterministic, in-process fake PDND API server.
type FakeServer struct {
	mu                   sync.RWMutex
	agreements           map[uuid.UUID]*StoredAgreement
	purposes             map[uuid.UUID][]StoredPurpose                 // keyed by agreementID
	eservices            map[uuid.UUID]*StoredEService
	descriptors          map[uuid.UUID]map[uuid.UUID]*StoredDescriptor // eserviceID -> descriptorID -> descriptor
	descriptorCount      map[uuid.UUID]int                             // eserviceID -> next version number
	certifiedAttributes  map[uuid.UUID]*StoredCertifiedAttribute
	declaredAttributes   map[uuid.UUID]*StoredDeclaredAttribute
	verifiedAttributes   map[uuid.UUID]*StoredVerifiedAttribute
	// descriptor attribute groups: eserviceID -> descriptorID -> groups
	descCertifiedAttrGroups map[uuid.UUID]map[uuid.UUID][]StoredDescriptorAttributeGroup
	descDeclaredAttrGroups  map[uuid.UUID]map[uuid.UUID][]StoredDescriptorAttributeGroup
	descVerifiedAttrGroups  map[uuid.UUID]map[uuid.UUID][]StoredDescriptorAttributeGroup
	standalonePurposes      map[uuid.UUID]*StoredPurpose // keyed by purposeID
	purposeApprovalThreshold int32                       // daily calls threshold for auto vs manual approval
	approvalPolicy       string    // "AUTOMATIC" (default) or "MANUAL"
	producerID           uuid.UUID // fixed per server instance
	consumerID           uuid.UUID // fixed per server instance
	mux                  *http.ServeMux
}

// NewFakeServer creates a new fake PDND server with default settings.
func NewFakeServer() *FakeServer {
	s := &FakeServer{
		agreements:              make(map[uuid.UUID]*StoredAgreement),
		purposes:                make(map[uuid.UUID][]StoredPurpose),
		eservices:               make(map[uuid.UUID]*StoredEService),
		descriptors:             make(map[uuid.UUID]map[uuid.UUID]*StoredDescriptor),
		descriptorCount:         make(map[uuid.UUID]int),
		certifiedAttributes:     make(map[uuid.UUID]*StoredCertifiedAttribute),
		declaredAttributes:      make(map[uuid.UUID]*StoredDeclaredAttribute),
		verifiedAttributes:      make(map[uuid.UUID]*StoredVerifiedAttribute),
		descCertifiedAttrGroups: make(map[uuid.UUID]map[uuid.UUID][]StoredDescriptorAttributeGroup),
		descDeclaredAttrGroups:  make(map[uuid.UUID]map[uuid.UUID][]StoredDescriptorAttributeGroup),
		descVerifiedAttrGroups:   make(map[uuid.UUID]map[uuid.UUID][]StoredDescriptorAttributeGroup),
		standalonePurposes:       make(map[uuid.UUID]*StoredPurpose),
		purposeApprovalThreshold: 10000,
		approvalPolicy:           "AUTOMATIC",
		producerID:              uuid.New(),
		consumerID:              uuid.New(),
	}
	s.setupRoutes()
	return s
}

// Start starts the server and returns an httptest.Server.
// The caller must call Close() on the returned server when done.
func (s *FakeServer) Start() *httptest.Server {
	return httptest.NewServer(s.mux)
}

// SetApprovalPolicy sets whether submit transitions to ACTIVE or PENDING.
// Valid values: "AUTOMATIC" (default), "MANUAL"
func (s *FakeServer) SetApprovalPolicy(policy string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.approvalPolicy = policy
}

// SeedAgreement pre-populates an agreement in the store.
func (s *FakeServer) SeedAgreement(a StoredAgreement) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.agreements[a.ID] = &a
}

// SeedPurpose adds a purpose associated with an agreement.
func (s *FakeServer) SeedPurpose(agreementID uuid.UUID, p StoredPurpose) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.purposes[agreementID] = append(s.purposes[agreementID], p)
}

// GetAgreement returns the current state of an agreement (for test assertions).
func (s *FakeServer) GetAgreement(id uuid.UUID) *StoredAgreement {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.agreements[id]
}

// ProducerID returns the fixed producer ID used by this server.
func (s *FakeServer) ProducerID() uuid.UUID {
	return s.producerID
}

// ConsumerID returns the fixed consumer ID used by this server.
func (s *FakeServer) ConsumerID() uuid.UUID {
	return s.consumerID
}

// SeedEService pre-populates an e-service in the store.
func (s *FakeServer) SeedEService(e StoredEService) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.eservices[e.ID] = &e
	if _, ok := s.descriptors[e.ID]; !ok {
		s.descriptors[e.ID] = make(map[uuid.UUID]*StoredDescriptor)
	}
	if _, ok := s.descriptorCount[e.ID]; !ok {
		s.descriptorCount[e.ID] = 0
	}
}

// SeedDescriptor pre-populates a descriptor in the store.
func (s *FakeServer) SeedDescriptor(d StoredDescriptor) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.descriptors[d.EServiceID]; !ok {
		s.descriptors[d.EServiceID] = make(map[uuid.UUID]*StoredDescriptor)
	}
	s.descriptors[d.EServiceID][d.ID] = &d
	// Track version count.
	current := s.descriptorCount[d.EServiceID]
	// Parse version to int and update count if higher.
	if v, err := strconv.Atoi(d.Version); err == nil && v >= current {
		s.descriptorCount[d.EServiceID] = v
	}
}

// GetEService returns the current state of an e-service (for test assertions).
func (s *FakeServer) GetEService(id uuid.UUID) *StoredEService {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.eservices[id]
}

// GetDescriptor returns the current state of a descriptor (for test assertions).
func (s *FakeServer) GetDescriptor(eserviceID, descriptorID uuid.UUID) *StoredDescriptor {
	s.mu.RLock()
	defer s.mu.RUnlock()
	descs := s.descriptors[eserviceID]
	if descs == nil {
		return nil
	}
	return descs[descriptorID]
}

// SeedCertifiedAttribute pre-populates a certified attribute in the store.
func (s *FakeServer) SeedCertifiedAttribute(a StoredCertifiedAttribute) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.certifiedAttributes[a.ID] = &a
}

// SeedDeclaredAttribute pre-populates a declared attribute in the store.
func (s *FakeServer) SeedDeclaredAttribute(a StoredDeclaredAttribute) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.declaredAttributes[a.ID] = &a
}

// SeedVerifiedAttribute pre-populates a verified attribute in the store.
func (s *FakeServer) SeedVerifiedAttribute(a StoredVerifiedAttribute) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.verifiedAttributes[a.ID] = &a
}

// SetPurposeApprovalThreshold sets the daily calls threshold above which activation requires approval.
func (s *FakeServer) SetPurposeApprovalThreshold(threshold int32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.purposeApprovalThreshold = threshold
}

// SeedStandalonePurpose pre-populates a standalone purpose in the store.
func (s *FakeServer) SeedStandalonePurpose(p StoredPurpose) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.standalonePurposes[p.ID] = &p
}

// GetStandalonePurpose returns the current state of a standalone purpose (for test assertions).
func (s *FakeServer) GetStandalonePurpose(id uuid.UUID) *StoredPurpose {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.standalonePurposes[id]
}

// SeedDescriptorAttributeGroups pre-populates attribute groups for a descriptor.
func (s *FakeServer) SeedDescriptorAttributeGroups(eserviceID, descriptorID uuid.UUID, attrType string, groups []StoredDescriptorAttributeGroup) {
	s.mu.Lock()
	defer s.mu.Unlock()
	store := s.getAttrGroupStore(attrType)
	if store[eserviceID] == nil {
		store[eserviceID] = make(map[uuid.UUID][]StoredDescriptorAttributeGroup)
	}
	store[eserviceID][descriptorID] = groups
}

// GetDescriptorAttributeGroups returns stored attribute groups for test assertions.
func (s *FakeServer) GetDescriptorAttributeGroups(eserviceID, descriptorID uuid.UUID, attrType string) []StoredDescriptorAttributeGroup {
	s.mu.RLock()
	defer s.mu.RUnlock()
	store := s.getAttrGroupStore(attrType)
	if store[eserviceID] == nil {
		return nil
	}
	return store[eserviceID][descriptorID]
}

func (s *FakeServer) getAttrGroupStore(attrType string) map[uuid.UUID]map[uuid.UUID][]StoredDescriptorAttributeGroup {
	switch attrType {
	case "certified":
		return s.descCertifiedAttrGroups
	case "declared":
		return s.descDeclaredAttrGroups
	case "verified":
		return s.descVerifiedAttrGroups
	default:
		return nil
	}
}

func (s *FakeServer) setupRoutes() {
	s.mux = http.NewServeMux()
	s.mux.HandleFunc("POST /agreements", s.handleCreateAgreement)
	s.mux.HandleFunc("GET /agreements", s.handleListAgreements)
	s.mux.HandleFunc("GET /agreements/{agreementId}", s.handleGetAgreement)
	s.mux.HandleFunc("DELETE /agreements/{agreementId}", s.handleDeleteAgreement)
	s.mux.HandleFunc("POST /agreements/{agreementId}/submit", s.handleSubmitAgreement)
	s.mux.HandleFunc("POST /agreements/{agreementId}/approve", s.handleApproveAgreement)
	s.mux.HandleFunc("POST /agreements/{agreementId}/reject", s.handleRejectAgreement)
	s.mux.HandleFunc("POST /agreements/{agreementId}/suspend", s.handleSuspendAgreement)
	s.mux.HandleFunc("POST /agreements/{agreementId}/unsuspend", s.handleUnsuspendAgreement)
	s.mux.HandleFunc("POST /agreements/{agreementId}/upgrade", s.handleUpgradeAgreement)
	s.mux.HandleFunc("POST /agreements/{agreementId}/clone", s.handleCloneAgreement)
	s.mux.HandleFunc("GET /agreements/{agreementId}/purposes", s.handleGetAgreementPurposes)

	// E-Service routes.
	s.mux.HandleFunc("POST /eservices", s.handleCreateEService)
	s.mux.HandleFunc("GET /eservices", s.handleListEServices)
	s.mux.HandleFunc("GET /eservices/{eserviceId}", s.handleGetEService)
	s.mux.HandleFunc("PATCH /eservices/{eserviceId}", s.handleUpdateDraftEService)
	s.mux.HandleFunc("DELETE /eservices/{eserviceId}", s.handleDeleteEService)
	s.mux.HandleFunc("PATCH /eservices/{eserviceId}/name", s.handleUpdatePublishedEServiceName)
	s.mux.HandleFunc("PATCH /eservices/{eserviceId}/description", s.handleUpdatePublishedEServiceDescription)
	s.mux.HandleFunc("PATCH /eservices/{eserviceId}/delegation", s.handleUpdatePublishedEServiceDelegation)
	s.mux.HandleFunc("PATCH /eservices/{eserviceId}/signalHub", s.handleUpdatePublishedEServiceSignalHub)

	// Descriptor routes.
	s.mux.HandleFunc("POST /eservices/{eserviceId}/descriptors", s.handleCreateDescriptor)
	s.mux.HandleFunc("GET /eservices/{eserviceId}/descriptors", s.handleListDescriptors)
	s.mux.HandleFunc("GET /eservices/{eserviceId}/descriptors/{descriptorId}", s.handleGetDescriptor)
	s.mux.HandleFunc("PATCH /eservices/{eserviceId}/descriptors/{descriptorId}", s.handleUpdateDraftDescriptor)
	s.mux.HandleFunc("DELETE /eservices/{eserviceId}/descriptors/{descriptorId}", s.handleDeleteDraftDescriptor)
	s.mux.HandleFunc("PATCH /eservices/{eserviceId}/descriptors/{descriptorId}/quotas", s.handleUpdatePublishedDescriptorQuotas)
	s.mux.HandleFunc("POST /eservices/{eserviceId}/descriptors/{descriptorId}/publish", s.handlePublishDescriptor)
	s.mux.HandleFunc("POST /eservices/{eserviceId}/descriptors/{descriptorId}/suspend", s.handleSuspendDescriptor)
	s.mux.HandleFunc("POST /eservices/{eserviceId}/descriptors/{descriptorId}/unsuspend", s.handleUnsuspendDescriptor)
	s.mux.HandleFunc("POST /eservices/{eserviceId}/descriptors/{descriptorId}/approve", s.handleApproveDelegatedDescriptor)

	// Purpose routes.
	s.mux.HandleFunc("POST /purposes", s.handleCreatePurpose)
	s.mux.HandleFunc("GET /purposes/{purposeId}", s.handleGetPurpose)
	s.mux.HandleFunc("DELETE /purposes/{purposeId}", s.handleDeletePurpose)
	s.mux.HandleFunc("PATCH /purposes/{purposeId}", s.handleUpdateDraftPurpose)
	s.mux.HandleFunc("POST /purposes/{purposeId}/activate", s.handleActivatePurpose)
	s.mux.HandleFunc("POST /purposes/{purposeId}/approve", s.handleApprovePurpose)
	s.mux.HandleFunc("POST /purposes/{purposeId}/suspend", s.handleSuspendPurpose)
	s.mux.HandleFunc("POST /purposes/{purposeId}/unsuspend", s.handleUnsuspendPurpose)
	s.mux.HandleFunc("POST /purposes/{purposeId}/archive", s.handleArchivePurpose)
	s.mux.HandleFunc("POST /purposes/{purposeId}/versions", s.handleCreatePurposeVersion)

	// Descriptor attribute group routes.
	for _, at := range []string{"certified", "declared", "verified"} {
		attrType := at // capture for closure
		prefix := "/eservices/{eserviceId}/descriptors/{descriptorId}/" + attrType + "Attributes"
		s.mux.HandleFunc("GET "+prefix, s.makeHandleListDescriptorAttributes(attrType))
		s.mux.HandleFunc("POST "+prefix+"/groups", s.makeHandleCreateDescriptorAttributeGroup(attrType))
		s.mux.HandleFunc("POST "+prefix+"/groups/{groupIndex}/attributes", s.makeHandleAssignDescriptorAttributesToGroup(attrType))
		s.mux.HandleFunc("DELETE "+prefix+"/groups/{groupIndex}/attributes/{attributeId}", s.makeHandleDeleteDescriptorAttributeFromGroup(attrType))
	}

	// Attribute routes.
	s.mux.HandleFunc("GET /certifiedAttributes", s.handleListCertifiedAttributes)
	s.mux.HandleFunc("GET /certifiedAttributes/{attributeId}", s.handleGetCertifiedAttribute)
	s.mux.HandleFunc("GET /declaredAttributes", s.handleListDeclaredAttributes)
	s.mux.HandleFunc("GET /declaredAttributes/{attributeId}", s.handleGetDeclaredAttribute)
	s.mux.HandleFunc("GET /verifiedAttributes", s.handleListVerifiedAttributes)
	s.mux.HandleFunc("GET /verifiedAttributes/{attributeId}", s.handleGetVerifiedAttribute)
}
