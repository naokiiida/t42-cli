# Agent 3: Authentication – Implementation & Plan

## Summary
Implements authentication for the t42 CLI using the **OAuth2 Web Application Flow (Authorization Code Flow)** by default, as recommended by the 42 API documentation. Provides a secure, user-friendly login experience by redirecting the user to the browser for authentication and handling the callback to obtain an access token. **Fallback to Client Credentials Flow** is available for advanced users or automation.

## Why Web Application Flow?
- **No manual UID/SECRET input by default:** Users authenticate via the 42 website, not by entering secrets.
- **Security:** The user authenticates via the 42 website, and the CLI receives an access token via a secure redirect/callback.
- **Standard practice:** This is the recommended approach for user-facing CLI tools and matches the 42 API's best practices ([see docs](https://api.intra.42.fr/apidoc/guides/web_application_flow)).

## Command Options
- `t42 auth login` (default):
    - Uses browser-based OAuth2 Web Application Flow (Authorization Code Flow).
    - Starts a local HTTP server for the callback.
    - Opens the browser to the 42 authorize URL.
    - Shows the redirect URL in the terminal in case the browser does not open automatically.
    - Prompts the user to paste the code if `--no-localhost` is used (manual entry, no local server).
    - Supports custom port for the redirect URL with `--redirect-port <port>`.
    - Supports custom client secret file with `--creds <file>` (defaults to OS config dir as previously implemented).
- `t42 auth login --with-secret`:
    - Fallback: Prompts for UID and SECRET, uses OAuth2 Client Credentials Flow (legacy/manual/automation mode).
    - For advanced users, CI, or automation only.
- `-h, --help`: Show help for the command.

## Flow Overview
1. **CLI generates a random local redirect URI** (e.g., `http://localhost:PORT/callback`).
2. **CLI opens the user's browser** to the 42 OAuth2 authorize URL with the correct parameters (`client_id`, `redirect_uri`, `response_type=code`, `scope`, `state`).
3. **User logs in and authorizes the app** in the browser.
4. **42 redirects back to the CLI's local server** with a `code` and `state` (unless `--no-localhost` is used, in which case the user copies the code manually).
5. **CLI exchanges the code for an access token** using the app's `client_id` and `client_secret` (from the creds file).
6. **CLI stores the access token** for future API requests.

## Implementation Tasks
- [ ] Implement `t42 auth login` using the OAuth2 Web Application Flow:
    - [ ] Start a local HTTP server to receive the callback (unless `--no-localhost`).
    - [ ] Generate a secure random `state` parameter.
    - [ ] Open the browser to the 42 authorize URL with the correct parameters.
    - [ ] Show the redirect URL in the terminal for manual use.
    - [ ] Handle the callback, validate `state`, and extract the `code`.
    - [ ] If `--no-localhost`, prompt the user to paste the code from the browser.
    - [ ] Exchange the code for an access token via POST to `/oauth/token`.
    - [ ] Store the access token securely.
    - [ ] Show success or error to the user.
    - [ ] Support `--creds <file>` for custom client secret file.
    - [ ] Support `--redirect-port <port>` for custom port.
- [ ] Implement `t42 auth login --with-secret` as a fallback:
    - [ ] Prompt for UID and SECRET, use OAuth2 Client Credentials Flow.
    - [ ] Store the access token securely.
    - [ ] Show success or error to the user.
- [x] Implement `t42 auth status` (already implemented):
    - Loads the access token and displays token info (expiry, scopes, app roles, etc.).
- [x] Implement `t42 auth logout` (already implemented):
    - Deletes the credentials file, logging the user out.
- [ ] Handle token refresh/expiry gracefully (future work).
- [ ] Write unit and integration tests for auth flows.

## References
- [42 API Web Application Flow](https://api.intra.42.fr/apidoc/guides/web_application_flow)
- [OAuth2 Authorization Code Flow](https://datatracker.ietf.org/doc/html/rfc6749#section-4.1)

## Notes
- The CLI must be registered as an OAuth2 application in the 42 API console, with the redirect URI set to `http://localhost:PORT/callback` (or similar).
- The CLI should use a secure, random port for the local server and validate the `state` parameter to prevent CSRF attacks.
- The user's browser will be opened automatically for login, but the redirect URL will always be shown for manual use.
- The CLI should provide clear instructions and error messages if the flow fails.
- Fallback to Client Credentials Flow is available for advanced/automation use only. 