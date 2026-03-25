package client

import (
	"fmt"
	"net/http"
)

// DPoPTransport is an http.RoundTripper that attaches DPoP authentication headers
// to every outgoing request.
type DPoPTransport struct {
	Base        http.RoundTripper
	AccessToken string
	ProofGen    *DPoPProofGenerator
}

// RoundTrip implements http.RoundTripper. It generates a DPoP proof for the request
// and attaches both the Authorization and DPoP headers.
func (t *DPoPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	proof, err := t.ProofGen.GenerateProof(req.Method, req.URL.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate DPoP proof: %w", err)
	}

	// Clone the request to avoid modifying the original.
	clone := req.Clone(req.Context())
	clone.Header.Set("Authorization", "DPoP "+t.AccessToken)
	clone.Header.Set("DPoP", proof)

	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}

	return base.RoundTrip(clone)
}
