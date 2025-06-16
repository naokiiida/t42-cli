# t42 CLI – Development Task List

## 🏗️ 1. Project Scaffolding & Tooling
- [ ] Initialize Go module and repository
- [ ] Scaffold CLI using [Cobra](https://github.com/spf13/cobra) (`cobra-cli`)
- [ ] Set up directory structure for commands and internal packages
- [ ] Integrate [huh](https://github.com/charmbracelet/huh) for interactive prompts and TUI/UX polish
- [ ] Set up configuration and credential storage (secure, cross-platform)
- [ ] Add basic logging and error handling utilities

## 🔑 2. Authentication (42 API)
- [ ] Implement `t42 auth login` using OAuth2 Client Credentials Flow ([42 API Guide](https://api.intra.42.fr/apidoc/guides/getting_started#make-basic-requests))
- [ ] Store and securely manage access tokens
- [ ] Implement `t42 auth status` (show token info, expiry, scopes, app roles)
- [ ] Implement `t42 auth logout` (clear credentials)
- [ ] Handle token refresh/expiry gracefully

## 📚 3. Core Commands (MVP)
- [ ] Implement `t42 project list` (paginated, supports `--json`, `--web`)
- [ ] Implement `t42 project show PROJECT` (detailed info, Git URL)
- [ ] Implement `t42 pj clone PROJECT [-- <git flags>]` (clone latest repo, pass extra flags)
- [ ] Implement `t42 pj view PROJECT` (show project description)
- [ ] Implement `t42 completion -s <shell>` (generate shell completions)
- [ ] Add aliases (e.g., `pj` for `project`)

## 🌐 4. 42 API Integration
- [ ] Implement API client with proper error handling and retries
- [ ] Support pagination (`page[number]`, `page[size]`), parse `Link` headers
- [ ] Respect rate limits (2 req/sec, 1200 req/hour)
- [ ] Use access token in `Authorization: Bearer ...` header
- [ ] Expose low-level `t42 api` passthrough for unsupported endpoints
- [ ] Display app roles from `X-Application-Roles` header
- [ ] Implement `GET /oauth/token/info` for debugging/status

## 🧩 5. UX & CLI Best Practices
- [ ] Write clear, actionable error messages (see [Charm CLI design](charm.md))
- [ ] Ensure all commands have `--help` and piped help goes to stdout
- [ ] Add `--json` output for scripting
- [ ] Add progress indicators (spinner/progress bar via huh)
- [ ] Default output <80 columns, add `--wide` option
- [ ] Ensure idempotent operations
- [ ] Design for pipeline-friendliness
- [ ] Add manpage and shell completions
- [ ] Apply playful, human branding (see [Charm branding](charm.md))

## 🚀 6. Install & Distribution
- [ ] Set up CI for cross-platform builds (GitHub Actions)
- [ ] Automate releases with [goreleaser](https://goreleaser.com/)
- [ ] Provide install scripts (e.g., `curl | sh`)
- [ ] Package for Homebrew, Scoop, .deb/.rpm, universal script

## 📖 7. Documentation & Examples
- [ ] Write a README with GIFs, quick reference, and real examples (see [Charm README advice](charm.md))
- [ ] Add usage examples for all commands
- [ ] Document API integration and authentication flow
- [ ] Add contribution guidelines

## 🧪 8. Testing & Quality
- [ ] Write unit and integration tests for all commands
- [ ] Test authentication and API error cases
- [ ] Test CLI UX (help, errors, completions, JSON output)
- [ ] Lint and format code (go fmt, go vet)

## 🏁 9. Post-MVP / Stretch Goals
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
- [Charm CLI/README/Branding](charm.md) 