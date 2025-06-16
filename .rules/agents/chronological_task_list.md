# t42 CLI – Parallel Task List by Autonomous Agents

## Foundational Agents (Start in Parallel)

### Agent 1: Project Bootstrapper
- [x] Initialize Go module and repository
- [x] Scaffold CLI using Cobra (`cobra-cli`)
- [x] Set up directory structure for commands and internal packages
- [x] Integrate huh for interactive prompts and TUI/UX polish
- [x] Set up configuration and credential storage (secure, cross-platform)
- [x] Add basic logging and error handling utilities

### Agent 2: API Client Core
- [x] Implement minimal API client (HTTP, error handling, config)
- [x] Expand API client: support pagination, retries, rate limits, error handling
- [x] Use access token in `Authorization: Bearer ...` header for all requests
- [ ] Expose low-level `t42 api` passthrough for unsupported endpoints

---

## Feature Agents (Start Once API Client & Bootstrap Are Ready)

### Agent 3: Authentication
- [ ] Implement `t42 auth login` using OAuth2 Client Credentials Flow
- [ ] Store and securely manage access tokens
- [ ] Implement `t42 auth status` (token info, expiry, scopes, app roles)
- [ ] Implement `t42 auth logout` (clear credentials)
- [ ] Handle token refresh/expiry gracefully
- [ ] Write unit and integration tests for auth flows

### Agent 4: Project Commands
- [ ] Implement `t42 project list` (paginated, supports `--json`, `--web`)
- [ ] Implement `t42 project show PROJECT` (detailed info, Git URL)
- [ ] Implement `t42 pj clone PROJECT [-- <git flags>]`
- [ ] Implement `t42 pj view PROJECT`
- [ ] Add aliases (e.g., `pj` for `project`)

---

## UX & Polish Agents (Can Work in Parallel After Command Stubs Exist)

### Agent 5: UX, CLI Polish & Completion
- [ ] Write clear, actionable error messages (see Charm CLI design)
- [ ] Ensure all commands have `--help` and piped help goes to stdout
- [ ] Add `--json` output for scripting
- [ ] Add progress indicators (spinner/progress bar via huh)
- [ ] Default output <80 columns, add `--wide` option
- [ ] Ensure idempotent operations
- [ ] Design for pipeline-friendliness
- [ ] Add manpage and shell completions (`t42 completion -s <shell>`)
- [ ] Apply playful, human branding (see Charm branding)

---

## Supporting Agents (Can Work in Parallel After Core Features Exist)

### Agent 6: Install, Distribution & CI
- [x] Set up CI for cross-platform builds (GitHub Actions)
- [x] Automate releases with goreleaser
- [x] Provide install scripts (e.g., `curl | sh`)
- [ ] Package for Homebrew, Scoop, .deb/.rpm, universal script

### Agent 7: Documentation & Examples
- [ ] Write a README with GIFs, quick reference, and real examples
- [ ] Add usage examples for all commands
- [ ] Document API integration and authentication flow
- [ ] Add contribution guidelines

### Agent 8: Testing & Quality Assurance
- [~] Write unit and integration tests for all commands
- [~] Test authentication and API error cases
- [~] Test CLI UX (help, errors, completions, JSON output)
- [~] Lint and format code (go fmt, go vet)

### Agent 8 QA progress: see .rules/agents/8_qa.md for details

---

## Post-MVP / Stretch Goals (Parallelizable)
- [ ] Implement `t42 project register PROJECT` (sign up via POST)
- [ ] Implement `t42 project retry PROJECT` (retry logic via PATCH)
- [ ] Implement `t42 session PROJECT` (fetch project sessions)
- [ ] Implement `t42 status` (overview of session data, skills, retries)
- [ ] Add more advanced TUI features with huh
- [ ] Add analytics/telemetry (opt-in)
- [ ] use go-keyring for OS keychain support
---

**Key Points:**
- Bootstrapper and API Client can start immediately and in parallel.
- Auth and Project Commands can start as soon as the API client and CLI skeleton are ready.
- UX, CI, Docs, and Testing can proceed in parallel once command stubs exist.
- Post-MVP features can be tackled as soon as the relevant infrastructure is in place.

**References:**
- [Cobra Docs](https://github.com/spf13/cobra)
- [huh Docs](https://github.com/charmbracelet/huh)
- [42 API Guide](https://api.intra.42.fr/apidoc/guides/getting_started)
- [Charm CLI/README/Branding](../charm.md) 