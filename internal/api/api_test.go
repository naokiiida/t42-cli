package api

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/naokiiida/t42-cli/internal/config"
)

func TestNewClient(t *testing.T) {
	token := "test_token"
	client := NewClient(token)

	if client.token != token {
		t.Errorf("Expected token %s, got %s", token, client.token)
	}

	if client.baseURL != DefaultBaseURL {
		t.Errorf("Expected base URL %s, got %s", DefaultBaseURL, client.baseURL)
	}

	if client.httpClient.Timeout != DefaultTimeout {
		t.Errorf("Expected timeout %v, got %v", DefaultTimeout, client.httpClient.Timeout)
	}

	if client.userAgent != "t42-cli/1.0" {
		t.Errorf("Expected user agent 't42-cli/1.0', got %s", client.userAgent)
	}
}

func TestNewClientWithOptions(t *testing.T) {
	token := "test_token"
	customBaseURL := "https://api.test.42.fr"
	customTimeout := 60 * time.Second
	customUserAgent := "test-client/2.0"

	client := NewClient(token,
		WithBaseURL(customBaseURL),
		WithTimeout(customTimeout),
		WithUserAgent(customUserAgent),
	)

	if client.baseURL != customBaseURL {
		t.Errorf("Expected base URL %s, got %s", customBaseURL, client.baseURL)
	}

	if client.httpClient.Timeout != customTimeout {
		t.Errorf("Expected timeout %v, got %v", customTimeout, client.httpClient.Timeout)
	}

	if client.userAgent != customUserAgent {
		t.Errorf("Expected user agent %s, got %s", customUserAgent, client.userAgent)
	}
}

func TestWithBaseURLTrimsSlash(t *testing.T) {
	token := "test_token"
	baseURLWithSlash := "https://api.test.42.fr/"
	expectedBaseURL := "https://api.test.42.fr"

	client := NewClient(token, WithBaseURL(baseURLWithSlash))

	if client.baseURL != expectedBaseURL {
		t.Errorf("Expected base URL %s, got %s", expectedBaseURL, client.baseURL)
	}
}

func TestGetToken(t *testing.T) {
	token := "test_token_123"
	client := NewClient(token)

	if client.GetToken() != token {
		t.Errorf("Expected token %s, got %s", token, client.GetToken())
	}
}

// Integration tests - these require a valid 42 API token
func TestIntegration(t *testing.T) {
	// Skip integration tests if not in development environment
	if os.Getenv("T42_ENV") != "development" {
		t.Skip("Skipping integration tests - set T42_ENV=development to run")
	}

	// Load credentials from development environment
	credentials, err := config.LoadCredentials()
	if err != nil {
		t.Skipf("Skipping integration tests - no valid credentials found: %v", err)
	}

	if credentials.AccessToken == "" {
		t.Skip("Skipping integration tests - empty access token")
	}

	client := NewClient(credentials.AccessToken)
	ctx := context.Background()

	t.Run("GetMe", func(t *testing.T) {
		user, err := client.GetMe(ctx)
		if err != nil {
			t.Fatalf("GetMe() error = %v", err)
		}

		if user.ID == 0 {
			t.Error("Expected user ID to be non-zero")
		}

		if user.Login == "" {
			t.Error("Expected user login to be non-empty")
		}

		if user.Email == "" {
			t.Error("Expected user email to be non-empty")
		}

		t.Logf("Successfully retrieved user: %s (ID: %d)", user.Login, user.ID)
	})

	t.Run("IsAuthenticated", func(t *testing.T) {
		if !client.IsAuthenticated(ctx) {
			t.Error("Expected client to be authenticated")
		}
	})

	t.Run("ListCursuses", func(t *testing.T) {
		cursuses, err := client.ListCursuses(ctx)
		if err != nil {
			t.Fatalf("ListCursuses() error = %v", err)
		}

		if len(cursuses) == 0 {
			t.Error("Expected at least one cursus")
		}

		// Log first few cursuses for verification
		for i, cursus := range cursuses {
			if i >= 3 {
				break
			}
			t.Logf("Cursus %d: %s (ID: %d)", i+1, cursus.Name, cursus.ID)
		}
	})

	t.Run("ListCampuses", func(t *testing.T) {
		campuses, err := client.ListCampuses(ctx)
		if err != nil {
			t.Fatalf("ListCampuses() error = %v", err)
		}

		if len(campuses) == 0 {
			t.Error("Expected at least one campus")
		}

		// Log first few campuses for verification
		for i, campus := range campuses {
			if i >= 3 {
				break
			}
			t.Logf("Campus %d: %s (ID: %d)", i+1, campus.Name, campus.ID)
		}
	})

	t.Run("ListProjects", func(t *testing.T) {
		opts := &ListProjectsOptions{
			Page:    1,
			PerPage: 5, // Limit to reduce test time
		}

		projects, meta, err := client.ListProjects(ctx, opts)
		if err != nil {
			t.Fatalf("ListProjects() error = %v", err)
		}

		if len(projects) == 0 {
			t.Error("Expected at least one project")
		}

		if meta == nil {
			t.Error("Expected pagination metadata")
		} else {
			t.Logf("Retrieved %d projects (page %d, total: %d)", len(projects), meta.Page, meta.TotalCount)
		}

		// Log first few projects for verification
		for i, project := range projects {
			if i >= 3 {
				break
			}
			t.Logf("Project %d: %s (ID: %d, Slug: %s)", i+1, project.Name, project.ID, project.Slug)
		}
	})

	t.Run("GetProjectBySlug", func(t *testing.T) {
		// Try to get a common project that should exist
		commonSlugs := []string{"libft", "get_next_line", "ft_printf", "push_swap"}
		
		var project *Project
		var err error
		
		for _, slug := range commonSlugs {
			project, err = client.GetProjectBySlug(ctx, slug)
			if err == nil {
				break
			}
		}

		if err != nil {
			t.Skipf("Could not find any common project - this may be expected: %v", err)
		}

		if project.Name == "" {
			t.Error("Expected project name to be non-empty")
		}

		if project.Slug == "" {
			t.Error("Expected project slug to be non-empty")
		}

		t.Logf("Successfully retrieved project: %s (ID: %d, Slug: %s)", project.Name, project.ID, project.Slug)
	})

	t.Run("GetUserByLogin", func(t *testing.T) {
		// First get the current user to use their login
		me, err := client.GetMe(ctx)
		if err != nil {
			t.Fatalf("GetMe() error = %v", err)
		}

		user, err := client.GetUserByLogin(ctx, me.Login)
		if err != nil {
			t.Fatalf("GetUserByLogin() error = %v", err)
		}

		if user.ID != me.ID {
			t.Errorf("Expected user ID %d, got %d", me.ID, user.ID)
		}

		if user.Login != me.Login {
			t.Errorf("Expected user login %s, got %s", me.Login, user.Login)
		}

		t.Logf("Successfully retrieved user by login: %s (ID: %d)", user.Login, user.ID)
	})

	t.Run("ListUserProjects", func(t *testing.T) {
		// Get current user first
		me, err := client.GetMe(ctx)
		if err != nil {
			t.Fatalf("GetMe() error = %v", err)
		}

		opts := &ListUserProjectsOptions{
			Page:    1,
			PerPage: 5, // Limit to reduce test time
		}

		projectUsers, meta, err := client.ListUserProjects(ctx, me.ID, opts)
		if err != nil {
			t.Fatalf("ListUserProjects() error = %v", err)
		}

		if meta == nil {
			t.Error("Expected pagination metadata")
		} else {
			t.Logf("Retrieved %d user projects (page %d)", len(projectUsers), meta.Page)
		}

		// Log first few projects for verification (user might not have any projects)
		for i, projectUser := range projectUsers {
			if i >= 3 {
				break
			}
			status := "unknown"
			if projectUser.Status != "" {
				status = projectUser.Status
			}
			t.Logf("User Project %d: %s (Status: %s)", i+1, projectUser.Project.Name, status)
		}
	})
}

func TestIntegrationWithInvalidToken(t *testing.T) {
	client := NewClient("invalid_token_123")
	ctx := context.Background()

	t.Run("IsAuthenticated with invalid token", func(t *testing.T) {
		if client.IsAuthenticated(ctx) {
			t.Error("Expected client with invalid token to not be authenticated")
		}
	})

	t.Run("GetMe with invalid token", func(t *testing.T) {
		_, err := client.GetMe(ctx)
		if err == nil {
			t.Error("Expected GetMe() to fail with invalid token")
		}
		t.Logf("Expected error with invalid token: %v", err)
	})
}

func TestListProjectsOptions(t *testing.T) {
	// Skip integration tests if not in development environment
	if os.Getenv("T42_ENV") != "development" {
		t.Skip("Skipping integration tests - set T42_ENV=development to run")
	}

	credentials, err := config.LoadCredentials()
	if err != nil {
		t.Skipf("Skipping integration tests - no valid credentials found: %v", err)
	}

	client := NewClient(credentials.AccessToken)
	ctx := context.Background()

	t.Run("ListProjects with nil options", func(t *testing.T) {
		projects, meta, err := client.ListProjects(ctx, nil)
		if err != nil {
			t.Fatalf("ListProjects() error = %v", err)
		}

		if meta == nil {
			t.Error("Expected pagination metadata")
		} else {
			// Should use defaults
			if meta.Page != 1 {
				t.Errorf("Expected default page 1, got %d", meta.Page)
			}
			if meta.PerPage != DefaultPerPage {
				t.Errorf("Expected default per_page %d, got %d", DefaultPerPage, meta.PerPage)
			}
		}

		t.Logf("Retrieved %d projects with default options", len(projects))
	})

	t.Run("ListProjects with custom page size", func(t *testing.T) {
		opts := &ListProjectsOptions{
			Page:    1,
			PerPage: 2, // Very small page size
		}

		projects, meta, err := client.ListProjects(ctx, opts)
		if err != nil {
			t.Fatalf("ListProjects() error = %v", err)
		}

		if len(projects) > 2 {
			t.Errorf("Expected at most 2 projects, got %d", len(projects))
		}

		if meta == nil {
			t.Error("Expected pagination metadata")
		} else {
			if meta.Page != 1 {
				t.Errorf("Expected page 1, got %d", meta.Page)
			}
			if meta.PerPage != 2 {
				t.Errorf("Expected per_page 2, got %d", meta.PerPage)
			}
		}

		t.Logf("Retrieved %d projects with custom page size", len(projects))
	})
}

func TestErrorHandling(t *testing.T) {
	// Test with invalid base URL
	client := NewClient("test_token", WithBaseURL("invalid-url"))
	ctx := context.Background()

	t.Run("Invalid base URL", func(t *testing.T) {
		_, err := client.GetMe(ctx)
		if err == nil {
			t.Error("Expected error with invalid base URL")
		}
		t.Logf("Expected error with invalid URL: %v", err)
	})
}

func TestContextCancellation(t *testing.T) {
	client := NewClient("test_token")
	
	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	t.Run("Cancelled context", func(t *testing.T) {
		_, err := client.GetMe(ctx)
		if err == nil {
			t.Error("Expected error with cancelled context")
		}
		
		if err != context.Canceled {
			t.Logf("Error with cancelled context: %v", err)
		}
	})
}