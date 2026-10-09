// This package handles calls for all API requests.
package dialer

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

// UserAgent is sent with every request. NOAA requires one that identifies the app;
// main adds the version.
var UserAgent = "vaporwair (https://github.com/jeff-bruemmer/vaporwair)"

// Get sends a GET request that gives up after timeout.
// A failed request returns an error that names the host but never the URL,
// since URLs can carry API keys.
func Get(rawURL string, timeout time.Duration) (*http.Response, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("bad request URL: %w", stripURL(err))
	}
	req.Header.Set("User-Agent", UserAgent)

	resp, err := (&http.Client{Timeout: timeout}).Do(req)
	if err != nil {
		return nil, describe(req.URL.Host, timeout, stripURL(err))
	}
	return resp, nil
}

// stripURL drops the URL that net/http puts in its errors.
func stripURL(err error) error {
	var uerr *url.Error
	if errors.As(err, &uerr) {
		return uerr.Err
	}
	return err
}

// requestError is a network failure described for a person. It still unwraps to the
// underlying error, whose low-level text ("lookup ...: no such host") it leaves out.
type requestError struct {
	msg string
	err error
}

func (e *requestError) Error() string { return e.msg }
func (e *requestError) Unwrap() error { return e.err }

// describe rewrites a network failure as something a person can act on.
func describe(host string, timeout time.Duration, err error) error {
	var dnsErr *net.DNSError
	var netErr net.Error
	switch {
	case errors.As(err, &dnsErr):
		return &requestError{fmt.Sprintf("can't reach %s (no network connection?)", host), err}
	case errors.As(err, &netErr) && netErr.Timeout():
		return &requestError{fmt.Sprintf("%s did not respond within %v", host, timeout), err}
	default:
		return &requestError{fmt.Sprintf("can't connect to %s: %v", host, err), err}
	}
}
