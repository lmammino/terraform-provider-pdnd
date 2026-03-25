package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// APIError represents a PDND API error response (RFC 7807 Problem Details).
type APIError struct {
	StatusCode    int
	ProblemType   string
	Title         string
	Detail        string
	CorrelationID string
	Errors        []ProblemError
}

// ProblemError represents a single error entry in a Problem Details response.
type ProblemError struct {
	Code   string `json:"code"`
	Detail string `json:"detail"`
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("PDND API error (HTTP %d): %s - %s", e.StatusCode, e.Title, e.Detail)
	}
	return fmt.Sprintf("PDND API error (HTTP %d): %s", e.StatusCode, e.Title)
}

// IsNotFound returns true if the error is a 404 Not Found.
func IsNotFound(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 404
	}
	return false
}

// IsConflict returns true if the error is a 409 Conflict.
func IsConflict(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 409
	}
	return false
}

// problemJSON is the internal struct for parsing Problem Details JSON.
type problemJSON struct {
	Type          string         `json:"type"`
	Status        int            `json:"status"`
	Title         string         `json:"title"`
	Detail        string         `json:"detail"`
	CorrelationID string         `json:"correlationId"`
	Errors        []ProblemError `json:"errors"`
}

// CheckResponse reads the response and returns an APIError if status is not 2xx.
// The response body is read and closed for non-2xx responses.
// For 2xx responses, returns nil and leaves the body unread.
func CheckResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	defer resp.Body.Close() //nolint:errcheck // best-effort close
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &APIError{
			StatusCode: resp.StatusCode,
			Title:      http.StatusText(resp.StatusCode),
		}
	}

	var problem problemJSON
	if err := json.Unmarshal(body, &problem); err != nil {
		return &APIError{
			StatusCode: resp.StatusCode,
			Title:      http.StatusText(resp.StatusCode),
		}
	}

	return &APIError{
		StatusCode:    resp.StatusCode,
		ProblemType:   problem.Type,
		Title:         problem.Title,
		Detail:        problem.Detail,
		CorrelationID: problem.CorrelationID,
		Errors:        problem.Errors,
	}
}
