// This package handles calls for all API requests.
package dialer

import (
	"net/http"
	"time"
)

// NetReq returns an *http.Response, or times out after a specified duration.
func NetReq(url string, s time.Duration, gzip bool) (*http.Response, error) {
	t := time.Duration(s * time.Second)
	c := http.Client{
		Timeout: t,
	}
	req, _ := http.NewRequest("GET", url, nil)
	// Dark Sky uses gzip
	if gzip {
		req.Header.Set("Accept-Encoding", "gzip")
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// NetReqWithUserAgent returns an *http.Response with a custom User-Agent header,
// or times out after a specified duration.
func NetReqWithUserAgent(url string, s time.Duration, gzip bool, userAgent string) (*http.Response, error) {
	t := time.Duration(s * time.Second)
	c := http.Client{
		Timeout: t,
	}
	req, _ := http.NewRequest("GET", url, nil)
	// Set User-Agent header (required for NOAA API)
	if userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}
	if gzip {
		req.Header.Set("Accept-Encoding", "gzip")
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
