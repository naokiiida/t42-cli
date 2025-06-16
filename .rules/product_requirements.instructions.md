Here’s a set of product requirements based on your vision—mirroring GitHub’s `gh` CLI design, with a single-binary, cross-platform, no-external dependency model:

---

## 💡 Vision: “t42” CLI, GitHub‑style

* **Single self‑contained binary** (via Go or Rust) supporting macOS, Linux, and Windows.
* **No external dependencies**; everything baked into the binary.
* **Modular subcommands** (like `gh`) with clear help, version, tab‑completion, and JSON output for scripting.
* **Installable via releases**: users can `curl | sh` or use `winget`, `brew`, `scoop`—all via a single asset per platform.

---

## 📦 Tech Stack & Build

* **Language**: Go

  * Go: use `cobra` + `go build`.
* **Cross‑compilation**: Set up CI to produce binaries for `linux-amd64`, `linux-arm64`, `darwin-amd64`, `darwin-arm64`, `windows-amd64`.
* **Release automation**: Use GitHub Actions + `goreleaser`, outputting single-archive installs per platform ([github.blog][1], [github.com][2], [reddit.com][3]).

---

## 🧭 Core Commands Design

Inspired by `gh`, implement key domains:

```
t42 auth       # login, logout, status
t42 project    # list, view, register, clone, retry
t42 session    # view session rules/info
t42 status     # show summary/status of enrolled projects
t42 api        # low‑level passthrough for unsupported endpoints
t42 completion # shell completion
```

Plus global flags: `--help`, `--version`, `--json`, `--verbose`, `--quiet`.

---

## ✅ MVP

Implement these commands with minimal UI and requirements:

### `t42 auth`

* `t42 auth login`: OAuth flow or personal token.
* `t42 auth status`: Show authentication state.
* `t42 auth logout`.

### `t42 project`

* `t42 project list`: your subscribed projects with statuses.
* `t42 project show PROJECT`: detailed info incl. Git URL.

* support `pj` alias for `project`
* `t42 pj clone PROJECT`: clones your latest project url.
* `t42 pj clone PROJECT -- --depth=1` Clone a project with additional git clone flags
* `t42 pj view PROJECT`: display description of project.

Support for `--json` (`t42 project list|view --json`) helps automation ([boxpiper.com][4], [webinstall.dev][5]).
Support for `--web` (`t42 project list|view --web`) as a fallback

### `t42 completion`
Generate shell completion scripts for bash/zsh/fish
* `t42 completion -s <shell>`

---

## 🚀 Post‑MVP

Add these once MVP stabilized:

* `t42 project register PROJECT`: sign up using `POST /v2/projects/:id/register`.
* `t42 project retry PROJECT`: retry logic via `PATCH /v2/projects_users/:id/retry`.
* `t42 session PROJECT`: fetch via `GET /v2/projects/:id/project_sessions`.
* `t42 status`: show overview of session data, skills, retries.
* `t42 api` for direct endpoint access.

---

## 🧩 UX & CLI Best Practices

Informed by GitHub and general CLI UX principles:

* **Reliable output**: no silent failures; errors on stderr; exit codes standard ([evilmartians.com][6]).
* **Help docs**: `t42 help` and `--help` for all commands; piped help goes to stdout ([hackmd.io][7]).
* **JSON support**: `--json` for programmatic parsing ([github.blog][1]).
* **Progress indicators**: Use spinner or progress bar during network/API ops ([evilmartians.com][6]).
* **Accessibility & width**: Keep output <80 columns by default, with `--wide` option ([reddit.com][8]).
* **Idempotent operations**: Running commands multiple times yields consistent results .
* **Consistent naming**: Use short, predictable subcommand names (no camelCase) ([evilmartians.com][6]).

---

## 🛠 Install & Distribution

* Build CI publishing per-platform binaries.

* Provide install scripts like:

  ```bash
  curl -s https://github.com/you/t42/releases/latest/download/install.sh | sh
  ```

* Support package managers:

  * Linux: `.deb`, `.rpm`
  * macOS: Homebrew
  * Windows: Scoop, Scoop
  * Universal script: like `ubi` ([github.com][9])

---

## ✅ Next Steps

1. Pick language (Go).
2. Scaffold CLI using `cobra` and `huh` (Go).
3. Define MVP commands (`auth`, `project`, `completion`).
4. Set up CI for cross-platform builds.
5. Write initial prototyping CLI:
6. Add UX polish: JSON, progress, error messages, help.

---

Want help bootstrapping the code, GitHub Actions CI config, or specific endpoint implementation?

[1]: https://github.blog/engineering/user-experience/building-a-more-accessible-github-cli/ "Building a more accessible GitHub CLI"
[2]: https://github.com/cli/cli "cli/cli: GitHub's official command line tool - GitHub"
[3]: https://www.reddit.com/r/rust/comments/1irf87q/easyinstall_a_crossplatform_cli_installation_tool/ "A cross-platform CLI installation tool based on github release : r/rust"
[4]: https://www.boxpiper.com/posts/github-cli "GitHub CLI - GitHub and command line in 2025 - Box Piper"
[5]: https://webinstall.dev/gh "GitHub CLI - webinstall.dev"
[6]: https://evilmartians.com/chronicles/cli-ux-best-practices-3-patterns-for-improving-progress-displays "CLI UX best practices: 3 patterns for improving progress displays"
[7]: https://hackmd.io/%40arturtamborski/cli-best-practices "[cli-best-practices](https://hackmd.io/@arturtamborski/cli-best ..."
[8]: https://www.reddit.com/r/UXDesign/comments/11rc0mi/any_ux_designers_working_on_clicommand_line/ "Any UX designers working on CLI(command line interfaces ... - Reddit"
[9]: https://github.com/houseabsolute/ubi "houseabsolute/ubi: The Universal Binary Installer - GitHub"
