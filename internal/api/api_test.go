package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/naokiiida/t42-cli/internal/config"
)

// TestParseNextLink provides unit tests for the utility function that parses
// the 'Link' HTTP header for pagination.
func TestParseNextLink(t *testing.T) {
	testCases := []struct {
		name     string
		header   string
		expected string
	}{
		{
			name:     "No Link Header",
			header:   "",
			expected: "",
		},
		{
			name:     "Only Next Link",
			header:   `<https://api.intra.42.fr/v2/users?page[size]=1&page[number]=2>; rel="next"`,
			expected: "/v2/users?page[size]=1&page[number]=2",
		},
		{
			name:     "Next and Last Links",
			header:   `<https://api.intra.42.fr/v2/users?page[size]=1&page[number]=2>; rel="next", <https://api.intra.42.fr/v2/users?page[size]=1&page[number]=10>; rel="last"`,
			expected: "/v2/users?page[size]=1&page[number]=2",
		},
		{
			name:     "Prev, Next, First, Last Links",
			header:   `<https://api.intra.42.fr/v2/users?page[size]=1&page[number]=1>; rel="prev", <https://api.intra.42.fr/v2/users?page[size]=1&page[number]=3>; rel="next", <https://api.intra.42.fr/v2/users?page[size]=1&page[number]=1>; rel="first", <https://api.intra.42.fr/v2/users?page[size]=1&page[number]=10>; rel="last"`,
			expected: "/v2/users?page[size]=1&page[number]=3",
		},
		{
			name:     "Only Prev Link",
			header:   `<https://api.intra.42.fr/v2/users?page[size]=1&page[number]=1>; rel="prev"`,
			expected: "",
		},
		{
			name:     "Malformed Header",
			header:   `this is not a valid header`,
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := parseNextLink(tc.header)
			if got != tc.expected {
				t.Errorf("parseNextLink() = %q, want %q", got, tc.expected)
			}
		})
	}
}

// TestAPIErrorFormatting checks that the custom APIError type produces the correct error message.
func TestAPIErrorFormatting(t *testing.T) {
	t.Run("Error with Message", func(t *testing.T) {
		apiErr := &APIError{
			StatusCode: 404,
			Response: ErrorResponse{
				Message: "The resource you are looking for could not be found.",
			},
		}
		expected := "API error (status 404): The resource you are looking for could not be found."
		if apiErr.Error() != expected {
			t.Errorf("Error() = %q, want %q", apiErr.Error(), expected)
		}
	})

	t.Run("Error with ErrorDescription", func(t *testing.T) {
		apiErr := &APIError{
			StatusCode: 401,
			Response: ErrorResponse{
				Error:            "unauthorized",
				ErrorDescription: "The access token is invalid or has expired.",
			},
		}
		expected := "API error (status 401): The access token is invalid or has expired."
		if apiErr.Error() != expected {
			t.Errorf("Error() = %q, want %q", apiErr.Error(), expected)
		}
	})
}

// TestGetPaginated tests the pagination logic using a mock server.
func TestGetPaginated(t *testing.T) {
	// Mock server that simulates paginated responses.
	// We must declare the server variable before creating the handler,
	// so that the handler can form a closure over it.
	var server *httptest.Server
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page[number]")
		if page == "" || page == "1" {
			// Serve page 1 and link to page 2
			w.Header().Set("Link", `<`+server.URL+`/test?page[number]=2>; rel="next"`)
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, `[{"id": 1, "name": "item1"}, {"id": 2, "name": "item2"}]`)
		} else if page == "2" {
			// Serve page 2 (last page)
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, `[{"id": 3, "name": "item3"}]`)
		} else {
			http.NotFound(w, r)
		}
	})
	server = httptest.NewServer(handler)
	defer server.Close()

	// Dummy client that uses the mock server
	client := NewClient("fake-token")
	client.baseURL = server.URL
	defer client.StopRateLimiter()

	type TestItem struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	results, err := getPaginated[TestItem](client, "/test")
	if err != nil {
		t.Fatalf("getPaginated() failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 items, got %d", len(results))
	}

	expectedNames := []string{"item1", "item2", "item3"}
	for i, item := range results {
		if item.Name != expectedNames[i] {
			t.Errorf("Item %d name = %q, want %q", i, item.Name, expectedNames[i])
		}
	}
}

// TestAPIClient_Integration performs a live integration test against the 42 API.
// It requires a valid credentials file to be present in the user's config directory.
// To run this test, first authenticate using the CLI or place a valid credentials.json manually.
// The test is skipped if credentials are not found or if `go test -short` is used.
func TestAPIClient_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode.")
	}

	// Set a temporary config directory for this test to avoid interfering with the user's real config.
	// However, for the integration test to work as described in the plan, it needs to find the *real* credentials.
	// The test plan says to manually create the file, so we will use the default path logic.
	// If we were to mock this, we'd set a temp home dir like in the config tests.
	creds, err := config.LoadCredentials()
	if err != nil {
		// This is not a failure, but a signal that the test cannot run.
		t.Skipf("Skipping integration test: could not load credentials: %v. Please run 'auth login' or create a valid credentials file.", err)
	}

	// If the token is expired, we must also skip.
	if creds.IsAccessTokenExpired() {
		t.Skipf("Skipping integration test: access token is expired. Please run 'auth login' to refresh it.")
	}

	client := NewClient(creds.AccessToken)
	defer client.StopRateLimiter()

	// As per the test plan, we will verify the client by fetching the cursus list.
	t.Run("Get Cursus List", func(t *testing.T) {
		cursusList, err := client.GetCursus()
		if err != nil {
			// Check if the error is an APIError, which could indicate an invalid token
			if apiErr, ok := err.(*APIError); ok {
				t.Fatalf("GetCursus() API error: %s. Your access token may be invalid.", apiErr)
			}
			t.Fatalf("GetCursus() failed with a non-API error: %v", err)
		}

		if len(cursusList) == 0 {
			t.Fatal("GetCursus() returned an empty list, but at least one cursus (e.g., 42cursus) was expected.")
		}

		// Log a snippet of the response for manual verification, as requested by the test plan.
		t.Logf("Successfully fetched %d cursus entries. The first one is: ID=%d, Name=%s, Slug=%s",
			len(cursusList), cursusList[0].ID, cursusList[0].Name, cursusList[0].Slug)

		// Assert that the primary "42cursus" is present.
		found := false
		for _, c := range cursusList {
			if c.Slug == "42cursus" {
				found = true
				break
			}
		}
		if !found {
			t.Error("The '42cursus' was not found in the list of fetched cursus.")
		}
	})

	// Also test the /v2/me endpoint to be thorough.
	t.Run("Get Authenticated User", func(t *testing.T) {
		user, err := client.GetMe()
		if err != nil {
			t.Fatalf("GetMe() failed: %v", err)
		}

		if user == nil || user.Login == "" {
			t.Fatal("GetMe() returned a nil user or a user with an empty login.")
		}

		t.Logf("Successfully fetched user profile for login: %s (ID: %d)", user.Login, user.ID)
	})
}

