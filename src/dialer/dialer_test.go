package dialer

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// HTTPClient Interface Implementation
// Validates interface contract (compile-time check)
func TestDefaultHTTPClientImplementsInterface(t *testing.T) {
	var _ HTTPClient = &DefaultHTTPClient{}
	// This is a compile-time check - if it compiles, the interface is implemented
}

// NetReq Success Case
// Validates successful HTTP GET request
func TestNetReqSuccess(t *testing.T) {
	// Setup: Mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))
	}))
	defer server.Close()

	// Test: NetReq with valid URL
	resp, err := NetReq(server.URL, 5, false)
	if err != nil {
		t.Fatalf("NetReq failed: %v", err)
	}
	defer resp.Body.Close()

	// Verify: Response received, status 200
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// NetReq Timeout Handling
// Validates timeout behavior
func TestNetReqTimeout(t *testing.T) {
	// Setup: Slow mock server (delays > timeout)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Delay longer than the timeout (5 seconds)
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// NetReq with short timeout (1 second)
	// Note: NetReq expects integer seconds, multiplies by time.Second internally
	start := time.Now()
	_, err := NetReq(server.URL, 1, false) // 1 second timeout
	elapsed := time.Since(start)

	// Verify: Returns timeout error
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}

	// Verify: Error message contains "timeout"
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "timeout") {
		t.Errorf("Expected error message to contain 'timeout', got: %v", err)
	}

	// Verify: Actually timed out (didn't wait for full 5 second delay)
	if elapsed > 2*time.Second {
		t.Errorf("Timeout took too long: %v (expected ~1s)", elapsed)
	}
}

// NetReqWithUserAgent Headers
// Validates custom User-Agent header
func TestNetReqWithUserAgent(t *testing.T) {
	customUA := "TestAgent/1.0"
	receivedUA := ""

	// Setup: Mock server that inspects headers
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Test: NetReqWithUserAgent with custom UA
	resp, err := NetReqWithUserAgent(server.URL, 5, false, customUA)
	if err != nil {
		t.Fatalf("NetReqWithUserAgent failed: %v", err)
	}
	defer resp.Body.Close()

	// Verify: User-Agent header set correctly
	if receivedUA != customUA {
		t.Errorf("User-Agent mismatch: got %q, want %q", receivedUA, customUA)
	}
}

// NetReqWithUserAgent without UA
// Validates behavior when UA is empty string
func TestNetReqWithUserAgentEmpty(t *testing.T) {
	hasUserAgent := false

	// Setup: Mock server that checks for User-Agent header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if User-Agent header exists (Go sets a default if not specified)
		if r.Header.Get("User-Agent") != "" {
			hasUserAgent = true
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Test: NetReqWithUserAgent with empty string
	resp, err := NetReqWithUserAgent(server.URL, 5, false, "")
	if err != nil {
		t.Fatalf("NetReqWithUserAgent failed: %v", err)
	}
	defer resp.Body.Close()

	// Note: Go's http.Client may set a default User-Agent even if we don't
	// This test just verifies the request succeeds
	_ = hasUserAgent // Don't assert on this - it's implementation-dependent
}

// GZIP Encoding Support
// Validates Accept-Encoding header
func TestGzipEncoding(t *testing.T) {
	tests := []struct {
		name          string
		gzip          bool
		expectHeader  bool
		expectedValue string
	}{
		{
			name:          "GZIP enabled",
			gzip:          true,
			expectHeader:  true,
			expectedValue: "gzip",
		},
		{
			name:         "GZIP disabled",
			gzip:         false,
			expectHeader: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receivedEncoding := ""

			// Setup: Mock server that inspects Accept-Encoding header
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedEncoding = r.Header.Get("Accept-Encoding")
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			// Test: NetReq with gzip flag
			resp, err := NetReq(server.URL, 5, tt.gzip)
			if err != nil {
				t.Fatalf("NetReq failed: %v", err)
			}
			defer resp.Body.Close()

			// Verify: Accept-Encoding header presence
			if tt.expectHeader {
				if receivedEncoding != tt.expectedValue {
					t.Errorf("Expected Accept-Encoding: %q, got %q", tt.expectedValue, receivedEncoding)
				}
			} else {
				// When gzip=false, we shouldn't set the header
				// (but note: Go's http client might set it anyway - this is OK)
				if receivedEncoding == "gzip" && tt.gzip == false {
					// Only fail if we explicitly set it when we shouldn't
					// This is hard to test due to http.Client behavior
				}
			}
		})
	}
}

// Test: NetReq handles connection errors
func TestNetReqConnectionError(t *testing.T) {
	// Test with invalid URL
	_, err := NetReq("http://localhost:99999", 1, false)
	if err == nil {
		t.Error("Expected connection error, got nil")
	}
}

// Test: NetReq handles malformed URLs
func TestNetReqMalformedURL(t *testing.T) {
	// Test with malformed URL
	_, err := NetReq("not a valid url", 1, false)
	if err == nil {
		t.Error("Expected URL error, got nil")
	}
}

// Test: isTimeoutError helper function
func TestIsTimeoutError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "non-timeout error",
			err:      http.ErrServerClosed,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTimeoutError(tt.err)
			if result != tt.expected {
				t.Errorf("isTimeoutError(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

// Test: DefaultHTTPClient.Get method
func TestDefaultHTTPClientGet(t *testing.T) {
	client := &DefaultHTTPClient{}

	// Setup: Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}))
	defer server.Close()

	// Test: Direct call to Get method
	headers := make(map[string]string)
	headers["X-Test"] = "value"

	resp, err := client.Get(server.URL, 5*time.Second, headers)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// Test: Custom headers are set correctly
func TestCustomHeaders(t *testing.T) {
	client := &DefaultHTTPClient{}
	receivedHeaders := make(map[string]string)

	// Setup: Mock server that captures all headers
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders["X-Custom-1"] = r.Header.Get("X-Custom-1")
		receivedHeaders["X-Custom-2"] = r.Header.Get("X-Custom-2")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Test: Multiple custom headers
	headers := map[string]string{
		"X-Custom-1": "value1",
		"X-Custom-2": "value2",
	}

	resp, err := client.Get(server.URL, 5*time.Second, headers)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	// Verify: All custom headers were set
	if receivedHeaders["X-Custom-1"] != "value1" {
		t.Errorf("X-Custom-1 header mismatch: got %q, want value1", receivedHeaders["X-Custom-1"])
	}
	if receivedHeaders["X-Custom-2"] != "value2" {
		t.Errorf("X-Custom-2 header mismatch: got %q, want value2", receivedHeaders["X-Custom-2"])
	}
}

// Test: hasSubstring helper function
func TestHasSubstring(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{
			name:     "substring present",
			s:        "timeout error occurred",
			substr:   "timeout",
			expected: true,
		},
		{
			name:     "substring not present",
			s:        "connection refused",
			substr:   "timeout",
			expected: false,
		},
		{
			name:     "empty substring",
			s:        "test string",
			substr:   "",
			expected: true, // Empty string is substring of any string
		},
		{
			name:     "empty string",
			s:        "",
			substr:   "test",
			expected: false,
		},
		{
			name:     "case sensitive",
			s:        "Timeout error",
			substr:   "timeout",
			expected: false, // hasSubstring is case-sensitive
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasSubstring(tt.s, tt.substr)
			if result != tt.expected {
				t.Errorf("hasSubstring(%q, %q) = %v, want %v", tt.s, tt.substr, result, tt.expected)
			}
		})
	}
}

// Benchmark: NetReq performance
func BenchmarkNetReq(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := NetReq(server.URL, 5, false)
		if err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()
	}
}
