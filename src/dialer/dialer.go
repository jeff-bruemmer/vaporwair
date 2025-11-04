// This package handles calls for all API requests.
package dialer

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// isTimeoutError checks if the error is a timeout error
func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return errMsg != "" && (strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "Timeout"))
}

// NetReq returns an *http.Response, or times out after a specified duration.
func NetReq(url string, s time.Duration, gzip bool) (*http.Response, error) {
	headers := make(map[string]string)
	if gzip {
		headers["Accept-Encoding"] = "gzip"
	}

	client := &http.Client{Timeout: s * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		if isTimeoutError(err) {
			return nil, fmt.Errorf("request timed out after %v: %w", s*time.Second, err)
		}
		return nil, fmt.Errorf("network request failed: %w", err)
	}

	return resp, nil
}

// NetReqWithUserAgent returns an *http.Response with a custom User-Agent header,
// or times out after a specified duration.
func NetReqWithUserAgent(url string, s time.Duration, gzip bool, userAgent string) (*http.Response, error) {
	headers := make(map[string]string)
	if userAgent != "" {
		headers["User-Agent"] = userAgent
	}
	if gzip {
		headers["Accept-Encoding"] = "gzip"
	}

	client := &http.Client{Timeout: s * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		if isTimeoutError(err) {
			return nil, fmt.Errorf("request timed out after %v: %w", s*time.Second, err)
		}
		return nil, fmt.Errorf("network request failed: %w", err)
	}

	return resp, nil
}
