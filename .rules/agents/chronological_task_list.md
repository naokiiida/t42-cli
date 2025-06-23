# t42-cli Chronological Task List (Architecture-Aligned)

This document outlines the development plan for the `t42-cli`, organized into sequential phases. Each phase builds upon the previous one and includes explicit, reproducible testing commands to ensure quality and correctness.

---

## Phase 1: Foundational Packages (`internal/`)

This phase focuses on creating the core, non-UI logic for configuration and API communication.

### Agent 1: Configuration Manager
- **Goal**: Implement robust, secure, and layered configuration management.
- **Tasks**:
    - [ ] **`internal/config/paths.go`**: Implement functions to resolve OS-specific paths for config files (`os.UserConfigDir`).
    - [ ] **`internal/config/config.go`**:
        - [ ] Implement loading/saving of session credentials (`credentials.json`) with secure `0600` file permissions.
        - [ ] Implement loading/saving of user preferences (`config.yaml`).
        - [ ] Implement loading of development secrets from `secret/.env`.
- **Testing**:
    - **How**: Unit tests are required to verify all configuration logic without building the full application.
    - **Command**:
      ```sh
      go test ./internal/config
      ```
    - **Verification**: The test suite should pass, confirming that paths are resolved correctly for the host OS and that files are read and written with the correct permissions and content.

### Agent 2: 42 API Client
- **Goal**: Create a reusable, low-level client for all 42 API interactions.
- **Tasks**:
    - [ ] **`internal/api/types.go`**: Define Go structs for common API responses (e.g., `Token`, `Project`, `ErrorResponse`).
    - [ ] **`internal/api/api.go`**:
        - [ ] Implement `NewClient` to initialize the client with an access token.
        - [ ] Create a core request method that adds the `Authorization: Bearer <token>` header.
        - [ ] Implement logic to handle API pagination and rate limiting.
        - [ ] Add structured error handling for 42 API error responses.
- **Testing**:
    - **How**: An integration test is needed to validate the client against the live 42 API. This requires a valid access token.
    - **Steps**:
      1.  **Prepare credentials**:
          - Create a `secret/.env` file in your project root with your 42 API credentials:
            ```
            FT_UID=your_client_id
            FT_SECRET=your_client_secret
            ```
          - These variable names (`FT_UID` and `FT_SECRET`) are used for consistency in all development and test scripts.
      2.  **Get an access token**:
          ```sh
          # Uses FT_UID and FT_SECRET from secret/.env
          eval $(cat secret/.env) && curl -X POST --data "grant_type=client_credentials&client_id=$FT_UID&client_secret=$FT_SECRET" https://api.intra.42.fr/oauth/token > secret/credentials.json
          ```
          - This command will save the token JSON to `secret/credentials.json` in the format:
            ```json
            {
              "access_token": "ACCESS_TOKEN",
              "token_type": "bearer",
              "expires_in": 2090,
              "scope": "public",
              "created_at": 1750672673,
              "secret_valid_until": 1752476801
            }
            ```
      3.  **Run the test**:
          ```sh
          T42_ENV=development go test ./internal/api -v
          ```
          - The integration test will look for `secret/credentials.json` when `T42_ENV=development` is set.
          - If the credentials file is missing or expired, the test will be skipped with a helpful message.
    - **Verification**: The test should pass and log a snippet of the API response (e.g., from `/v2/cursus`), confirming successful authentication and data retrieval.
    - **Reference**: See the [42 API documentation](https://api.intra.42.fr/apidoc/guides/getting_started) for details on authentication and token usage.
    - **Note**: The API client does not generate tokens; it only uses them. Token acquisition is handled externally (e.g., via the above `curl` command or by the `auth` command in later agents).

---

## Phase 2: CLI Scaffolding & Authentication

This phase builds the user-facing command structure and implements the primary authentication flow.

### Agent 3: CLI Bootstrapper
- **Goal**: Set up the main application skeleton using Cobra.
- **Tasks**:
    - [ ] **`main.go`**: Create the main entry point to execute the root command.
    - [ ] **`cmd/root.go`**: Initialize the root Cobra command (`t42`), add global flags (`--json`), and integrate logging.
- **Testing**:
    - **How**: Build the CLI and run its most basic commands.
    - **Commands**:
      ```sh
      go build .
      ./t42-cli --help
      ./t42-cli --version
      ```
    - **Verification**: The commands should execute without errors and print the help text and version number.

### Agent 4: Authentication Commands
- **Goal**: Implement the complete user authentication and session management lifecycle.
- **Tasks**:
    - [ ] **`cmd/auth.go`**: Create the `auth` command and its subcommands: `login`, `logout`, `status`.
    - [ ] **`auth login`**: Implement the OAuth2 Web Application Flow, including the local callback server.
    - [ ] **`auth logout`**: Implement credential file deletion.
    - [ ] **`auth status`**: Implement token info fetching and display.
- **Testing**:
    - **How**: Perform a full, end-to-end authentication lifecycle test.
    - **Steps**:
      1.  **Prepare**: Create `secret/.env` with your `CLIENT_ID` and `REDIRECT_URL`.
      2.  **Build**: `go build .`
      3.  **Login**:
          ```sh
          ./t42-cli auth login
          ```
          - **Verification**: The browser should open. After you authorize the app, `credentials.json` should be created in your user config directory.
      4.  **Check Status**:
          ```sh
          ./t42-cli auth status
          ```
          - **Verification**: The command should print your token's scope, expiry, and other details.
      5.  **Logout**:
          ```sh
          ./t42-cli auth logout
          ```
          - **Verification**: The `credentials.json` file should be deleted.

---

## Phase 3: Core Feature Commands

With authentication in place, this phase implements the core project-related features.

### Agent 5: Project Commands
- **Goal**: Build the `project` subcommand to manage user projects.
- **Tasks**:
    - [ ] **`cmd/project.go`**: Create the `project` command (`pj` alias) and its subcommands: `list`, `show`, `clone`.
- **Testing**:
    - **How**: After logging in, test each subcommand to ensure it interacts with the API correctly.
    - **Prerequisite**: You must be logged in. Run `./t42-cli auth login`.
    - **Commands**:
      ```sh
      # Test list command (table output)
      ./t42-cli project list

      # Test list command (JSON output)
      ./t42-cli project list --json

      # Test show command with a valid project slug
      ./t42-cli pj show libft

      # Test clone command
      ./t42-cli pj clone libft
      ```
    - **Verification**:
        - `list` and `show` should display formatted data from the API.
        - `clone` should create a new directory named `libft` containing the project's Git repository.

---

## Phase 4: Polish, Distribution & Documentation

This final phase focuses on improving the user experience, automating builds, and creating documentation.

### Agent 6: UX & Polish
- **Goal**: Refine the CLI to be professional, intuitive, and enjoyable to use.
- **Tasks**:
    - [ ] **Completions**: Implement `t42 completion`.
    - [ ] **Help Text**: Review and improve all command help messages.
    - [ ] **Progress Indicators**: Add `huh` spinners for all network operations.
- **Testing**:
    - **How**: Manually verify the UX enhancements.
    - **Commands**:
      ```sh
      # Generate and test completion script
      ./t42-cli completion zsh > _t42-cli
      source ./_t42-cli
      # (try tabbing after ./t42-cli p<TAB>)

      # Review help text for clarity
      ./t42-cli project clone --help
      ```
    - **Verification**: Completions should work, and help text should be clear and accurate. Spinners should appear during commands like `project list`.

### Agent 7: CI/CD & Release
- **Goal**: Automate the build, test, and release process.
- **Tasks**:
    - [ ] **CI**: Set up a GitHub Actions workflow to run `go build`, `go test`, and `go vet`.
    - [ ] **Release**: Configure `.goreleaser.yml` for cross-platform builds.
    - [ ] **Installation**: Create an `install.sh` script.
- **Testing**:
    - **How**: Run the release process locally and test the generated artifacts.
    - **Commands**:
      ```sh
      # Run a local, temporary release to the dist/ folder
      goreleaser release --snapshot --clean

      # Test the install script (requires a real release)
      # curl -sSL "https://github.com/user/repo/releases/latest/download/install.sh" | sh
      ```
    - **Verification**: The `dist/` directory should contain binaries for all target platforms. The CI pipeline should pass on GitHub.

### Agent 8: Documentation
- **Goal**: Create a high-quality README that encourages adoption.
- **Tasks**:
    - [ ] Write a comprehensive `README.md`.
    - [ ] Use `vhs` to create animated GIFs demonstrating key features.
- **Testing**:
    - **How**: Manually review the documentation and generate the assets.
    - **Commands**:
      ```sh
      # Example for generating a GIF from a tape file
      vhs < docs/tapes/auth.tape
      ```
    - **Verification**: The `README.md` should render correctly on GitHub, and the GIFs should accurately demonstrate the CLI's functionality.
