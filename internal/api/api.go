package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultBaseURL is the default 42 API base URL
	DefaultBaseURL = "https://api.intra.42.fr"

	// DefaultTimeout is the default HTTP client timeout
	DefaultTimeout = 30 * time.Second

	// DefaultPerPage is the default number of items per page
	DefaultPerPage = 100

	// MaxRetries is the maximum number of retries for failed requests
	MaxRetries = 3

	// RetryDelay is the delay between retries
	RetryDelay = 1 * time.Second
)

// Client represents a 42 API client
type Client struct {
	baseURL        string
	httpClient     *http.Client
	token          string
	userAgent      string
	tokenRefresher func() (string, error) // Optional callback to refresh the token
}

// ClientOption represents a client configuration option
type ClientOption func(*Client)

// WithBaseURL sets a custom base URL for the client
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = strings.TrimSuffix(baseURL, "/")
	}
}

// WithTimeout sets a custom timeout for the HTTP client
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// WithUserAgent sets a custom user agent for requests
func WithUserAgent(userAgent string) ClientOption {
	return func(c *Client) {
		c.userAgent = userAgent
	}
}

// WithTokenRefresher sets a callback function to refresh the access token
func WithTokenRefresher(refresher func() (string, error)) ClientOption {
	return func(c *Client) {
		c.tokenRefresher = refresher
	}
}

// NewClient creates a new 42 API client with the given access token
func NewClient(token string, options ...ClientOption) *Client {
	client := &Client{
		baseURL: DefaultBaseURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		token:     token,
		userAgent: "t42-cli/1.0",
	}

	// Apply options
	for _, option := range options {
		option(client)
	}

	return client
}

// makeRequest performs an HTTP request with authentication and error handling
func (c *Client) makeRequest(ctx context.Context, method, endpoint string, body interface{}) (*http.Response, error) {
	// Try request with current token
	resp, err := c.doRequest(ctx, method, endpoint, body)
	if err != nil {
		return nil, err
	}

	// If we get 401 Unauthorized and have a token refresher, try refreshing the token
	if resp.StatusCode == 401 && c.tokenRefresher != nil {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to close response body: %v\n", err)
		}

		// Attempt to refresh the token
		newToken, refreshErr := c.tokenRefresher()
		if refreshErr != nil {
			return nil, fmt.Errorf("token refresh failed: %w", refreshErr)
		}

		// Update the client's token
		c.token = newToken

		// Retry the request with the new token
		resp, err = c.doRequest(ctx, method, endpoint, body)
		if err != nil {
			return nil, err
		}
	}

	return resp, nil
}

// doRequest performs the actual HTTP request with retries
func (c *Client) doRequest(ctx context.Context, method, endpoint string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	// Construct full URL
	fullURL := c.baseURL + endpoint

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Perform request with retries
	var resp *http.Response
	var lastErr error

	for attempt := 0; attempt <= MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retrying
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(RetryDelay * time.Duration(attempt)):
			}
		}

		resp, lastErr = c.httpClient.Do(req)
		if lastErr != nil {
			continue // Retry on network errors
		}

		// Check if we should retry based on status code
		if resp.StatusCode >= 500 || resp.StatusCode == 429 {
			if err := resp.Body.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to close response body: %v\n", err)
			}
			continue // Retry on server errors and rate limiting
		}

		// Success or client error (don't retry)
		break
	}

	if lastErr != nil {
		return nil, fmt.Errorf("request failed after %d attempts: %w", MaxRetries+1, lastErr)
	}

	return resp, nil
}

// handleResponse processes an HTTP response and unmarshals JSON data
func (c *Client) handleResponse(resp *http.Response, target interface{}) error {
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to close response body: %v\n", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for API errors
	if resp.StatusCode >= 400 {
		var apiError ErrorResponse
		if err := json.Unmarshal(body, &apiError); err != nil {
			// If we can't parse the error response, return a generic error
			return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
		}

		// Set status code if not present in the error response
		if apiError.Status == 0 {
			apiError.Status = resp.StatusCode
		}

		return fmt.Errorf("API error (status %d): %s", apiError.Status, apiError.Message)
	}

	// Parse successful response
	if target != nil {
		if err := json.Unmarshal(body, target); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// GetMe returns information about the authenticated user
func (c *Client) GetMe(ctx context.Context) (*User, error) {
	resp, err := c.makeRequest(ctx, "GET", "/v2/me", nil)
	if err != nil {
		return nil, err
	}

	var user User
	if err := c.handleResponse(resp, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUser returns information about a specific user by ID
func (c *Client) GetUser(ctx context.Context, userID int) (*User, error) {
	endpoint := fmt.Sprintf("/v2/users/%d", userID)
	resp, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var user User
	if err := c.handleResponse(resp, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByLogin returns information about a specific user by login
func (c *Client) GetUserByLogin(ctx context.Context, login string) (*User, error) {
	endpoint := fmt.Sprintf("/v2/users/%s", url.PathEscape(login))
	resp, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var user User
	if err := c.handleResponse(resp, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// ListProjectsOptions represents options for listing projects
type ListProjectsOptions struct {
	Page     int
	PerPage  int
	CursusID int
	Sort     string
}

// ListProjects returns a list of projects with optional filtering
func (c *Client) ListProjects(ctx context.Context, opts *ListProjectsOptions) ([]Project, *PaginationMeta, error) {
	if opts == nil {
		opts = &ListProjectsOptions{}
	}

	// Set defaults
	if opts.PerPage == 0 {
		opts.PerPage = DefaultPerPage
	}
	if opts.Page == 0 {
		opts.Page = 1
	}

	// Build query parameters
	params := url.Values{}
	params.Set("page", strconv.Itoa(opts.Page))
	params.Set("per_page", strconv.Itoa(opts.PerPage))

	if opts.CursusID > 0 {
		params.Set("filter[cursus_id]", strconv.Itoa(opts.CursusID))
	}
	if opts.Sort != "" {
		params.Set("sort", opts.Sort)
	}

	endpoint := "/v2/projects?" + params.Encode()
	resp, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, nil, err
	}

	var projects []Project
	if err := c.handleResponse(resp, &projects); err != nil {
		return nil, nil, err
	}

	// Extract pagination metadata from headers
	meta := c.extractPaginationMeta(resp, len(projects))

	return projects, meta, nil
}

// GetProject returns information about a specific project by ID
func (c *Client) GetProject(ctx context.Context, projectID int) (*Project, error) {
	endpoint := fmt.Sprintf("/v2/projects/%d", projectID)
	resp, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var project Project
	if err := c.handleResponse(resp, &project); err != nil {
		return nil, err
	}

	return &project, nil
}

// GetProjectBySlug returns information about a specific project by slug
func (c *Client) GetProjectBySlug(ctx context.Context, slug string) (*Project, error) {
	// Search for project by slug using the projects endpoint with filter
	params := url.Values{}
	params.Set("filter[slug]", slug)
	params.Set("per_page", "1")

	endpoint := "/v2/projects?" + params.Encode()
	resp, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var projects []Project
	if err := c.handleResponse(resp, &projects); err != nil {
		return nil, err
	}

	if len(projects) == 0 {
		return nil, fmt.Errorf("project with slug '%s' not found", slug)
	}

	return &projects[0], nil
}

// ListUserProjectsOptions represents options for listing user projects
type ListUserProjectsOptions struct {
	Page    int
	PerPage int
	Sort    string
}

// ListUserProjects returns a list of projects for a specific user
func (c *Client) ListUserProjects(ctx context.Context, userID int, opts *ListUserProjectsOptions) ([]ProjectUser, *PaginationMeta, error) {
	if opts == nil {
		opts = &ListUserProjectsOptions{}
	}

	// Set defaults
	if opts.PerPage == 0 {
		opts.PerPage = DefaultPerPage
	}
	if opts.Page == 0 {
		opts.Page = 1
	}

	// Build query parameters
	params := url.Values{}
	params.Set("page", strconv.Itoa(opts.Page))
	params.Set("per_page", strconv.Itoa(opts.PerPage))

	if opts.Sort != "" {
		params.Set("sort", opts.Sort)
	}

	endpoint := fmt.Sprintf("/v2/users/%d/projects_users?%s", userID, params.Encode())
	resp, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, nil, err
	}

	var projectUsers []ProjectUser
	if err := c.handleResponse(resp, &projectUsers); err != nil {
		return nil, nil, err
	}

	// Extract pagination metadata from headers
	meta := c.extractPaginationMeta(resp, len(projectUsers))

	return projectUsers, meta, nil
}

// ListCampuses returns a list of campuses
func (c *Client) ListCampuses(ctx context.Context) ([]Campus, error) {
	resp, err := c.makeRequest(ctx, "GET", "/v2/campus", nil)
	if err != nil {
		return nil, err
	}

	var campuses []Campus
	if err := c.handleResponse(resp, &campuses); err != nil {
		return nil, err
	}

	return campuses, nil
}

// ListCursuses returns a list of cursuses
func (c *Client) ListCursuses(ctx context.Context) ([]Cursus, error) {
	resp, err := c.makeRequest(ctx, "GET", "/v2/cursus", nil)
	if err != nil {
		return nil, err
	}

	var cursuses []Cursus
	if err := c.handleResponse(resp, &cursuses); err != nil {
		return nil, err
	}

	return cursuses, nil
}

// extractPaginationMeta extracts pagination metadata from response headers
func (c *Client) extractPaginationMeta(resp *http.Response, count int) *PaginationMeta {
	meta := &PaginationMeta{
		Count: count,
	}

	// Try to extract pagination info from headers
	if totalStr := resp.Header.Get("X-Total"); totalStr != "" {
		if total, err := strconv.Atoi(totalStr); err == nil {
			meta.TotalCount = total
		}
	}

	if pageStr := resp.Header.Get("X-Page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			meta.Page = page
		}
	}

	if perPageStr := resp.Header.Get("X-Per-Page"); perPageStr != "" {
		if perPage, err := strconv.Atoi(perPageStr); err == nil {
			meta.PerPage = perPage
		}
	}

	if totalPagesStr := resp.Header.Get("X-Total-Pages"); totalPagesStr != "" {
		if totalPages, err := strconv.Atoi(totalPagesStr); err == nil {
			meta.TotalPages = totalPages
		}
	}

	return meta
}

// IsAuthenticated checks if the client has a valid token by making a simple API call
func (c *Client) IsAuthenticated(ctx context.Context) bool {
	_, err := c.GetMe(ctx)
	return err == nil
}

// GetProjectUser returns information about a specific project user by ID
func (c *Client) GetProjectUser(ctx context.Context, projectUserID int) (*ProjectUser, error) {
	endpoint := fmt.Sprintf("/v2/projects_users/%d", projectUserID)
	resp, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var projectUser ProjectUser
	if err := c.handleResponse(resp, &projectUser); err != nil {
		return nil, err
	}

	return &projectUser, nil
}

// GetToken returns the current access token
func (c *Client) GetToken() string {
	return c.token
}

// ListUsersOptions represents options for listing users
type ListUsersOptions struct {
	Page    int
	PerPage int
	Sort    string
	// Filter options
	FilterLogin    string
	FilterCampusID int
	FilterCursusID int
	FilterActive   *bool
	FilterStaff    *bool
	FilterAlumni   *bool
}

// ListUsers returns a list of users with optional filtering
func (c *Client) ListUsers(ctx context.Context, opts *ListUsersOptions) ([]User, *PaginationMeta, error) {
	if opts == nil {
		opts = &ListUsersOptions{}
	}

	// Set defaults
	if opts.PerPage == 0 {
		opts.PerPage = DefaultPerPage
	}
	if opts.Page == 0 {
		opts.Page = 1
	}

	// Build query parameters
	params := url.Values{}
	params.Set("page", strconv.Itoa(opts.Page))
	params.Set("per_page", strconv.Itoa(opts.PerPage))

	if opts.FilterCampusID > 0 {
		params.Set("filter[campus_id]", strconv.Itoa(opts.FilterCampusID))
	}
	if opts.FilterCursusID > 0 {
		params.Set("filter[cursus_id]", strconv.Itoa(opts.FilterCursusID))
	}
	if opts.FilterLogin != "" {
		params.Set("filter[login]", opts.FilterLogin)
	}
	if opts.FilterActive != nil {
		params.Set("filter[active]", strconv.FormatBool(*opts.FilterActive))
	}
	if opts.FilterStaff != nil {
		params.Set("filter[staff?]", strconv.FormatBool(*opts.FilterStaff))
	}
	if opts.FilterAlumni != nil {
		params.Set("filter[alumni?]", strconv.FormatBool(*opts.FilterAlumni))
	}
	if opts.Sort != "" {
		params.Set("sort", opts.Sort)
	}

	endpoint := "/v2/users?" + params.Encode()
	resp, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, nil, err
	}

	var users []User
	if err := c.handleResponse(resp, &users); err != nil {
		return nil, nil, err
	}

	// Extract pagination metadata from headers
	meta := c.extractPaginationMeta(resp, len(users))

	return users, meta, nil
}

// ListCursusUsersOptions represents options for listing cursus users
// Note: /v2/cursus_users does not support alumni filter (only /v2/users does)
type ListCursusUsersOptions struct {
	Page         int
	PerPage      int
	CampusID     int
	Sort         string
	FilterActive *bool
}

// ListCursusUsers returns a list of cursus users with full data (level, blackhole, etc.)
// This endpoint provides more detailed data than ListCampusUsers
func (c *Client) ListCursusUsers(ctx context.Context, cursusID int, opts *ListCursusUsersOptions) ([]CursusUser, *PaginationMeta, error) {
	if opts == nil {
		opts = &ListCursusUsersOptions{}
	}

	// Set defaults
	if opts.PerPage == 0 {
		opts.PerPage = DefaultPerPage
	}
	if opts.Page == 0 {
		opts.Page = 1
	}

	// Build query parameters
	params := url.Values{}
	params.Set("page", strconv.Itoa(opts.Page))
	params.Set("per_page", strconv.Itoa(opts.PerPage))
	params.Set("filter[cursus_id]", strconv.Itoa(cursusID))

	if opts.CampusID > 0 {
		params.Set("filter[campus_id]", strconv.Itoa(opts.CampusID))
	}
	if opts.FilterActive != nil {
		params.Set("filter[active]", strconv.FormatBool(*opts.FilterActive))
	}
	if opts.Sort != "" {
		params.Set("sort", opts.Sort)
	}

	endpoint := "/v2/cursus_users?" + params.Encode()
	resp, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, nil, err
	}

	var cursusUsers []CursusUser
	if err := c.handleResponse(resp, &cursusUsers); err != nil {
		return nil, nil, err
	}

	// Extract pagination metadata from headers
	meta := c.extractPaginationMeta(resp, len(cursusUsers))

	return cursusUsers, meta, nil
}

// ListCampusUsers returns a list of users from a specific campus
func (c *Client) ListCampusUsers(ctx context.Context, campusID int, opts *ListUsersOptions) ([]User, *PaginationMeta, error) {
	if opts == nil {
		opts = &ListUsersOptions{}
	}

	// Set defaults
	if opts.PerPage == 0 {
		opts.PerPage = DefaultPerPage
	}
	if opts.Page == 0 {
		opts.Page = 1
	}

	// Build query parameters
	params := url.Values{}
	params.Set("page", strconv.Itoa(opts.Page))
	params.Set("per_page", strconv.Itoa(opts.PerPage))

	if opts.FilterCursusID > 0 {
		params.Set("filter[cursus_id]", strconv.Itoa(opts.FilterCursusID))
	}
	if opts.FilterLogin != "" {
		params.Set("filter[login]", opts.FilterLogin)
	}
	if opts.FilterActive != nil {
		params.Set("filter[active]", strconv.FormatBool(*opts.FilterActive))
	}
	if opts.FilterStaff != nil {
		params.Set("filter[staff?]", strconv.FormatBool(*opts.FilterStaff))
	}
	if opts.FilterAlumni != nil {
		params.Set("filter[alumni?]", strconv.FormatBool(*opts.FilterAlumni))
	}
	if opts.Sort != "" {
		params.Set("sort", opts.Sort)
	}

	endpoint := fmt.Sprintf("/v2/campus/%d/users?%s", campusID, params.Encode())
	resp, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, nil, err
	}

	var users []User
	if err := c.handleResponse(resp, &users); err != nil {
		return nil, nil, err
	}

	// Extract pagination metadata from headers
	meta := c.extractPaginationMeta(resp, len(users))

	return users, meta, nil
}
