# t42-cli Architecture

This document outlines the architecture of the `t42-cli`, a command-line interface for interacting with the 42 API. The design emphasizes modularity, security, and a clean separation of concerns to ensure the codebase is maintainable and extensible.

## 1. Guiding Principles

- **Separation of Concerns**: The CLI's presentation layer (commands) is decoupled from the business logic (API interaction) and configuration management.
- **Security First**: Sensitive data like API credentials are never hardcoded or stored insecurely. Development secrets are kept separate from user credentials.
- **User-Centric Design**: The tool prioritizes a good user experience, using interactive prompts (`huh`) and clear, consistent command structures (`cobra`).
- **Portability**: The application is a single, self-contained binary that runs on major operating systems (macOS, Linux, Windows) without external dependencies.

## 2. Directory Structure

The project follows a standard Go project layout to organize its components effectively.

```
t42-cli/
├── .github/              # GitHub Actions workflows for CI/CD
├── .goreleaser.yml       # Release automation configuration
├── .rules/               # Project documentation and agent instructions
├── cmd/                  # All CLI commands, one file per command
│   ├── auth.go           # `t42 auth` command
│   ├── project.go        # `t42 project` command (includes clone-mine for repo_url)
│   ├── user.go           # `t42 user` command (list, show with filtering)
│   └── root.go           # Root command setup (`t42`)
├── docs/                 # Project documentation
│   └── architecture.md   # This file
├── internal/             # Private application and library code
│   ├── api/              # 42 API client wrapper
│   │   ├── api.go        # Core API client, request logic, retries
│   │   └── types.go      # Go types for API responses
│   └── config/           # Configuration and credentials management
│       ├── config.go     # Loading/saving user config and credentials
│       └── paths.go      # Logic for resolving config file paths
├── secret/               # Development-only secrets (gitignored)
│   └── .env              # CLIENT_ID, REDIRECT_URL for OAuth flow
└── main.go               # Main application entry point
```

- **`cmd/`**: Contains all user-facing commands built with `cobra`. Each command is responsible for parsing flags, handling user input (via `huh`), and calling the appropriate `internal` packages to perform its task. It should contain minimal business logic.
- **`internal/`**: This is the core of the application.
    - **`internal/api`**: A dedicated package that acts as a wrapper around the 42 API. It handles HTTP requests, authentication (attaching the bearer token), pagination, rate limiting, and parsing JSON responses into Go structs. All API interactions from the `cmd/` layer must go through this client. This includes access to project user data with team repositories via the `repo_url` field.
    - **`internal/config`**: Manages loading and saving all configuration and credential files. It provides a simple interface for the rest of the application to access configuration values without needing to know the underlying storage details.
- **`secret/`**: This directory is explicitly for development and is included in `.gitignore`. The `.env` file within it stores the `CLIENT_ID` and other secrets required to run the application locally for testing the OAuth authentication flow. This ensures that developer secrets are never committed to version control.

## 3. Configuration Management

Configuration is split into three distinct types to ensure security and clarity.

### a. Session Credentials (`credentials.json`)

This file stores the OAuth2 access token response from the 42 API. It is the most sensitive piece of user data and is treated as a session cookie.

- **Purpose**: To store the `access_token`, `refresh_token`, `expires_in`, and other token-related data for making authenticated API calls.
- **Location**: Stored in a secure, OS-specific user configuration directory. The path is determined by Go's `os.UserConfigDir()`.
    - **macOS**: `~/Library/Application Support/t42/credentials.json`
    - **Linux**: `~/.config/t42/credentials.json`
    - **Windows**: `%APPDATA%\t42\credentials.json`
- **Security**: File permissions are set to `0600` (read/write for user only). The `client_id` and `client_secret` are **never** stored here.

### b. User Configuration (`config.yaml`)

This file stores non-sensitive, user-configurable settings that control the CLI's behavior.

- **Purpose**: To allow users to customize the tool, e.g., setting the default output format (`--json`), disabling interactive prompts, or defining command aliases.
- **Location**: Stored alongside `credentials.json` in the user's config directory.
- **Format**: YAML is chosen for its human-readability.

### c. Development Secrets (`secret/.env`)

This file provides the necessary credentials for developers to test the application locally, particularly the OAuth Web Application Flow.

- **Purpose**: To store the application's `CLIENT_ID` and `REDIRECT_URL` needed for the initial authentication handshake. The `CLIENT_SECRET` may also be included for automated flows.
- **Location**: `secret/.env` in the project root.
- **Security**: This directory and its contents are listed in `.gitignore` and must never be committed to the repository.

### d. Setting the `T42_ENV` Environment Variable

The `T42_ENV` environment variable controls which configuration and credential paths are used by the application. This is especially important during development and testing.

- **Purpose**: When `T42_ENV` is set to `development`, the CLI will use credentials and configuration files from the local `secret/` directory (e.g., `secret/credentials.json`) instead of the OS user config directory. This allows developers to safely test authentication and integration flows without affecting real user data.
- **How to Set**: You should set `T42_ENV=development` in your shell environment before running the CLI or integration tests. For example:
    - On Unix/macOS:
      ```
      export T42_ENV=development
      ```
    - On Windows (Command Prompt):
      ```
      set T42_ENV=development
      ```
    - On Windows (PowerShell):
      ```
      $env:T42_ENV="development"
      ```
- **When to Set**: Always set `T42_ENV=development` when running integration tests or developing locally. In production or normal user usage, do not set this variable; the CLI will default to using the user's OS-specific config directory.

## 4. Component Interaction Flow

Here is a typical flow for a command like `t42 project list`:

1.  **User**: Runs `t42 project list`.
2.  **`main.go`**: Executes the `cobra` root command.
3.  **`cmd/project.go`**:
    - The `project list` command is invoked.
    - It calls `internal/config` to load the `credentials.json` file.
    - If credentials exist, it initializes the API client from `internal/api` with the loaded access token.
    - It calls the `ListProjects()` method on the API client.
4.  **`internal/api/api.go`**:
    - The `ListProjects()` method constructs the HTTP `GET` request to `/v2/projects`.
    - It adds the `Authorization: Bearer <token>` header.
    - It executes the request, handling potential retries and rate limiting.
    - It parses the JSON response into Go structs. If the response is paginated, it handles fetching subsequent pages as needed.
5.  **`cmd/project.go`**:
    - Receives the list of projects from the API client.
    - Formats the data for display in the terminal (e.g., as a table).
    - If an error occurred at any stage, it prints a user-friendly error message to `stderr`.

### Repository Cloning with `repo_url`

For commands that need to access individual team repositories (like `t42 project clone-mine`), the flow involves additional API calls:

1. **User**: Runs `t42 project clone-mine libft`.
2. **API Interaction**:
   - First, the command calls `GetMe()` to get the current user information.
   - Then, it calls `ListUserProjects()` to find the specified project in the user's project list.
   - Next, it calls `GetProjectUser()` using the `/v2/projects_users/:id` endpoint to get detailed team information.
   - From the team data, it extracts the `repo_url` field which contains the Git repository URL for the user's team submission.
3. **Git Clone**: The command uses the `repo_url` (not the base project `git_url`) to clone the user's actual submission repository.

This approach ensures that users clone their own team repositories rather than the base project template, which is crucial for accessing their actual work and submissions.

### User Listing with Smart Endpoint Selection

The `t42 user list` command implements intelligent endpoint selection based on the requested filters:

1. **User**: Runs `t42 user list --campus-id 26 --cursus-id 21 --min-level 5`.
2. **Smart Endpoint Detection**:
   - The command detects that `--min-level` requires full user data (level, blackhole info).
   - Standard endpoints (`/v2/campus/{id}/users`) return users with `cursus_users: null`.
   - When `--cursus-id` is specified with level/blackhole filters, it switches to `/v2/cursus_users`.
3. **API Interaction**:
   - Calls `ListCursusUsers()` with `filter[cursus_id]` and `filter[campus_id]`.
   - This endpoint returns `CursusUser` objects with embedded user and full cursus data.
4. **Data Transformation**:
   - `convertCursusUsersToUsers()` transforms `[]CursusUser` to `[]User` for unified filtering.
   - The cursus data is embedded into the user's `CursusUsers` slice.
5. **Client-Side Filtering**:
   - `filterUsers()` applies additional filters (min-level, blackhole-status) that aren't API-supported.
6. **Output**: Results are formatted as a table showing login, name, level, projects, and blackhole status.

This design ensures that users get accurate, filterable data while minimizing unnecessary API calls. See `docs/api_endpoints_guide.md` for detailed endpoint documentation.
