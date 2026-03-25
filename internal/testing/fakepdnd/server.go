package fakepdnd

import (
	"net/http"
	"net/http/httptest"
	"sync"

	"github.com/google/uuid"
)

// FakeServer is a deterministic, in-process fake PDND API server.
type FakeServer struct {
	mu             sync.RWMutex
	agreements     map[uuid.UUID]*StoredAgreement
	purposes       map[uuid.UUID][]StoredPurpose // keyed by agreementID
	approvalPolicy string                        // "AUTOMATIC" (default) or "MANUAL"
	producerID     uuid.UUID                     // fixed per server instance
	consumerID     uuid.UUID                     // fixed per server instance
	mux            *http.ServeMux
}

// NewFakeServer creates a new fake PDND server with default settings.
func NewFakeServer() *FakeServer {
	s := &FakeServer{
		agreements:     make(map[uuid.UUID]*StoredAgreement),
		purposes:       make(map[uuid.UUID][]StoredPurpose),
		approvalPolicy: "AUTOMATIC",
		producerID:     uuid.New(),
		consumerID:     uuid.New(),
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
}
