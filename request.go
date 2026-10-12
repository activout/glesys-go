package glesys

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

const maxResponseBytes = 16 << 20

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

type statusEnvelope struct {
	Response struct {
		Status struct {
			Code int    `json:"code"`
			Text string `json:"text"`
		} `json:"status"`
	} `json:"response"`
}

// SetHTTPClient supplies a client without changing the process-wide default.
// Callers must configure redirect policy to protect credentials.
func (c *Client) SetHTTPClient(client *http.Client) { c.httpClient = client }

var endpointPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+/[A-Za-z0-9_-]+$`)
var parameterPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// ValidateEndpoint accepts only a relative service/operation, never a URL.
func ValidateEndpoint(endpoint string) error {
	if !endpointPattern.MatchString(endpoint) {
		return fmt.Errorf("endpoint must be a relative service/operation without URL syntax or traversal")
	}
	return nil
}

// Request returns the complete provider envelope through the existing transport.
// GET scalar parameters use the documented /name/value path convention.
func (c *Client) Request(ctx context.Context, method, endpoint string, params map[string]any) (json.RawMessage, error) {
	if err := ValidateEndpoint(endpoint); err != nil {
		return nil, err
	}
	method = strings.ToUpper(method)
	var data json.RawMessage
	switch method {
	case http.MethodPost:
		var payload any
		if params != nil {
			payload = params
		}
		err := c.post(ctx, endpoint, &data, payload)
		return data, err
	case http.MethodGet:
		keys := make([]string, 0, len(params))
		for k := range params {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if !parameterPattern.MatchString(k) {
				return nil, fmt.Errorf("invalid GET parameter name")
			}
			var value string
			switch v := params[k].(type) {
			case string:
				value = v
			case json.Number:
				value = v.String()
			case bool:
				if v {
					value = "true"
				} else {
					value = "false"
				}
			default:
				return nil, fmt.Errorf("GET parameters must be strings, JSON numbers or booleans")
			}
			if value == "." || value == ".." {
				return nil, fmt.Errorf("invalid GET parameter value")
			}
			endpoint += "/" + k + "/" + url.PathEscape(value)
		}
		err := c.get(ctx, endpoint, &data)
		return data, err
	default:
		return nil, fmt.Errorf("method must be GET or POST")
	}
}
