package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	defaultEndpoint = "https://share.techops.services"
	defaultTimeout  = 30 * time.Second
	maxBodySize     = 5 * 1024 * 1024 // 5MB
)

// UploadResponse is the JSON body returned by POST /api/upload.
type UploadResponse struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	ExpiresAt string `json:"expires_at"`
	Size      int    `json:"size"`
	TTL       int    `json:"ttl"`
}

// Share represents a single shared page in a list response.
type Share struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	CreatedAt string `json:"created_at"`
	ExpiresAt string `json:"expires_at"`
	Size      int    `json:"size"`
	Views     int    `json:"views"`
}

// ListResponse is the JSON body returned by GET /api/list.
type ListResponse struct {
	Shares []Share `json:"shares"`
	Cursor string  `json:"cursor,omitempty"`
	Total  int     `json:"total"`
}

// HealthResponse is the JSON body returned by GET /api/health.
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// Client is an HTTP client for the share Worker API.
type Client struct {
	endpoint   string
	apiKey     string
	httpClient *http.Client

	// LastRateLimit is populated after each request.
	LastRateLimit *RateLimitInfo
}

// Option configures a Client.
type Option func(*Client)

// WithEndpoint sets a custom API endpoint.
func WithEndpoint(endpoint string) Option {
	return func(c *Client) {
		c.endpoint = strings.TrimRight(endpoint, "/")
	}
}

// WithAPIKey sets the API key for authenticated requests.
func WithAPIKey(key string) Option {
	return func(c *Client) {
		c.apiKey = key
	}
}

// WithHTTPClient sets a custom http.Client (useful for testing).
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		c.httpClient = hc
	}
}

// New creates a new API client.
func New(opts ...Option) *Client {
	c := &Client{
		endpoint: defaultEndpoint,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Upload sends HTML content to the upload endpoint and returns the response.
func (c *Client) Upload(body io.Reader, ttlSeconds int) (*UploadResponse, error) {
	req, err := http.NewRequest(http.MethodPost, c.endpoint+"/api/upload", body)
	if err != nil {
		return nil, fmt.Errorf("creating upload request: %w", err)
	}

	req.Header.Set("Content-Type", "text/html; charset=utf-8")
	if ttlSeconds > 0 {
		req.Header.Set("X-TTL", strconv.Itoa(ttlSeconds))
	}
	if c.apiKey != "" {
		req.Header.Set("X-Api-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("uploading: %w", err)
	}
	defer resp.Body.Close()

	c.parseRateLimit(resp)

	if resp.StatusCode != http.StatusCreated {
		return nil, c.parseError(resp)
	}

	var result UploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding upload response: %w", err)
	}
	return &result, nil
}

// List retrieves shared pages. Requires an API key.
func (c *Client) List(limit int, cursor string) (*ListResponse, error) {
	url := fmt.Sprintf("%s/api/list?limit=%d", c.endpoint, limit)
	if cursor != "" {
		url += "&cursor=" + cursor
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating list request: %w", err)
	}

	if c.apiKey != "" {
		req.Header.Set("X-Api-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("listing shares: %w", err)
	}
	defer resp.Body.Close()

	c.parseRateLimit(resp)

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result ListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding list response: %w", err)
	}
	return &result, nil
}

// Delete removes a shared page by ID. Requires an API key.
func (c *Client) Delete(id string) error {
	req, err := http.NewRequest(http.MethodDelete, c.endpoint+"/api/"+id, nil)
	if err != nil {
		return fmt.Errorf("creating delete request: %w", err)
	}

	if c.apiKey != "" {
		req.Header.Set("X-Api-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("deleting share: %w", err)
	}
	defer resp.Body.Close()

	c.parseRateLimit(resp)

	if resp.StatusCode != http.StatusNoContent {
		return c.parseError(resp)
	}
	return nil
}

// Health checks the API health endpoint.
func (c *Client) Health() (*HealthResponse, error) {
	req, err := http.NewRequest(http.MethodGet, c.endpoint+"/api/health", nil)
	if err != nil {
		return nil, fmt.Errorf("creating health request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("checking health: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("health check failed with status %d", resp.StatusCode)
	}

	var result HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding health response: %w", err)
	}
	return &result, nil
}

// parseRateLimit extracts rate limit info from response headers.
func (c *Client) parseRateLimit(resp *http.Response) {
	info := &RateLimitInfo{}
	if v := resp.Header.Get("X-RateLimit-Limit"); v != "" {
		info.Limit, _ = strconv.Atoi(v)
	}
	if v := resp.Header.Get("X-RateLimit-Remaining"); v != "" {
		info.Remaining, _ = strconv.Atoi(v)
	}
	if v := resp.Header.Get("X-RateLimit-Reset"); v != "" {
		info.Reset, _ = strconv.ParseInt(v, 10, 64)
	}
	c.LastRateLimit = info
}

// parseError reads an error response body and returns a structured APIError.
func (c *Client) parseError(resp *http.Response) error {
	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 1024*64))
	if err != nil {
		return &APIError{
			StatusCode: resp.StatusCode,
			Code:       "read_error",
			Message:    fmt.Sprintf("HTTP %d (could not read response body)", resp.StatusCode),
		}
	}

	// Try to parse structured error: { "error": { ... } }
	var envelope struct {
		Error APIError `json:"error"`
	}
	if jsonErr := json.Unmarshal(bodyBytes, &envelope); jsonErr == nil && envelope.Error.Message != "" {
		envelope.Error.StatusCode = resp.StatusCode
		return &envelope.Error
	}

	// Fallback to plain text
	return &APIError{
		StatusCode: resp.StatusCode,
		Code:       "unknown",
		Message:    fmt.Sprintf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(bodyBytes))),
	}
}
