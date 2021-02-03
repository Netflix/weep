package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBrowserFilterMiddleware(t *testing.T) {
	cases := []struct {
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
			Description:    "origin header set",
			HeaderName:     "Origin",
			HeaderValue:    "anything",
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
			HeaderValue:    "127.0.0.1",
			ExpectedStatus: http.StatusOK,
		},
		{
			Description:    "host header in allowlist (169.254.169.254)",
			HeaderName:     "Host",
			HeaderValue:    "127.0.0.1",
			ExpectedStatus: http.StatusOK,
		},
	}
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	for i, tc := range cases {
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
