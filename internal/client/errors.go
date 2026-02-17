package client

import "fmt"

// APIError represents a structured error from the share API.
type APIError struct {
	StatusCode int
	Code       string `json:"code"`
	Message    string `json:"message"`
	RetryAfter int    `json:"retry_after,omitempty"`
}

func (e *APIError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("%s (retry after %ds)", e.Message, e.RetryAfter)
	}
	return e.Message
}

// IsRateLimit returns true if the error is a rate limit error.
func (e *APIError) IsRateLimit() bool {
	return e.StatusCode == 429
}

// IsPayloadTooLarge returns true if the uploaded content exceeds the size limit.
func (e *APIError) IsPayloadTooLarge() bool {
	return e.StatusCode == 413
}

// RateLimitInfo holds rate limit state from response headers.
type RateLimitInfo struct {
	Limit     int
	Remaining int
	Reset     int64
}
