# Agent 8: Testing & Quality Assurance – t42 CLI

## 1. Unit/Integration Test Coverage
- **Current:**
  - `internal/api_test.go` provides integration tests for the API client, including config loading and a real API call to `/v2/cursus`.
  - No unit/integration tests found for CLI commands (e.g., `auth`, `project`, etc.) yet.
- **How to run:**
  - `go test ./internal` (integration)
- **Recommendation:**
  - Add unit tests for each command in `cmd/` (e.g., using Cobra's test helpers).

## 2. Auth/API Error Testing
- **Current:**
  - `internal/api_test.go` tests API error handling (skips if no credentials, fails on HTTP error).
  - No explicit tests for `auth` command error flows (e.g., invalid credentials, token expiry).
- **Recommendation:**
  - Add tests for `auth` command covering login, logout, status, and error cases.

## 3. CLI UX Testing (Help, Errors, Completions, JSON Output)
- **Current:**
  - No automated tests found for CLI UX (help output, error messages, completions, JSON output).
- **Recommendation:**
  - Add tests/scripts to verify:
    - `t42 --help` and all subcommands' help output
    - Error messages for invalid usage
    - Completion script generation (`t42 completion -s <shell>`)
    - JSON output for commands supporting `--json`

## 4. Lint/Format
- **Current:**
  - No explicit linting or formatting config found, but Go projects use:
    - `go fmt ./...` for formatting
    - `go vet ./...` for static analysis
- **Recommendation:**
  - Add CI step to run `go fmt` and `go vet` on all code.

## 5. Recommendations/Next Steps
- Add unit/integration tests for all CLI commands in `cmd/`
- Add error/edge case tests for authentication and API
- Add CLI UX tests (help, errors, completions, JSON)
- Set up CI to enforce lint/format checks
- Reference: see `.rules/agents/2_api.md` for API test details 