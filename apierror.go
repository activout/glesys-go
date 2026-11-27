package glesys

import "fmt"

// APIError retains transport and provider status without exposing response bodies.
type APIError struct {
	HTTPStatus     int
	ProviderStatus int
	Message        string
	RetryAfter     string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("request failed with HTTP error: %d (%s)", e.HTTPStatus, e.Message)
}
