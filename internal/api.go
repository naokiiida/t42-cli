package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Minimal 42 API client for t42 CLI
// Implements:
// - HTTP requests with error handling (see 42 API error codes)
// - Config/token loading and storage
// - Pagination, retries, and rate limiting (2 req/sec per 42 API docs)
// - Authorization: Bearer <token> header
// - Low-level passthrough for unsupported endpoints
//
// References:
// - https://api.intra.42.fr/apidoc/guides/getting_started
// - https://api.intra.42.fr/apidoc/guides/specification
// - .rules/42api.llms.md
//
// Extend as needed for richer error types, response parsing, etc.

// Config holds API credentials and config
// (expand as needed for more config)
type Config struct {
	AccessToken  string `json:"access_token"`
	ClientID     string `json:"client_id,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`
}

func (c *Config) GetClientID() string       { return c.ClientID }
func (c *Config) GetClientSecret() string   { return c.ClientSecret }

// configFilePath returns the path to the config file
func configFilePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "t42", "credentials.json"), nil
}

// ConfigFilePath returns the path to the config file (exported)
func ConfigFilePath() (string, error) {
	return configFilePath()
}

// LoadConfig loads the config from disk
func LoadConfig() (*Config, error) {
	path, err := configFilePath()
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// SaveConfig saves the config to disk
func SaveConfig(cfg *Config) error {
	path, err := configFilePath()
	if err != nil {
		return err
	}
	os.MkdirAll(filepath.Dir(path), 0700)
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(cfg)
}

// APIClient is a minimal HTTP client for the 42 API
type APIClient struct {
	BaseURL     string
	AccessToken string
	HTTPClient  *http.Client
}

// NewAPIClient creates a new API client using config
func NewAPIClient(cfg *Config) *APIClient {
	return &APIClient{
		BaseURL:     "https://api.intra.42.fr",
		AccessToken: cfg.AccessToken,
		HTTPClient:  http.DefaultClient,
	}
}

// DoRequest performs an HTTP request with error handling and token
func (c *APIClient) DoRequest(method, path string, body any) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(b)
	}
	url := c.BaseURL + path
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, err
	}
	if c.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, errors.New(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(b)))
	}
	return resp, nil
}

// DoRequestWithRetry supports retries and rate limiting
func (c *APIClient) DoRequestWithRetry(method, path string, body any, maxRetries int) (*http.Response, error) {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		resp, err := c.DoRequest(method, path, body)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		// Simple rate limit: sleep and retry
		time.Sleep(600 * time.Millisecond) // 2 req/sec per 42 API docs
	}
	return nil, lastErr
}

// PaginatedGet fetches all pages for a GET endpoint (returns all items as []byte for now)
func (c *APIClient) PaginatedGet(path string, perPage int) ([][]byte, error) {
	var all [][]byte
	page := 1
	for {
		p := fmt.Sprintf("%s?page[number]=%d&page[size]=%d", path, page, perPage)
		resp, err := c.DoRequestWithRetry("GET", p, nil, 3)
		if err != nil {
			return nil, err
		}
		b, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		if len(b) == 0 || string(b) == "[]" {
			break
		}
		all = append(all, b)
		// Check for Link header for next page (not implemented, just increment for now)
		page++
	}
	return all, nil
}

// Passthrough allows low-level API calls (t42 api ...)
func (c *APIClient) Passthrough(method, path string, body any) (*http.Response, error) {
	return c.DoRequestWithRetry(method, path, body, 3)
} 