package client

import (
	"bytes"
	"io"
	"math"
	"math/rand"
	"net/http"
	"time"
)

// RetryTransport wraps another http.RoundTripper with retry logic for transient errors.
type RetryTransport struct {
	Base       http.RoundTripper
	MaxRetries int
	MinBackoff time.Duration
	MaxBackoff time.Duration
}

// RoundTrip implements http.RoundTripper with retry logic.
// It retries on 429, 502, 503, and 504 responses with exponential backoff and jitter.
func (t *RetryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	maxRetries := t.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}
	minBackoff := t.MinBackoff
	if minBackoff <= 0 {
		minBackoff = 500 * time.Millisecond
	}
	maxBackoff := t.MaxBackoff
	if maxBackoff <= 0 {
		maxBackoff = 10 * time.Second
	}

	// Buffer the request body so it can be replayed on retries.
	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		req.Body.Close()
	}

	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}

	var resp *http.Response
	var err error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			backoff := float64(minBackoff) * math.Pow(2, float64(attempt-1))
			if backoff > float64(maxBackoff) {
				backoff = float64(maxBackoff)
			}
			jitter := rand.Float64() * 0.5 * backoff
			time.Sleep(time.Duration(backoff + jitter))
		}

		// Reset the body for each attempt.
		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		} else {
			req.Body = nil
		}

		resp, err = base.RoundTrip(req)
		if err != nil {
			return nil, err
		}

		if !isRetryableStatus(resp.StatusCode) {
			return resp, nil
		}

		// Discard the response body before retrying.
		if resp.Body != nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}

	// All retries exhausted; return the last response.
	return resp, nil
}

func isRetryableStatus(status int) bool {
	switch status {
	case 429, 502, 503, 504:
		return true
	default:
		return false
	}
}
