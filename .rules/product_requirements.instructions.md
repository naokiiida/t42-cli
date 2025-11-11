# t42 CLI – Product Requirements

## 💡 Vision: “t42” CLI, GitHub‑style, with Charm

- **Single self‑contained binary** (Go, using `cobra` and `huh`), cross-platform (macOS, Linux, Windows).
- **No external dependencies**; everything baked into the binary.
- **Modular subcommands** (like `gh`), with clear help, version, tab‑completion, and JSON output for scripting.
- **Installable via releases**: users can `curl | sh` or use package managers, with a single asset per platform.
- **Human, approachable branding**: playful, memorable, and developer-friendly, inspired by the Charm philosophy.

---

## 📦 Tech Stack & Build

- **Language**: Go (`cobra` for CLI, `huh` for TUI/UX polish).
- `log/slog` for logging with OpenTelemetry support
- **Cross‑compilation**: CI builds for `linux-amd64`, `linux-arm64`, `darwin-amd64`, `darwin-arm64`, `windows-amd64`.
- **Release automation**: GitHub Actions + `goreleaser`, single-archive installs per platform.
- **README**: Show, don’t tell—feature GIFs, quick references, and real examples (see Charm’s README approach).

---

## 🧭 Core Commands Design

Inspired by `gh` and best CLI practices:

```
t42 auth       # login, logout, status
t42 project    # list, view, register, clone, retry
t42 session    # view session rules/info
t42 status     # show summary/status of enrolled projects
t42 api        # low‑level passthrough for unsupported endpoints
t42 completion # shell completion
```

- **Aliases**: e.g., `t42 pj` for `t42 project`
- **Global flags**: `--help`, `--version`, `--json`, `--verbose`, `--quiet`, `--web`, `--wide`
- **Consistent, short, predictable subcommand names** (no camelCase)

---

## ✅ MVP

### `t42 auth`
- `t42 auth login`: OAuth2 Client Credentials Flow (see 42 API), or personal token.
- `t42 auth status`: Show authentication state (token info, scopes, expiry).
- `t42 auth logout`: Remove credentials.

### `t42 project`
- `t42 project list`: List your subscribed projects with statuses (paginated, supports `--json` and `--web`).
- `t42 project show PROJECT`: Detailed info incl. Git URL, Session Status, Teammates, evaluations, final mark, repository url
- `t42 pj clone PROJECT [-- <git flags>]`: Clone your latest project repo, passing extra flags to `git clone`.
- `t42 pj view PROJECT`: Display project description, and subject pdf URL
- `t42 pj evaluate `: show interactive view of all tasks as a reviewer
- `t42 pj evaluate PROJECT`: clone specified evaluation project
- `t42 pj evaluate --latest`: clone latest evaluation project repo

### `t42 completion`
- Generate shell completion scripts for bash/zsh/fish: `t42 completion -s <shell>`

---

## 🧩 UX & CLI Best Practices

- **Help docs**: `t42 help` and `--help` for all commands; piped help goes to stdout.
- **JSON support**: `--json` for programmatic parsing.
- **Progress indicators**: Use spinner/progress bar during network/API ops.
- **Reliable output**: Errors on stderr, standard exit codes.
- **Accessibility & width**: Output <80 columns by default, with `--wide` option.
- **Idempotent operations**: Running commands multiple times yields consistent results.
- **Error messages**: Clear, actionable, and help users learn the CLI.
- **Pipeline-friendly**: Output designed for piping and scripting.
- **Branding**: Name, logo, and tone should be playful, human, and memorable (see Charm’s branding advice).

---

## 🛠 Install & Distribution

- **CI**: Build and publish per-platform binaries.
- **Install scripts**: e.g., `curl -s https://github.com/you/t42/releases/latest/download/install.sh | sh`
- **Package managers**: Homebrew, Scoop, .deb/.rpm, universal script (see `ubi`).
- **README**: Feature GIFs, quick reference, and real examples up top.

---

## 🔗 42 API Integration

- **Auth**: Use OAuth2 Client Credentials Flow (`POST /oauth/token` with `grant_type=client_credentials`).
- **API requests**: Use access token in `Authorization: Bearer ...` header.
- **Pagination**: Support `page[number]`/`page[size]` params; parse `Link` headers for navigation.
- **Rate limits**: Respect 2 req/sec, 1200 req/hour.
- **Roles**: Display app roles from `X-Application-Roles` header.
- **Token info**: `GET /oauth/token/info` for debugging/status.

---

## ✅ Next Steps

1. Scaffold CLI using `cobra` and `huh`.
2. Implement MVP commands (`auth`, `project`, `completion`).
3. Integrate 42 API flows (auth, pagination, error handling).
4. Set up CI for cross-platform builds.
5. Polish UX: help, JSON, progress, error messages, branding.
6. Write README with GIFs, quick reference, and examples.

---

**References:**
- [42 API Guide](https://api.intra.42.fr/apidoc/guides/getting_started)
- [Charm CLI/README/Branding](https://github.com/charmbracelet/huh)
- [CLI UX Best Practices](https://evilmartians.com/chronicles/cli-ux-best-practices-3-patterns-for-improving-progress-displays)
