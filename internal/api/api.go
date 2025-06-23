package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// BaseURL is the base URL for the 42 intra API.
	BaseURL = "https://api.intra.42.fr"
	// RateLimit is the time to wait between requests to respect the API's rate limit.
	// The 42 API allows 2 requests per second. We'll use a slightly more conservative
	// 600ms to be safe.
	RateLimit = 600 * time.Millisecond
)

// Client is a client for interacting with the 42 API.
type Client struct {
	httpClient  *http.Client
	baseURL     string
	accessToken string
	rateLimiter *time.Ticker
}

// APIError represents an error returned by the 42 API.
// It includes the HTTP status code and the parsed error response from the API.
type APIError struct {
	StatusCode int
	Response   ErrorResponse
}

func (e *APIError) Error() string {
	msg := e.Response.Message
	if msg == "" {
		msg = e.Response.ErrorDescription
	}
	return fmt.Sprintf("API error (status %d): %s", e.StatusCode, msg)
}

// NewClient creates and returns a new API client.
// It requires a valid OAuth2 access token.
func NewClient(accessToken string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 15 * time.Second, // Set a reasonable timeout for requests.
		},
		baseURL:     BaseURL,
		accessToken: accessToken,
		rateLimiter: time.NewTicker(RateLimit),
	}
}

// StopRateLimiter stops the client's rate limiter ticker.
// It should be called when the client is no longer needed to prevent goroutine leaks.
func (c *Client) StopRateLimiter() {
	c.rateLimiter.Stop()
}

// doRequest is the core method for making requests to the API.
// It handles authentication, rate limiting, and basic error handling.
func (c *Client) doRequest(method, path string, body io.Reader) (*http.Response, error) {
	<-c.rateLimiter.C // Wait for the rate limiter.

	fullURL := c.baseURL + path
	req, err := http.NewRequest(method, fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	// Check for non-successful status codes.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		var apiErr ErrorResponse
		// Try to decode the error response from the API.
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
			// If decoding fails, return a generic error with the status code.
			return nil, fmt.Errorf("API request failed with status %d and could not parse error response", resp.StatusCode)
		}
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Response:   apiErr,
		}
	}

	return resp, nil
}

// get is a convenience wrapper around doRequest for GET requests.
// It unmarshals the JSON response body into the provided `target` interface.
func (c *Client) get(path string, target interface{}) error {
	resp, err := c.doRequest(http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("failed to decode response body: %w", err)
	}

	return nil
}

// getPaginated fetches all pages for a given endpoint and aggregates the results.
// It is a generic function that works for any type T that can be unmarshaled from the API response.
func getPaginated[T any](c *Client, path string) ([]T, error) {
	var allItems []T
	nextURL := path

	// The 42 API has a hard limit of 100 items per page.
	// We add this parameter to every paginated request to minimize the number of requests.
	separator := "?"
	if strings.Contains(path, "?") {
		separator = "&"
	}
	nextURL += separator + "page[size]=100"

	for nextURL != "" {
		var pageItems []T

		resp, err := c.doRequest(http.MethodGet, nextURL, nil)
		if err != nil {
			return nil, err
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to read paginated response body: %w", err)
		}
		resp.Body.Close()

		if err := json.Unmarshal(bodyBytes, &pageItems); err != nil {
			return nil, fmt.Errorf("failed to decode paginated response: %w", err)
		}

		allItems = append(allItems, pageItems...)

		// Check for the 'Link' header to find the next page.
		linkHeader := resp.Header.Get("Link")
		nextURL = parseNextLink(linkHeader)
	}

	return allItems, nil
}

// parseNextLink extracts the URL for the next page from the 'Link' header.
// Example: <https://api.intra.42.fr/v2/users?page[size]=1&page[number]=2>; rel="next", ...
// It returns an empty string if no 'next' link is found.
func parseNextLink(linkHeader string) string {
	if linkHeader == "" {
		return ""
	}

	links := strings.Split(linkHeader, ",")
	for _, link := range links {
		parts := strings.Split(strings.TrimSpace(link), ";")
		if len(parts) == 2 && strings.TrimSpace(parts[1]) == `rel="next"` {
			nextURL := strings.Trim(parts[0], "<>")
			// The API returns a full URL, but we only need the path and query part for doRequest.
			u, err := url.Parse(nextURL)
			if err != nil {
				return "" // Should not happen with valid API responses.
			}
			return u.RequestURI()
		}
	}

	return ""
}

// --- Public API Methods ---

// GetMe fetches the authenticated user's profile information from `/v2/me`.
func (c *Client) GetMe() (*User, error) {
	var user User
	err := c.get("/v2/me", &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetTokenInfo fetches details about the current access token from `/oauth/token/info`.
func (c *Client) GetTokenInfo() (*TokenInfo, error) {
	var info TokenInfo
	err := c.get("/oauth/token/info", &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

// GetUserProjects fetches all projects for a given user ID.
// It handles pagination to retrieve the complete list.
func (c *Client) GetUserProjects(userID int) ([]ProjectsUser, error) {
	path := fmt.Sprintf("/v2/users/%d/projects_users", userID)
	return getPaginated[ProjectsUser](c, path)
}

// GetCursus fetches all available cursus.
func (c *Client) GetCursus() ([]Cursus, error) {
	return getPaginated[Cursus](c, "/v2/cursus")
}

// GetProjectDetails fetches detailed information about a specific project by its slug.
// Note: This endpoint is not officially documented but is commonly used.
func (c *Client) GetProjectDetails(slug string) ([]Project, error) {
	path := fmt.Sprintf("/v2/projects?filter[slug]=%s", slug)
	return getPaginated[Project](c, path)
}