# t42 CLI – Chronological Task List by Autonomous Agents

## Agent 1: Project Bootstrapper
- [ ] Initialize Go module and repository
- [ ] Scaffold CLI using Cobra (`cobra-cli`)
- [ ] Set up directory structure for commands and internal packages
- [ ] Integrate huh for interactive prompts and TUI/UX polish
- [ ] Set up configuration and credential storage (secure, cross-platform)
- [ ] Add basic logging and error handling utilities
- [ ] Handoff: Notify Agent 2 when scaffolding and basic infra are ready

---

## Agent 2: Auth & Minimal API Client
- [ ] Implement minimal API client (HTTP, error handling, config)
- [ ] Implement `t42 auth login` using OAuth2 Client Credentials Flow
- [ ] Store and securely manage access tokens
- [ ] Implement `t42 auth status` (token info, expiry, scopes, app roles)
- [ ] Implement `t42 auth logout` (clear credentials)
- [ ] Handle token refresh/expiry gracefully
- [ ] Write unit and integration tests for auth flows
- [ ] Handoff: Notify Agent 3 when auth and minimal API client are tested and stable

---

## Agent 3: Core API Client & Project Commands
- [ ] Expand API client: support pagination, retries, rate limits, error handling
- [ ] Use access token in `Authorization: Bearer ...` header for all requests
- [ ] Implement `t42 project list` (paginated, supports `--json`, `--web`)
- [ ] Implement `t42 project show PROJECT` (detailed info, Git URL)
- [ ] Implement `t42 pj clone PROJECT [-- <git flags>]`
- [ ] Implement `t42 pj view PROJECT`
- [ ] Add aliases (e.g., `pj` for `project`)
- [ ] Expose low-level `t42 api` passthrough for unsupported endpoints
- [ ] Handoff: Notify Agent 4 when core commands are functional and tested

---

## Agent 4: UX, CLI Polish & Completion
- [ ] Write clear, actionable error messages (see Charm CLI design)
- [ ] Ensure all commands have `--help` and piped help goes to stdout
- [ ] Add `--json` output for scripting
- [ ] Add progress indicators (spinner/progress bar via huh)
- [ ] Default output <80 columns, add `--wide` option
- [ ] Ensure idempotent operations
- [ ] Design for pipeline-friendliness
- [ ] Add manpage and shell completions (`t42 completion -s <shell>`)
- [ ] Apply playful, human branding (see Charm branding)
- [ ] Handoff: Notify Agent 5 when CLI polish and UX are complete

---

## Agent 5: Install, Distribution & CI
- [ ] Set up CI for cross-platform builds (GitHub Actions)
- [ ] Automate releases with goreleaser
- [ ] Provide install scripts (e.g., `curl | sh`)
- [ ] Package for Homebrew, Scoop, .deb/.rpm, universal script
- [ ] Handoff: Notify Agent 6 when distribution pipeline is tested

---

## Agent 6: Documentation & Examples
- [ ] Write a README with GIFs, quick reference, and real examples
- [ ] Add usage examples for all commands
- [ ] Document API integration and authentication flow
- [ ] Add contribution guidelines
- [ ] Handoff: Notify Agent 7 when docs are ready

---

## Agent 7: Testing & Quality Assurance
- [ ] Write unit and integration tests for all commands
- [ ] Test authentication and API error cases
- [ ] Test CLI UX (help, errors, completions, JSON output)
- [ ] Lint and format code (go fmt, go vet)
- [ ] Handoff: Notify Agent 8 for post-MVP features

---

## Agent 8: Post-MVP / Stretch Goals
- [ ] Implement `t42 project register PROJECT` (sign up via POST)
- [ ] Implement `t42 project retry PROJECT` (retry logic via PATCH)
- [ ] Implement `t42 session PROJECT` (fetch project sessions)
- [ ] Implement `t42 status` (overview of session data, skills, retries)
- [ ] Add more advanced TUI features with huh
- [ ] Add analytics/telemetry (opt-in)

---

**References:**
- [Cobra Docs](https://github.com/spf13/cobra)
- [huh Docs](https://github.com/charmbracelet/huh)
- [42 API Guide](https://api.intra.42.fr/apidoc/guides/getting_started)
- [Charm CLI/README/Branding](../charm.md) 