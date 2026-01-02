# t42-cli

[![Go CI](https://github.com/naokiiida/t42-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/naokiiida/t42-cli/actions/workflows/ci.yml)

A command-line interface for the 42 API with OAuth2 authentication and PKCE support.

## Features

- OAuth2 authentication with PKCE (Proof Key for Code Exchange)
- Automatic token refresh
- XDG Base Directory specification compliance
- Support for multiple configuration methods
- JSON and table output formats

## Install

### From Source

```sh
go install github.com/naokiiida/t42-cli@latest
```

### Prebuilt Binaries

Prebuilt binaries will be available on the [Releases](https://github.com/naokiiida/t42-cli/releases) page after the first release.

### Install Script

```sh
curl -sSfL https://raw.githubusercontent.com/naokiiida/t42-cli/main/install.sh | sh
```

## Quick Start

### 1. Get OAuth2 Credentials

Visit [https://profile.intra.42.fr/oauth/applications](https://profile.intra.42.fr/oauth/applications) and create a new application to get your:
- Client ID (`FT_UID`)
- Client Secret (`FT_SECRET`)

### 2. Configure OAuth2 Secrets

Choose one of the following methods:

#### Option A: Config File (Recommended for deployment)

Create a secrets file in your config directory:

**Linux/BSD:**
```bash
mkdir -p ~/.config/t42
cat > ~/.config/t42/secrets.env <<EOF
FT_UID=your_client_id
FT_SECRET=your_client_secret
EOF
chmod 600 ~/.config/t42/secrets.env
```

**macOS:**
```bash
mkdir -p ~/Library/Application\ Support/t42
cat > ~/Library/Application\ Support/t42/secrets.env <<EOF
FT_UID=your_client_id
FT_SECRET=your_client_secret
EOF
chmod 600 ~/Library/Application\ Support/t42/secrets.env
```

**Windows (PowerShell):**
```powershell
New-Item -Path "$env:APPDATA\t42" -ItemType Directory -Force
Set-Content -Path "$env:APPDATA\t42\secrets.env" -Value @"
FT_UID=your_client_id
FT_SECRET=your_client_secret
"@
```

#### Option B: Environment Variables

```bash
export FT_UID=your_client_id
export FT_SECRET=your_client_secret
```

#### Option C: Development Mode

For local development, create `secret/.env` in the project directory:

```bash
echo "FT_UID=your_client_id" > secret/.env
echo "FT_SECRET=your_client_secret" >> secret/.env
```

### 3. Authenticate

```bash
t42 auth login
```

This will open your browser for OAuth2 authentication. After authorizing, you're ready to use the CLI!

### 4. Verify Authentication

```bash
t42 auth status
```

## Configuration Priority

The CLI checks for OAuth2 secrets in this order:

1. **Development**: `secret/.env` (local project directory)
2. **XDG Config**: Platform-specific config directory
   - Linux/BSD: `~/.config/t42/secrets.env` (respects `$XDG_CONFIG_HOME`)
   - macOS: `~/Library/Application Support/t42/secrets.env`
   - Windows: `%APPDATA%\t42\secrets.env`
3. **Environment Variables**: `FT_UID` and `FT_SECRET`

## Usage

```bash
# Authenticate
t42 auth login
t42 auth status
t42 auth logout

# User management
t42 user list                              # List users with filters
t42 user list --campus tokyo --cursus-id 21  # Filter by campus and cursus
t42 user list --blackhole-status upcoming  # Users with upcoming blackhole
t42 user list --min-projects 10 --active   # Active users with 10+ projects
t42 user show <login>                      # Show detailed user information

# Projects
t42 project list                # List projects
t42 project list --mine         # List your projects
t42 project show <slug>         # Show project details
t42 project clone <slug>        # Clone project repository
t42 project clone-mine <slug>   # Clone your project repository

# JSON output
t42 user list --json
t42 project list --json

# Verbose mode
t42 auth login -v
```

## Documentation

- [Deployment Guide](docs/deployment.md) - Detailed deployment and configuration
- [Secret Management](docs/secret_management.md) - OAuth2 secret rotation and security
- [OAuth2 Implementation](docs/oauth2_implementation.md) - PKCE implementation details
- [Architecture](docs/architecture.md) - System architecture overview

## Security

This CLI implements OAuth2 with PKCE (Proof Key for Code Exchange) for enhanced security. See [docs/secret_management.md](docs/secret_management.md) for details on:

- Why PKCE is important for native applications
- Secret rotation procedures
- Security best practices

## CI & Release

- Automated tests and linting via GitHub Actions
- Release automation via [goreleaser](https://goreleaser.com/)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[Add your license here] 