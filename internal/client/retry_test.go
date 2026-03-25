package client_test

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lmammino/terraform-provider-pdnd/internal/client"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func makeResponse(statusCode int) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader("")),
		Header:     make(http.Header),
	}
}

func TestRetryTransport_RetryOn429(t *testing.T) {
	var calls int32
	mock := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		n := atomic.AddInt32(&calls, 1)
		if n <= 2 {
			return makeResponse(429), nil
		}
		return makeResponse(200), nil
	})

	transport := &client.RetryTransport{
		Base:       mock,
		MaxRetries: 3,
		MinBackoff: 1 * time.Millisecond,
		MaxBackoff: 5 * time.Millisecond,
	}

	req, _ := http.NewRequest("GET", "https://api.example.com/v3/test", nil)
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if atomic.LoadInt32(&calls) != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestRetryTransport_RetryOn502(t *testing.T) {
	var calls int32
	mock := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		n := atomic.AddInt32(&calls, 1)
		if n <= 1 {
			return makeResponse(502), nil
		}
		return makeResponse(200), nil
	})

	transport := &client.RetryTransport{
		Base:       mock,
		MaxRetries: 3,
		MinBackoff: 1 * time.Millisecond,
		MaxBackoff: 5 * time.Millisecond,
	}

	req, _ := http.NewRequest("GET", "https://api.example.com/v3/test", nil)
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Errorf("expected 2 calls, got %d", calls)
	}
}

func TestRetryTransport_RetryOn503(t *testing.T) {
	var calls int32
	mock := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		n := atomic.AddInt32(&calls, 1)
		if n <= 1 {
			return makeResponse(503), nil
		}
		return makeResponse(200), nil
	})

	transport := &client.RetryTransport{
		Base:       mock,
		MaxRetries: 3,
		MinBackoff: 1 * time.Millisecond,
		MaxBackoff: 5 * time.Millisecond,
	}

	req, _ := http.NewRequest("GET", "https://api.example.com/v3/test", nil)
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Errorf("expected 2 calls, got %d", calls)
	}
}

func TestRetryTransport_RetryOn504(t *testing.T) {
	var calls int32
	mock := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		n := atomic.AddInt32(&calls, 1)
		if n <= 1 {
			return makeResponse(504), nil
		}
		return makeResponse(200), nil
	})

	transport := &client.RetryTransport{
		Base:       mock,
		MaxRetries: 3,
		MinBackoff: 1 * time.Millisecond,
		MaxBackoff: 5 * time.Millisecond,
	}

	req, _ := http.NewRequest("GET", "https://api.example.com/v3/test", nil)
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Errorf("expected 2 calls, got %d", calls)
	}
}

func TestRetryTransport_NoRetryOn200(t *testing.T) {
	var calls int32
	mock := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		atomic.AddInt32(&calls, 1)
		return makeResponse(200), nil
	})

	transport := &client.RetryTransport{
		Base:       mock,
		MaxRetries: 3,
		MinBackoff: 1 * time.Millisecond,
		MaxBackoff: 5 * time.Millisecond,
	}

	req, _ := http.NewRequest("GET", "https://api.example.com/v3/test", nil)
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestRetryTransport_NoRetryOn400(t *testing.T) {
	var calls int32
	mock := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		atomic.AddInt32(&calls, 1)
		return makeResponse(400), nil
	})

	transport := &client.RetryTransport{
		Base:       mock,
		MaxRetries: 3,
		MinBackoff: 1 * time.Millisecond,
		MaxBackoff: 5 * time.Millisecond,
	}

	req, _ := http.NewRequest("GET", "https://api.example.com/v3/test", nil)
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestRetryTransport_NoRetryOn401(t *testing.T) {
	var calls int32
	mock := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		atomic.AddInt32(&calls, 1)
		return makeResponse(401), nil
	})

	transport := &client.RetryTransport{
		Base:       mock,
		MaxRetries: 3,
		MinBackoff: 1 * time.Millisecond,
		MaxBackoff: 5 * time.Millisecond,
	}

	req, _ := http.NewRequest("GET", "https://api.example.com/v3/test", nil)
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 401 {
		t.Errorf("expected status 401, got %d", resp.StatusCode)
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestRetryTransport_NoRetryOn404(t *testing.T) {
	var calls int32
	mock := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		atomic.AddInt32(&calls, 1)
		return makeResponse(404), nil
	})

	transport := &client.RetryTransport{
		Base:       mock,
		MaxRetries: 3,
		MinBackoff: 1 * time.Millisecond,
		MaxBackoff: 5 * time.Millisecond,
	}

	req, _ := http.NewRequest("GET", "https://api.example.com/v3/test", nil)
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 404 {
		t.Errorf("expected status 404, got %d", resp.StatusCode)
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestRetryTransport_NoRetryOn500(t *testing.T) {
	var calls int32
	mock := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		atomic.AddInt32(&calls, 1)
		return makeResponse(500), nil
	})

	transport := &client.RetryTransport{
		Base:       mock,
		MaxRetries: 3,
		MinBackoff: 1 * time.Millisecond,
		MaxBackoff: 5 * time.Millisecond,
	}

	req, _ := http.NewRequest("GET", "https://api.example.com/v3/test", nil)
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 500 {
		t.Errorf("expected status 500, got %d", resp.StatusCode)
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestRetryTransport_MaxRetriesExhausted(t *testing.T) {
	var calls int32
	mock := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		atomic.AddInt32(&calls, 1)
		return makeResponse(429), nil
	})

	maxRetries := 3
	transport := &client.RetryTransport{
		Base:       mock,
		MaxRetries: maxRetries,
		MinBackoff: 1 * time.Millisecond,
		MaxBackoff: 5 * time.Millisecond,
	}

	req, _ := http.NewRequest("GET", "https://api.example.com/v3/test", nil)
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have made maxRetries+1 total calls (1 initial + maxRetries retries).
	expectedCalls := int32(maxRetries + 1)
	if atomic.LoadInt32(&calls) != expectedCalls {
		t.Errorf("expected %d calls, got %d", expectedCalls, calls)
	}
	if resp.StatusCode != 429 {
		t.Errorf("expected status 429 after exhausting retries, got %d", resp.StatusCode)
	}
}

func TestRetryTransport_BodyReplayable(t *testing.T) {
	var calls int32
	var bodies []string

	mock := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		n := atomic.AddInt32(&calls, 1)
		if req.Body != nil {
			bodyBytes, _ := io.ReadAll(req.Body)
			bodies = append(bodies, string(bodyBytes))
		}
		if n <= 1 {
			return makeResponse(429), nil
		}
		return makeResponse(200), nil
	})

	transport := &client.RetryTransport{
		Base:       mock,
		MaxRetries: 3,
		MinBackoff: 1 * time.Millisecond,
		MaxBackoff: 5 * time.Millisecond,
	}

	body := "test request body"
	req, _ := http.NewRequest("POST", "https://api.example.com/v3/test", bytes.NewBufferString(body))
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Errorf("expected 2 calls, got %d", calls)
	}
	if len(bodies) != 2 {
		t.Fatalf("expected 2 body reads, got %d", len(bodies))
	}
	if bodies[0] != body || bodies[1] != body {
		t.Errorf("body was not replayed correctly: first=%q, second=%q", bodies[0], bodies[1])
	}
}
