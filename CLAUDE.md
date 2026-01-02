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

```
t42-cli/
├── main.go                 # Entry point - calls cmd.Execute()
├── cmd/                    # CLI commands (Cobra)
│   ├── root.go             # Root command, global flags, API client factory
│   ├── auth.go             # OAuth2 login/logout/status commands
│   ├── user.go             # User list/show commands with filters
│   └── project.go          # Project list/show/clone commands
├── internal/
│   ├── api/                # 42 API client
│   │   ├── api.go          # HTTP client with retry, pagination, token refresh
│   │   ├── types.go        # API response types (User, Project, Campus, etc.)
│   │   └── api_test.go     # API client tests
│   ├── config/             # Configuration management
│   │   ├── config.go       # Credentials and config loading/saving
│   │   ├── paths.go        # XDG-compliant path resolution
│   │   └── config_test.go  # Config tests
│   └── oauth/              # OAuth2 PKCE implementation
│       ├── pkce.go         # RFC 7636 PKCE code verifier/challenge
│       └── pkce_test.go    # PKCE tests
├── Makefile                # Build automation
├── go.mod                  # Go module definition
├── lefthook.yml            # Git hooks configuration
└── docs/                   # Documentation
```

**Data Flow**:
1. CLI commands in `cmd/` parse flags and call API methods
2. `cmd/root.go:NewAPIClient()` creates authenticated API client
3. `internal/api/` handles HTTP requests with automatic token refresh
4. `internal/config/` manages credentials and XDG-compliant file paths
5. `internal/oauth/` generates PKCE parameters for secure OAuth2 flow

<!-- END AUTO-MANAGED -->

<!-- AUTO-MANAGED: conventions -->
## Code Conventions

**Go Style**:
- Standard Go formatting (`gofmt`)
- Package names: lowercase, single word (`api`, `config`, `oauth`)
- File names: lowercase with underscores (`api_test.go`)
- Struct names: PascalCase (`Client`, `Credentials`, `PKCEParams`)
- Private functions: camelCase (`makeRequest`, `handleResponse`)

**Error Handling**:
- Wrap errors with context: `fmt.Errorf("failed to X: %w", err)`
- Return early on errors (guard clauses)
- Check errors immediately after function calls

**CLI Patterns**:
- Use Cobra for command structure
- Global flags in `rootCmd.PersistentFlags()`
- Use `--json` flag for JSON output
- Use `-v/--verbose` for debug output

**API Client Patterns**:
- Functional options pattern: `WithBaseURL()`, `WithTimeout()`
- Automatic retry on 5xx and 429 (rate limiting)
- Token refresh callback for 401 responses
- Pagination via `X-Total`, `X-Page` headers

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
- Switches from `/v2/campus/{id}/users` to `/v2/cursus_users` automatically
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

<!-- END AUTO-MANAGED -->

<!-- AUTO-MANAGED: git-insights -->
## Git Insights

**Git Hooks** (via Lefthook):
- `pre-commit`: Auto-fix lint issues with `golangci-lint run --fix`
- `pre-push`: Run lint check before push

**Branch Strategy**:
- `main` branch for releases
- Feature branches: `claude/feature-name-*`

**Recent Feature Additions**:
- User query with 42cursus progress filters (commit 12e839d)
- Smart endpoint selection for optimal API data retrieval
- Level and blackhole status filtering for user listings

<!-- END AUTO-MANAGED -->

<!-- MANUAL -->
## Notes

Add project-specific notes here. This section is never auto-modified.

<!-- END MANUAL -->
