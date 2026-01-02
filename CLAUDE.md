# t42-cli

<!-- AUTO-MANAGED: project-description -->
## Overview

**t42-cli** is a command-line interface for the 42 School API with OAuth2 authentication and PKCE support.

Key features:
- OAuth2 authentication with PKCE (Proof Key for Code Exchange)
- Automatic token refresh with retry logic
- XDG Base Directory specification compliance
- JSON and table output formats
- User, project, and campus management commands

<!-- END AUTO-MANAGED -->

<!-- AUTO-MANAGED: build-commands -->
## Build & Development Commands

```bash
# Build
make build              # Build t42 binary
go build -o t42 .       # Direct Go build

# Test
make test               # Run all tests
go test ./...           # Direct Go test
go test -v ./...        # Verbose test output

# Lint
make lint               # Run golangci-lint
golangci-lint run       # Direct lint
golangci-lint run --fix # Auto-fix lint issues

# Dependencies
make tidy               # go mod tidy
go mod tidy             # Clean up dependencies

# Clean
make clean              # Remove built binary

# Run - Authentication
./t42 auth login        # OAuth2 login flow
./t42 auth status       # Check authentication

# Run - User commands
./t42 user list                                  # List all users
./t42 user list --campus tokyo --cursus-id 21    # Filter by campus and cursus
./t42 user list --min-level 5 --cursus-id 21     # Filter by level (requires cursus-id)
./t42 user list --blackhole-status upcoming      # Filter by blackhole status
./t42 user show <login>                          # Show user details

# Run - Project commands
./t42 project list      # List projects
```

<!-- END AUTO-MANAGED -->

<!-- AUTO-MANAGED: architecture -->
## Architecture

```text
t42-cli/
├── main.go                 # Entry point - calls cmd.Execute()
├── cmd/                    # CLI commands (Cobra)
│   ├── root.go             # Root command, global flags, API client factory
│   ├── auth.go             # OAuth2 login/logout/status commands
│   ├── user.go             # User list/show commands with filters (campus, cursus, level, blackhole)
│   ├── user_test.go        # Tests for blackhole filtering, cursus matching, flag validation
│   ├── project.go          # Project list/show/clone commands
│   └── util.go             # Shared utility functions (truncateString, etc.)
├── internal/
│   ├── api/                # 42 API client
│   │   ├── api.go          # HTTP client with retry, pagination, token refresh, error handling
│   │   ├── types.go        # API response types (User, Project, Campus, CursusUser, etc.)
│   │   └── api_test.go     # API client tests
│   ├── config/             # Configuration management
│   │   ├── config.go       # Credentials and config loading/saving
│   │   ├── paths.go        # XDG-compliant path resolution
│   │   └── config_test.go  # Config tests
│   └── oauth/              # OAuth2 PKCE implementation
│       ├── pkce.go         # RFC 7636 PKCE code verifier/challenge
│       └── pkce_test.go    # PKCE tests
├── docs/                   # Documentation
│   ├── api_endpoints_guide.md       # API endpoint data completeness analysis
│   ├── architecture.md              # Detailed architecture documentation
│   ├── features.md                  # Complete CLI command reference
│   └── claude_code_development.md   # Development workflow guide
├── Makefile                # Build automation
├── go.mod                  # Go module definition
└── lefthook.yml            # Git hooks configuration
```

**Data Flow**:
1. CLI commands in `cmd/` parse flags and call API methods
2. Flag validation: check for mutually exclusive flags and endpoint compatibility
3. `cmd/root.go:NewAPIClient()` creates authenticated API client with token refresh callback
4. `internal/api/` handles HTTP requests with automatic token refresh on 401 responses
5. Smart endpoint selection: switches between `/v2/campus/{id}/users` and `/v2/cursus_users` based on filter requirements
6. `internal/config/` manages credentials and XDG-compliant file paths
7. `internal/oauth/` generates PKCE parameters for secure OAuth2 flow

<!-- END AUTO-MANAGED -->

<!-- AUTO-MANAGED: conventions -->
## Code Conventions

**Go Style**:
- Standard Go formatting (`gofmt`)
- Package names: lowercase, single word (`api`, `config`, `oauth`)
- File names: lowercase with underscores (`api_test.go`, `util.go`)
- Struct names: PascalCase (`Client`, `Credentials`, `PKCEParams`)
- Private functions: camelCase (`makeRequest`, `handleResponse`, `truncateString`)
- Shared utilities: Extract reusable helpers to `util.go` in relevant package

**Error Handling**:
- Wrap errors with context: `fmt.Errorf("failed to X: %w", err)`
- Return early on errors (guard clauses)
- Check errors immediately after function calls
- Always check and log errors from deferred `Close()` calls: `if err := resp.Body.Close(); err != nil { fmt.Fprintf(os.Stderr, ...) }`
- Never ignore errors from `json.Marshal()` or `json.MarshalIndent()` - always check and wrap

**CLI Patterns**:
- Use Cobra for command structure
- Global flags in `rootCmd.PersistentFlags()`
- Use `--json` flag for JSON output
- Use `-v/--verbose` for debug output
- Campus name resolution: convert user-friendly names (e.g., "tokyo") to campus IDs via API lookup
- Validate mutually exclusive flags early (return error before API calls)
- Validate endpoint compatibility: check if flags are supported by selected endpoint (e.g., --alumni incompatible with /v2/cursus_users)
- Helpful error messages: show available options when user input is invalid, explain endpoint limitations
- JSON output includes `filter_info` to explain client-side vs server-side filtering

**API Client Patterns**:
- Functional options pattern: `WithBaseURL()`, `WithTimeout()`, `WithTokenRefresher()`
- Automatic retry on 5xx and 429 (rate limiting)
- Token refresh callback for 401 responses
- Pagination via `X-Total`, `X-Page` headers
- Proper response body closing in deferred functions with error checking
- Boolean filter parameters use Ruby on Rails convention with `?` suffix: `filter[alumni?]`, `filter[staff?]`

**Testing Conventions**:
- Table-driven tests with struct slices containing name, input, and expected output
- Test function naming: `TestFunctionName` for unit tests
- Test case naming: descriptive strings explaining the scenario (e.g., "none: nil cursusUser returns true")
- Use `t.Run()` for subtests with descriptive names
- Time-based tests: use `time.Now()` as reference, create relative dates with `AddDate()`
- Test edge cases: nil values, missing fields, past/future dates, boundary conditions
- Validation function tests: test all combinations of related flags and parameters

<!-- END AUTO-MANAGED -->

<!-- AUTO-MANAGED: patterns -->
## Detected Patterns

**Functional Options Pattern** (`internal/api/api.go`):
```go
type ClientOption func(*Client)
func WithBaseURL(baseURL string) ClientOption { ... }
client := NewClient(token, WithBaseURL(url), WithTimeout(30*time.Second))
```

**XDG Base Directory** (`internal/config/paths.go`):
- Respects `$XDG_CONFIG_HOME` on Linux/BSD
- Uses `~/Library/Application Support/` on macOS
- Uses `%APPDATA%` on Windows

**OAuth2 with PKCE** (`internal/oauth/pkce.go`):
- RFC 7636 compliant code verifier (64 random bytes, base64url encoded)
- SHA256 code challenge for secure authorization

**Automatic Token Refresh** (`cmd/root.go`, `internal/api/api.go`):
- Proactive refresh 5 minutes before expiry
- Reactive refresh on 401 response via callback

**Smart Endpoint Selection** (`cmd/user.go`):
- Detects when full cursus data is needed (level, blackhole filters)
- Switches from `/v2/campus/{id}/users` to `/v2/cursus_users` automatically when cursus-id specified
- Applies FilterActive to `/v2/cursus_users` queries (FilterAlumni not supported by this endpoint)
- Transforms `CursusUser` to `User` via `convertCursusUsersToUsers()` for unified filtering

**Data Structure Transformation** (`cmd/user.go`):
```go
func convertCursusUsersToUsers(cursusUsers []CursusUser, cursusID int) []User {
    // Embeds CursusUser data into User.CursusUsers slice
    // Enables unified filtering across different API endpoint responses
}
```

**Client-Side Filtering** (`cmd/user.go`):
- API-unsupported filters applied after fetch (min-level, blackhole-status)
- Reduces API complexity while maintaining rich filtering capabilities

**Blackhole Status Filtering** (`cmd/user.go`, `cmd/user_test.go`):
- Four status types: `none` (no blackhole), `active` (has blackhole, not ended), `past` (blackholed and ended), `upcoming` (blackhole within threshold)
- Critical validation: `past` status requires both `BlackholedAt` in past AND `EndAt` set (distinguishes truly blackholed from active users)
- Time-relative logic using `time.Now()` for past/future comparisons
- `upcoming` status uses configurable day threshold for "approaching blackhole" detection

<!-- END AUTO-MANAGED -->

<!-- AUTO-MANAGED: git-insights -->
## Git Insights

**Git Hooks** (via Lefthook):
- `pre-commit`: Auto-fix lint issues with `golangci-lint run --fix`
- `pre-push`: Run lint check before push

**Branch Strategy**:
- `main` branch for releases
- Feature branches: `claude/feature-name-*`

**Recent Feature Additions** (commits):
- `55c7b4a`: Fixed ListCursusUsers to apply FilterActive, removed FilterAlumni (endpoint doesn't support it)
- `33a46f9`: Extracted truncateString helper to util.go for code reuse and clarity
- `cb862ba`: Comprehensive test suite for blackhole status filter logic
- `94c56e8`: Improved validation and error handling (mutually exclusive flags, better error messages)
- `7a5b598`: Fixed ListCampusUsers to apply filter options correctly

**Key Implementation Decisions**:
- Smart endpoint selection: Use `/v2/cursus_users` when level/blackhole data needed, fall back to `/v2/campus/{id}/users` for minimal data
- Endpoint-specific filters: `/v2/cursus_users` supports FilterActive but not FilterAlumni; `/v2/users` supports both
- Boolean filter parameters: Use Rails convention `filter[alumni?]` and `filter[staff?]` (with `?` suffix)
- Automatic token refresh: Proactive (5 min before expiry) + reactive (on 401 responses)
- Client-side filtering: Apply unsupported API filters (min-level, blackhole-status) after fetch
- Flag validation: Mutually exclusive flags checked early (--active/--inactive, --alumni/--non-alumni)
- Endpoint compatibility validation: Prevent incompatible flag combinations (e.g., --alumni with --cursus-id) with helpful error messages
- Blackhole "past" status: Requires EndAt field to distinguish truly blackholed users from active ones
- Error messages: Campus name lookup failure shows available options (first 10) to guide users
- Code reuse: Extract common utilities (truncateString) to util.go for shared use across commands

<!-- END AUTO-MANAGED -->

<!-- MANUAL -->
## Notes

Add project-specific notes here. This section is never auto-modified.

<!-- END MANUAL -->
