package client

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUpload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/upload" {
			t.Errorf("expected /api/upload, got %s", r.URL.Path)
		}
		if ct := r.Header.Get("Content-Type"); ct != "text/html; charset=utf-8" {
			t.Errorf("expected text/html content type, got %s", ct)
		}
		if ttl := r.Header.Get("X-TTL"); ttl != "3600" {
			t.Errorf("expected X-TTL 3600, got %s", ttl)
		}

		body, _ := io.ReadAll(r.Body)
		if string(body) != "<h1>test</h1>" {
			t.Errorf("unexpected body: %s", string(body))
		}

		w.Header().Set("X-RateLimit-Limit", "10")
		w.Header().Set("X-RateLimit-Remaining", "9")
		w.Header().Set("X-RateLimit-Reset", "1708300800")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"abc123","url":"https://sh.techops.services/abc123","expires_at":"2024-02-20T00:00:00Z","size":13,"ttl":3600}`))
	}))
	defer server.Close()

	c := New(
		WithEndpoint(server.URL),
		WithHTTPClient(server.Client()),
	)

	resp, err := c.Upload(strings.NewReader("<h1>test</h1>"), 3600)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "abc123" {
		t.Errorf("expected ID abc123, got %s", resp.ID)
	}
	if resp.URL != "https://sh.techops.services/abc123" {
		t.Errorf("expected URL https://sh.techops.services/abc123, got %s", resp.URL)
	}
	if resp.Size != 13 {
		t.Errorf("expected Size 13, got %d", resp.Size)
	}

	if c.LastRateLimit == nil {
		t.Fatal("expected rate limit info to be set")
	}
	if c.LastRateLimit.Limit != 10 {
		t.Errorf("expected rate limit 10, got %d", c.LastRateLimit.Limit)
	}
	if c.LastRateLimit.Remaining != 9 {
		t.Errorf("expected rate remaining 9, got %d", c.LastRateLimit.Remaining)
	}
}

func TestUploadWithAPIKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if key := r.Header.Get("X-Api-Key"); key != "sk_test_123" {
			t.Errorf("expected X-Api-Key sk_test_123, got %s", key)
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"xyz","url":"https://sh.techops.services/xyz","expires_at":"2024-02-20T00:00:00Z","size":5,"ttl":86400}`))
	}))
	defer server.Close()

	c := New(
		WithEndpoint(server.URL),
		WithAPIKey("sk_test_123"),
		WithHTTPClient(server.Client()),
	)

	resp, err := c.Upload(strings.NewReader("hello"), 86400)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "xyz" {
		t.Errorf("expected ID xyz, got %s", resp.ID)
	}
}

func TestUploadErrorParsing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":{"code":"rate_limited","message":"Too many requests","retry_after":300}}`))
	}))
	defer server.Close()

	c := New(
		WithEndpoint(server.URL),
		WithHTTPClient(server.Client()),
	)

	_, err := c.Upload(strings.NewReader("<h1>test</h1>"), 3600)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 429 {
		t.Errorf("expected status 429, got %d", apiErr.StatusCode)
	}
	if !apiErr.IsRateLimit() {
		t.Error("expected IsRateLimit() to be true")
	}
	if apiErr.RetryAfter != 300 {
		t.Errorf("expected RetryAfter 300, got %d", apiErr.RetryAfter)
	}
}

func TestUploadPayloadTooLarge(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		w.Write([]byte(`{"error":{"code":"payload_too_large","message":"Content exceeds 5MB limit"}}`))
	}))
	defer server.Close()

	c := New(
		WithEndpoint(server.URL),
		WithHTTPClient(server.Client()),
	)

	_, err := c.Upload(strings.NewReader("<h1>test</h1>"), 3600)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if !apiErr.IsPayloadTooLarge() {
		t.Error("expected IsPayloadTooLarge() to be true")
	}
}

func TestList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/list" {
			t.Errorf("expected /api/list, got %s", r.URL.Path)
		}
		if key := r.Header.Get("X-Api-Key"); key != "sk_test" {
			t.Errorf("expected api key sk_test, got %s", key)
		}
		if limit := r.URL.Query().Get("limit"); limit != "10" {
			t.Errorf("expected limit 10, got %s", limit)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"shares":[{"id":"abc","url":"https://sh.techops.services/abc","created_at":"2024-02-19T12:00:00Z","expires_at":"2024-02-20T12:00:00Z","size":100,"views":5}],"total":1}`))
	}))
	defer server.Close()

	c := New(
		WithEndpoint(server.URL),
		WithAPIKey("sk_test"),
		WithHTTPClient(server.Client()),
	)

	resp, err := c.List(10, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Total != 1 {
		t.Errorf("expected total 1, got %d", resp.Total)
	}
	if len(resp.Shares) != 1 {
		t.Fatalf("expected 1 share, got %d", len(resp.Shares))
	}
	if resp.Shares[0].ID != "abc" {
		t.Errorf("expected share ID abc, got %s", resp.Shares[0].ID)
	}
}

func TestDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/abc123" {
			t.Errorf("expected /api/abc123, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := New(
		WithEndpoint(server.URL),
		WithAPIKey("sk_test"),
		WithHTTPClient(server.Client()),
	)

	err := c.Delete("abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHealth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/health" {
			t.Errorf("expected /api/health, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","version":"1.0.0"}`))
	}))
	defer server.Close()

	c := New(
		WithEndpoint(server.URL),
		WithHTTPClient(server.Client()),
	)

	resp, err := c.Health()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "ok" {
		t.Errorf("expected status ok, got %s", resp.Status)
	}
	if resp.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", resp.Version)
	}
}
