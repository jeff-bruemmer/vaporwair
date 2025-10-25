// This package handles calls for all API requests.
package dialer

import (
	"fmt"
	"net/http"
	"time"
)

// HTTPClient defines an interface for making HTTP requests.
// This allows for easy mocking in tests.
type HTTPClient interface {
	Get(url string, timeout time.Duration, headers map[string]string) (*http.Response, error)
}

// DefaultHTTPClient implements HTTPClient using the standard http.Client.
type DefaultHTTPClient struct{}

// Get performs an HTTP GET request with the specified timeout and headers.
func (d *DefaultHTTPClient) Get(url string, timeout time.Duration, headers map[string]string) (*http.Response, error) {
	client := &http.Client{
		Timeout: timeout,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		// Provide more context for common network errors
		if isTimeoutError(err) {
			return nil, fmt.Errorf("request timed out after %v: %w", timeout, err)
		}
		return nil, fmt.Errorf("network request failed: %w", err)
	}

	return resp, nil
}

// isTimeoutError checks if the error is a timeout error
func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	// Check for timeout in error message
	return err.Error() != "" && (hasSubstring(err.Error(), "timeout") || hasSubstring(err.Error(), "Timeout"))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// DefaultClient is the default HTTP client used by the application.
var DefaultClient HTTPClient = &DefaultHTTPClient{}

// NetReq returns an *http.Response, or times out after a specified duration.
// Uses the default HTTP client.
func NetReq(url string, s time.Duration, gzip bool) (*http.Response, error) {
	headers := make(map[string]string)
	if gzip {
		headers["Accept-Encoding"] = "gzip"
	}
	return DefaultClient.Get(url, s*time.Second, headers)
}

// NetReqWithUserAgent returns an *http.Response with a custom User-Agent header,
// or times out after a specified duration.
// Uses the default HTTP client.
func NetReqWithUserAgent(url string, s time.Duration, gzip bool, userAgent string) (*http.Response, error) {
	headers := make(map[string]string)
	if userAgent != "" {
		headers["User-Agent"] = userAgent
	}
	if gzip {
		headers["Accept-Encoding"] = "gzip"
	}
	return DefaultClient.Get(url, s*time.Second, headers)
}
