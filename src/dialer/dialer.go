// This package handles calls for all API requests.
package dialer

import (
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
		return nil, err
	}

	// Set custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return client.Do(req)
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
