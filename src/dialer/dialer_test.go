package dialer

import (
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGetSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))
	}))
	defer server.Close()

	resp, err := Get(server.URL, 5*time.Second)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// Every request identifies vaporwair; NOAA rejects requests without a User-Agent.
func TestGetSendsUserAgent(t *testing.T) {
	received := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header.Get("User-Agent")
	}))
	defer server.Close()

	old := UserAgent
	UserAgent = "vaporwair/test"
	defer func() { UserAgent = old }()

	resp, err := Get(server.URL, 5*time.Second)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	resp.Body.Close()
	if received != "vaporwair/test" {
		t.Errorf("User-Agent = %q, want %q", received, "vaporwair/test")
	}
}

func TestGetTimeout(t *testing.T) {
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-release
	}))
	defer server.Close()
	defer close(release)

	start := time.Now()
	_, err := Get(server.URL, 200*time.Millisecond)
	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "did not respond within 200ms") {
		t.Errorf("Timeout error = %q, want it to say the host did not respond", err)
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Errorf("Timeout took too long: %v", elapsed)
	}
}

// closedPortURL returns a URL on localhost where nothing is listening.
func closedPortURL(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := l.Addr().String()
	l.Close()
	return "http://" + addr
}

// AirNow's API key travels in the query string; a failed request must not print it.
func TestGetErrorOmitsURL(t *testing.T) {
	_, err := Get(closedPortURL(t)+"/aq/?zipCode=05401&API_KEY=SECRET-KEY", 2*time.Second)
	if err == nil {
		t.Fatal("Expected connection error, got nil")
	}
	if strings.Contains(err.Error(), "SECRET-KEY") || strings.Contains(err.Error(), "API_KEY") {
		t.Errorf("error leaks the URL: %q", err)
	}
	if !strings.HasPrefix(err.Error(), "can't connect to 127.0.0.1:") {
		t.Errorf("error = %q, want it to name the host", err)
	}
}

func TestGetUnknownHost(t *testing.T) {
	_, err := Get("http://vaporwair.invalid/", 5*time.Second)
	if err == nil {
		t.Fatal("Expected DNS error, got nil")
	}
	if !strings.HasPrefix(err.Error(), "can't reach vaporwair.invalid") {
		t.Errorf("error = %q, want \"can't reach vaporwair.invalid ...\"", err)
	}
	var dnsErr *net.DNSError
	if !errors.As(err, &dnsErr) {
		t.Error("error should still unwrap to *net.DNSError")
	}
}

func TestGetMalformedURL(t *testing.T) {
	if _, err := Get("not a valid url", time.Second); err == nil {
		t.Error("Expected URL error, got nil")
	}
}

func BenchmarkGet(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := Get(server.URL, 5*time.Second)
		if err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()
	}
}
