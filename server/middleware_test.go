package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

var browserHeaderTestCases = []struct {
	Description    string
	HeaderName     string
	HeaderValue    string
	ExpectedStatus int
}{
	{
		Description:    "valid request",
		HeaderName:     "User-Agent",
		HeaderValue:    "boto3/foo",
		ExpectedStatus: http.StatusOK,
	},
	{
		Description:    "Mozilla in user-agent",
		HeaderName:     "User-Agent",
		HeaderValue:    "Mozilla",
		ExpectedStatus: http.StatusForbidden,
	},
	{
		Description:    "mozilla in user-agent",
		HeaderName:     "User-Agent",
		HeaderValue:    "mozilla",
		ExpectedStatus: http.StatusForbidden,
	},
	{
		Description:    "referrer header set",
		HeaderName:     "Referrer",
		HeaderValue:    "anything",
		ExpectedStatus: http.StatusForbidden,
	},
	{
		Description:    "empty referrer header set",
		HeaderName:     "Referrer",
		HeaderValue:    "",
		ExpectedStatus: http.StatusForbidden,
	},
	{
		Description:    "origin header set",
		HeaderName:     "Origin",
		HeaderValue:    "anything",
		ExpectedStatus: http.StatusForbidden,
	},
	{
		Description:    "empty origin header set",
		HeaderName:     "Origin",
		HeaderValue:    "",
		ExpectedStatus: http.StatusForbidden,
	},
	{
		Description:    "host header not in allowlist",
		HeaderName:     "Host",
		HeaderValue:    "netflix.com",
		ExpectedStatus: http.StatusForbidden,
	},
	{
		Description:    "host header in allowlist (127.0.0.1)",
		HeaderName:     "Host",
		HeaderValue:    "localhost",
		ExpectedStatus: http.StatusOK,
	},
	{
		Description:    "host header in allowlist (127.0.0.1)",
		HeaderName:     "Host",
		HeaderValue:    "127.0.0.1",
		ExpectedStatus: http.StatusOK,
	},
	{
		Description:    "host header in allowlist (169.254.169.254)",
		HeaderName:     "Host",
		HeaderValue:    "169.254.169.254",
		ExpectedStatus: http.StatusOK,
	},
}

// TestBrowserFilterMiddleware ensures 403 Forbidden is returned for all requests that look like
// they came from a web browser.
func TestBrowserFilterMiddleware(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	for i, tc := range browserHeaderTestCases {
		t.Logf("test case %d: %s", i, tc.Description)
		bfmHandler := BrowserFilterMiddleware(nextHandler)
		req := httptest.NewRequest("GET", "http://localhost", nil)
		req.Header.Add(tc.HeaderName, tc.HeaderValue)
		rec := httptest.NewRecorder()
		bfmHandler.ServeHTTP(rec, req)
		if rec.Code != tc.ExpectedStatus {
			t.Errorf("%s failed: got status %d, expected %d", tc.Description, rec.Code, tc.ExpectedStatus)
			continue
		}
	}
}

// TestAWSHeaderMiddleware checks for headers added for consumption by AWS SDKs
func TestAWSHeaderMiddleware(t *testing.T) {
	description := "aws header middleware"
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	t.Logf("test case: %s", description)
	bfmHandler := CredentialServiceMiddleware(nextHandler)
	req := httptest.NewRequest("GET", "http://localhost", nil)
	rec := httptest.NewRecorder()
	bfmHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("%s failed: got status %d, expected %d", description, rec.Code, http.StatusOK)
	}
	if etag := rec.Header().Get("ETag"); etag == "" {
		t.Errorf("%s failed: ETag header not set", description)
	}
	if lastModified := rec.Header().Get("Last-Modified"); lastModified == "" {
		t.Errorf("%s failed: Last-Modified header not set", description)
	}
	if server := rec.Header().Get("Server"); server != "EC2ws" {
		t.Errorf("%s failed: got Server header %s, expected %s", description, server, "EC2ws")
	}
	if contentType := rec.Header().Get("Content-Type"); contentType != "text/plain" {
		t.Errorf("%s failed: got Content-Type header %s, expected %s", description, contentType, "text/plain")
	}
}

// TestCredentialServiceMiddleware is a superset of TestBrowserFilterMiddleware and TestAWSHeaderMiddleware
// since CredentialServiceMiddleware is a chain of BrowserFilterMiddleware and AWSHeaderMiddleware
func TestCredentialServiceMiddleware(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	for i, tc := range browserHeaderTestCases {
		t.Logf("test case %d: %s", i, tc.Description)
		bfmHandler := CredentialServiceMiddleware(nextHandler)
		req := httptest.NewRequest("GET", "http://localhost", nil)
		req.Header.Add(tc.HeaderName, tc.HeaderValue)
		rec := httptest.NewRecorder()
		bfmHandler.ServeHTTP(rec, req)
		if rec.Code != tc.ExpectedStatus {
			t.Errorf("%s failed: got status %d, expected %d", tc.Description, rec.Code, tc.ExpectedStatus)
			continue
		}
		if rec.Code == http.StatusOK {
			if etag := rec.Header().Get("ETag"); etag == "" {
				t.Errorf("%s failed: ETag header not set", tc.Description)
			}
			if lastModified := rec.Header().Get("Last-Modified"); lastModified == "" {
				t.Errorf("%s failed: Last-Modified header not set", tc.Description)
			}
			if server := rec.Header().Get("Server"); server != "EC2ws" {
				t.Errorf("%s failed: got Server header %s, expected %s", tc.Description, server, "EC2ws")
			}
			if contentType := rec.Header().Get("Content-Type"); contentType != "text/plain" {
				t.Errorf("%s failed: got Content-Type header %s, expected %s", tc.Description, contentType, "text/plain")
			}
		}
	}
}
