# deadenv

Replace `.env` files with encrypted, OS-native secret storage and a clean CLI developer experience.

`deadenv` is a cross-platform CLI tool written in Go that eliminates plaintext `.env` files from your development workflow. Instead of storing credentials as plaintext on disk, `deadenv` stores them in your operating system's native secret store — Keychain on macOS, Keyring on Linux, or Credential Manager on Windows.

Secrets are retrieved at runtime and injected directly into your subprocesses or shell session. They never touch the filesystem in plaintext.

---

## Table of Contents

- [Why deadenv?](#why-deadenv)
- [Key Features](#key-features)
- [Platform Support](#platform-support)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [CLI Reference](#cli-reference)
  - [Profile Management](#profile-management)
  - [Key Management](#key-management)
  - [Runtime Injection](#runtime-injection)
  - [Export & Import](#export--import)
  - [History & Audit](#history--audit)
  - [Shell Integration](#shell-integration)
- [Use Cases](#use-cases)
- [Architecture](#architecture)
- [Configuration](#configuration)
- [Error Handling](#error-handling)
- [Testing](#testing)
- [Contributing](#contributing)

---

## Why deadenv?

**The Problem:** `.env` files are security anti-patterns. They store credentials as plaintext on disk, risk accidental commits to git, and can be read by any process or user with filesystem access.

**The Solution:** `deadenv` uses your OS's native secret management:

- 🔐 **Encrypted Storage:** Secrets stored in Keychain, Keyring, or Credential Manager — not on disk
- 🔑 **OS-Native Auth:** Touch ID, biometrics, or device password gates access
- 📝 **Audit Trail:** Git-backed history tracking structural changes without storing values
- 📦 **Team Sharing:** Export encrypted profiles to safely share credentials with teammates
- 🚀 **Drop-in Replacement:** Works with existing `.env` file formats

---

## Key Features

✅ **Zero Plaintext Storage** — Secrets never touch the filesystem  
✅ **Cross-Platform** — macOS, Linux, Windows with native integration  
✅ **Biometric Support** — Touch ID and other OS authentication methods  
✅ **Profile-Based** — Organize secrets by environment or service  
✅ **Audit Logging** — Track who changed what and when (without exposing values)  
✅ **Secure Export/Import** — Share profiles encrypted with AES-256-GCM  
✅ **Editor Support** — Edit secrets interactively with `$EDITOR`  
✅ **Multiple Export Formats** — Shell scripts, fish syntax, JSON, eval  
✅ **Minimal Dependencies** — Pure Go with OS-native libraries only  

---

## Platform Support

| Platform    | Keychain Provider   | Status | Notes                              |
| ----------- | ------------------- | ------ | ---------------------------------- |
| **macOS**   | Security.framework  | ✅ Full | Touch ID & device password support |
| **Linux**   | libsecret / KWallet | ✅ Full | GNOME Keyring, KWallet support     |
| **Windows** | Credential Manager  | ✅ Full | Windows Hello integration ready    |

---

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/funinkina/deadenv.git
cd deadenv

# Build the binary
make build

# The binary is available at ./bin/deadenv
./bin/deadenv --help

# Optionally, install to your PATH
sudo mv ./bin/deadenv /usr/local/bin/
```

### Build Requirements

- **Go 1.26+**
- **macOS:** Xcode Command Line Tools (for cgo compilation)
- **Linux:** `libsecret-1-dev` and `pkg-config`
- **Windows:** Standard build tools

### Verify Installation

```bash
deadenv --help
```

---

## Quick Start

### 1. Create Your First Profile

```bash
# Interactive creation with editor
deadenv profile new myapp

# Create from an existing .env file
deadenv profile new myapp --from=.env.example

# List all profiles
deadenv profile list
```

The editor will open pre-filled with format instructions. Enter your secrets in `KEY=VALUE` format (one per line).

### 2. Add a Secret

```bash
# Set with value in argument
deadenv set myapp DATABASE_URL "postgresql://user:pass@localhost/db"

# Set interactively (hidden input)
deadenv set myapp API_KEY
```

### 3. Run Your App

```bash
# All secrets injected into subprocess environment
deadenv run myapp -- npm start

# Works with any command
deadenv run myapp -- python app.py
deadenv run myapp -- ./my-binary --config=prod
```

### 4. Export for Teammates

```bash
# Export the profile encrypted
deadenv export myapp --out=myapp.deadenv

# Share myapp.deadenv and the sharing password via separate channels
# Your teammate imports with:
deadenv import myapp.deadenv
```

---

## CLI Reference

### Profile Management

#### `deadenv profile new <name> [--from=<file>]`

Create a new profile.

```bash
# Create empty and edit interactively
deadenv profile new staging

# Create from existing .env file
deadenv profile new prod --from=.env.production

# Editor will show the file contents for you to confirm or edit
```

Profiles must use lowercase letters, digits, and hyphens (e.g., `api-service-dev`).

#### `deadenv profile list` (alias: `ls`)

List all available profiles.

```bash
deadenv profile ls
```

Output:
```
Profiles:
  • myapp-dev
  • myapp-staging
  • payments-service
```

#### `deadenv profile show <name> [--reveal]`

Show keys in a profile (values masked by default).

```bash
# Show with masked values (default)
deadenv profile show myapp

# Output:
# DATABASE_URL     [***]
# API_KEY          [***]
# DEBUG            public-value

# Reveal plaintext values (requires OS authentication)
deadenv profile show myapp --reveal
```

#### `deadenv profile delete <name>` (alias: `rm`)

Delete a profile and all its keys.

```bash
deadenv profile rm old-profile

# Requires confirmation unless --yes is passed
deadenv profile rm old-profile --yes
```

#### `deadenv profile rename <old> <new>`

Rename a profile (all keys moved to new name).

```bash
deadenv profile rename staging staging-old
```

#### `deadenv profile copy <src> <dst>`

Copy a profile to a new name (original stays intact).

```bash
deadenv profile copy myapp-dev myapp-staging
```

---

### Key Management

#### `deadenv set <profile> <KEY> [VALUE]`

Set a key in a profile.

```bash
# Set with inline value
deadenv set myapp DATABASE_URL "postgresql://localhost/mydb"

# Set interactively (hidden input prompt)
deadenv set myapp API_TOKEN

# Set empty value
deadenv set myapp DEBUG ""
```

**Keys must follow POSIX conventions:** uppercase letters, digits, underscores; cannot start with a digit.

#### `deadenv get <profile> <KEY> [--reveal]`

Retrieve a key's value (masked by default).

```bash
# Get masked
deadenv get myapp API_KEY
# Output: [***]

# Get plaintext (requires OS authentication)
deadenv get myapp API_KEY --reveal
# Output: sk_live_51234567890abcdef
```

#### `deadenv unset <profile> <KEY>`

Remove a key from a profile.

```bash
deadenv unset myapp OLD_CONFIG

# Requires confirmation unless --yes is passed
deadenv unset myapp OLD_CONFIG --yes
```

#### `deadenv edit <profile>`

Open the profile interactively in your editor.

```bash
deadenv edit myapp
```

This will:
1. Authenticate with OS (to read current values)
2. Open `$EDITOR` with all current keys pre-populated
3. Let you add, remove, or modify keys
4. Show a diff summary before applying changes
5. Write changes back to keychain with audit trail

**Changes are granular:** each modified key is a separate history entry.

---

### Runtime Injection

#### `deadenv run <profile> -- <command> [args...]`

Run a command with profile secrets injected into the environment.

```bash
# Basic usage
deadenv run myapp -- npm start

# With complex arguments
deadenv run myapp -- python -m flask run --host=0.0.0.0

# Chaining with pipes (entire expression runs with injected env)
deadenv run myapp -- bash -c "npm build && npm start"

# Database migrations with environment-specific connection string
deadenv run myapp-prod -- psql -c "CREATE TABLE ..."
```

**Important:** Use the `--` separator to prevent flag parsing conflicts.

The subprocess inherits all current environment variables plus the profile's secrets (profile values override existing vars, consistent with `.env` tools).

Exit code is propagated: if the command exits with code `42`, so does `deadenv`.

#### `deadenv export <profile> [--out=<file>] [--format=shell|fish|json]`

Export secrets for shell evaluation.

```bash
# Generate shell export commands (bash/zsh)
deadenv export myapp
# Output:
# export DATABASE_URL="postgresql://localhost/mydb"
# export API_KEY="sk_live_..."

# Eval into current shell (bash/zsh)
eval $(deadenv export myapp)

# Fish shell syntax
deadenv export myapp --format=fish | source

# JSON for machine consumption
deadenv export myapp --format=json
# Output: [{"key":"DATABASE_URL","value":"..."}...]

# Write to a shell script file
deadenv export myapp --out=./env.sh
source ./env.sh
```

**Values are properly shell-escaped.** The export is read-only; changes must be made with `deadenv set` or `deadenv edit`.

---

### Export & Import

#### `deadenv export <profile> --out=<file>`

Create an encrypted `.deadenv` export file (portable format for sharing).

```bash
# Export as encrypted file
deadenv export myapp --out=myapp.deadenv

# You will be prompted for a sharing password (separate from OS auth)
# Enter sharing password: ****
# Confirm password: ****
# ✓ Profile exported to myapp.deadenv
```

The export file is:
- **Self-contained:** includes all KDF parameters for decryption
- **Versioned:** supports future format migrations
- **AES-256-GCM encrypted:** uses a password-derived key (Argon2id)
- **Secure:** even with filesystem access, passwords are required to decrypt

#### `deadenv import <file> [--as=<profile>]`

Import an encrypted `.deadenv` file.

```bash
# Import with original profile name
deadenv import myapp.deadenv
# Enter sharing password: ****
# Import 8 keys from profile "myapp"? (y/n): y
# ✓ Profile "myapp" imported successfully

# Import as different profile name
deadenv import myapp.deadenv --as=myapp-imported
```

On import:
1. You're prompted for the sharing password
2. A summary of keys to import is shown
3. You confirm before any keychain writes
4. The import is recorded in the audit history

---

### History & Audit

#### `deadenv history <profile> [--key=<KEY>]`

View audit history of changes (structural changes only, no values).

```bash
# View full history for profile
deadenv history myapp

# Output:
# Profile: myapp
# ─────────────────────────────────────────────────
# 2025-04-21 14:30  set       API_KEY           (hash: a3f1b2...)
# 2025-04-21 14:25  modified  DATABASE_URL      (hash: 7c2d9e...)
# 2025-04-21 14:20  set       DEBUG             (hash: 5e41d3...)

# Filter to a specific key
deadenv history myapp --key=API_KEY

# Output:
# Profile: myapp (KEY: API_KEY)
# ────────────────────────────────
# 2025-04-21 14:30  set       (hash: a3f1b2...)
# 2025-04-19 10:15  modified  (hash: 9c8b4f...)
```

Each entry shows:
- **Timestamp:** when the change was made
- **Operation:** `set`, `modified`, `unset`, `delete`
- **Key name:** which secret was affected
- **Hash:** SHA-256(salt + value) — proves the value changed without revealing it

This history is stored in a local git repository (`~/.config/deadenv/history/`) for durability and traceability.

---

### Shell Integration

#### `deadenv init [--shell=zsh|bash|fish]`

Print a shell hook snippet for convenient profile switching.

```bash
# Generate for your shell
deadenv init --shell=zsh

# Output: (copy and paste into ~/.zshrc or ~/.bashrc)
# deadenv() {
#   if [[ "$1" == "use" ]]; then
#     export -p $(deadenv export "$2")
#   else
#     command deadenv "$@"
#   fi
# }

# After pasting and sourcing, you can use:
deadenv use myapp
# Now all vars are in your current shell (no subprocess)

# To run a command with profile active:
deadenv run myapp -- npm start
```

---

## Use Cases

### Solo Developer

**Problem:** You have `.env.local` files with credentials that could be accidentally committed.

**Solution:**
```bash
# One-time setup
deadenv profile new myapp --from=.env.local
rm .env.local .env.local.*.backup  # Remove plaintext copies

# Daily usage: run your app with secrets injected
deadenv run myapp -- npm start

# Or source into your shell for a persistent session
eval $(deadenv export myapp)
npm start
```

**Benefit:** No plaintext credentials on disk. No risk of accidental commits.

---

### Team Lead Sharing Secrets

**Problem:** A new team member joins and needs credentials. Sharing them via Slack or email is insecure.

**Solution:**

**Team Lead:**
```bash
# Export the profile
deadenv export myapp --out=myapp.deadenv

# Share the file via any channel (it's encrypted)
# Share the password via a separate, secure channel (Signal, 1Password, etc.)
```

**New Team Member:**
```bash
# Receives: myapp.deadenv file + password
deadenv import myapp.deadenv
# Enter sharing password: [provided via secure channel]
# ✓ All 12 keys imported
```

**Benefit:** Credentials never sent in plaintext. Easy one-time setup.

---

### CI/CD Integration

**Problem:** CI systems need credentials but shouldn't store them in plaintext.

**Solution:**
```bash
# In your CI pipeline:
deadenv run myapp-ci -- npm run build

# Secrets are injected into the build subprocess
# Exit code is propagated for CI failure detection
```

**Alternative:** Export to a file at deployment time (outside the repository):
```bash
deadenv export myapp-prod --format=json | \
  jq -r '.[] | "\(.key)=\(.value)"' > deploy-env
# Pass deploy-env to your container/lambda/etc.
```

---

### Multiple Environments

**Problem:** You work on dev, staging, and production configs with different secrets.

**Solution:**
```bash
# Create separate profiles
deadenv profile new myapp-dev --from=.env.dev.example
deadenv profile new myapp-staging --from=.env.staging.example
deadenv profile new myapp-prod --from=.env.prod.example

# Switch between environments
deadenv run myapp-dev -- npm start      # Dev database
deadenv run myapp-staging -- npm start  # Staging database
deadenv run myapp-prod -- npm start     # Production database (use with caution!)

# Edit a specific environment's secrets
deadenv edit myapp-staging
```

**Benefit:** Clear separation of environments. Audit trail shows which was changed when.

---

### Rotating a Leaked Secret

**Problem:** An API key is compromised and must be rotated.

**Solution:**
```bash
# Check current value (masked by default)
deadenv profile show myapp

# Update the compromised key
deadenv set myapp API_KEY "sk_live_new_secret_..."

# Audit history records both old and new (without values)
deadenv history myapp --key=API_KEY
# 2025-04-21 15:45  modified  API_KEY  (hash: new_hash...)
# 2025-04-21 10:00  set       API_KEY  (hash: old_hash...)
```

**Benefit:** Incident response is traceable. Values never logged or exposed.

---

## Architecture

### Package Organization

```
deadenv/
├── main.go                         # Entry point, error routing
├── cmd/                            # CLI command definitions (thin wrappers)
│   ├── profile.go                  # Profile subcommands
│   ├── set.go, get.go, unset.go    # Key management
│   ├── edit.go                     # Interactive profile editing
│   ├── run.go                      # Subprocess injection
│   ├── export.go, import.go        # Export/import logic
│   ├── history.go                  # Audit log display
│   └── init.go                     # Shell hook generation
├── internal/
│   ├── keychain/                   # OS keychain abstraction
│   │   ├── keychain.go             # Interface + types
│   │   ├── keychain_darwin.go      # macOS implementation
│   │   ├── keychain_linux.go       # Linux implementation
│   │   ├── keychain_windows.go     # Windows implementation
│   │   ├── fake.go                 # Test double
│   │   └── service.go              # Helper functions
│   ├── parser/                     # .env format parser (pure function)
│   │   ├── parser.go
│   │   └── fuzz_test.go
│   ├── crypto/                     # AES-256-GCM export encryption
│   │   ├── crypto.go
│   │   ├── types.go
│   │   └── errors.go
│   ├── history/                    # Git-backed audit log
│   │   ├── history.go              # Interface
│   │   ├── git_recorder.go         # Git implementation
│   │   ├── noop.go                 # No-op (for --no-history)
│   │   ├── fake_history.go         # Test double
│   │   └── hash.go                 # Value hashing
│   ├── profile/                    # Profile orchestration
│   │   ├── profile.go              # CRUD operations
│   │   ├── edit.go                 # Editor flow + diff logic
│   │   ├── populate.go             # Profile population from files
│   │   ├── serialize.go            # Format for editor display
│   │   ├── diff.go                 # Change detection
│   │   └── errors.go               # Sentinel errors
│   ├── tui/                        # Terminal UI helpers
│   │   ├── tui.go                  # Prompts, masked input, tables
│   │   └── colors.go               # ANSI styling
│   └── envPair/                    # Key-value pair type
│       └── envPair.go
├── testutil/                       # Shared test utilities
│   └── testutil.go
├── go.mod, go.sum                  # Dependencies
├── Makefile                        # Build targets
└── README.md
```

### Design Principles

**Interfaces Over Implementations**

- `keychain.Store` — abstracts platform-specific keychain access
- `history.Recorder` — abstracts git history recording
- Both are injected, making code testable with fakes

**Pure Functions**

- Parser takes a string, returns pairs or error — no I/O
- Crypto functions are deterministic — fully testable
- Parser fuzz tests validate robustness

**Thin CLI Layer**

- `cmd/` packages are wrappers around `internal/` logic
- No business logic in CLI handlers
- All errors flow through `main.go` for consistent exit codes

**Lenient Parsing**

- Unmatched quotes are treated as literals, not errors
- This matches `.env` tool behavior and improves compatibility

---

## Configuration

### Environment Variables

| Variable         | Purpose                   | Default             |
| ---------------- | ------------------------- | ------------------- |
| `DEADENV_CONFIG` | Override config directory | `~/.config/deadenv` |
| `EDITOR`         | Editor for `deadenv edit` | `$VISUAL` or `vi`   |
| `VISUAL`         | Alternative to `EDITOR`   | (not set)           |

### Config Directory

`~/.config/deadenv/` contains:

```
~/.config/deadenv/
├── history/                 # Git repository for audit log
│   └── <profile>.json       # History entries for each profile
└── .salt                    # Random salt for value hashing (created once)
```

### Global Flags

All `deadenv` commands support:

```bash
--config <dir>     # Override config directory
--no-history       # Skip git commit for this operation
--quiet            # Suppress informational output
--yes              # Skip confirmations (dangerous operations)
```

Examples:
```bash
# Use alternate config directory
deadenv --config=/data/deadenv profile list

# Skip history tracking for this operation
deadenv --no-history set myapp TEMP_KEY "value"

# Delete without prompt (be careful!)
deadenv profile rm myapp --yes
```

---

## Error Handling

### Exit Codes

| Code  | Meaning              | Example                                            |
| ----- | -------------------- | -------------------------------------------------- |
| `0`   | Success              | Any successful command                             |
| `1`   | General error        | Profile not found, bad args, validation error      |
| `2`   | Auth failure         | OS keychain denied access (user rejected Touch ID) |
| `3`   | Decrypt failure      | Wrong import password or corrupted file            |
| `4`   | Parse error          | Malformed env content, invalid key name            |
| `127` | Command not found    | Subprocess in `deadenv run` not found              |
| `N`   | Propagated exit code | Exit code from `deadenv run -- <command>`          |

### Error Messages

All errors:
- Go to `stderr` (normal output to `stdout`)
- Include helpful context
- **Never expose secret values**
- Suggest corrective action

Example:
```bash
$ deadenv get myapp NONEXISTENT
Error: key "NONEXISTENT" not found in profile "myapp"
Set it with: deadenv set myapp NONEXISTENT
```

---

## Testing

### Run All Tests

```bash
make test
```

This runs unit tests with race detection.

### Parser Tests

```bash
go test ./internal/parser -v
```

Includes table-driven tests for all `.env` format variants.

### Parser Fuzzing

```bash
make fuzz
```

Runs the parser against randomized input for 60 seconds to find edge cases.

### Integration Tests (Real Keychain)

```bash
make integration
```

Tests against real OS keychain and git. Requires:
- Keychain/Keyring/Credential Manager to be functional
- Git to be installed
- No prompt for authentication (in CI, use `--no-history`)

### Unit Tests with Coverage

```bash
go test -cover ./...
```

---

## Development Workflow

### Clone and Build

```bash
git clone https://github.com/funinkina/deadenv.git
cd deadenv
make build
./bin/deadenv --help
```

### Build for Specific Platform

```bash
# macOS (arm64/Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o bin/deadenv-darwin-arm64

# Linux (x86_64)
GOOS=linux GOARCH=amd64 go build -o bin/deadenv-linux-amd64

# Windows (x86_64)
GOOS=windows GOARCH=amd64 go build -o bin/deadenv-windows-amd64.exe
```

### Lint

```bash
make lint
```

Requires `golangci-lint`. Install with:
```bash
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh
```

### Add a New Command

1. Create `cmd/mycommand.go` with a `NewMyCommand()` function
2. Register in `cmd/root.go`
3. Add business logic to `internal/` packages
4. Write tests in `internal/<package>/*_test.go`
5. Update this README with examples

---

## Common Issues

### "git not found"

**Problem:** `deadenv history` displays a warning or history is disabled.

**Solution:** Install Git. Once installed, history will work automatically.

```bash
# macOS
brew install git

# Ubuntu/Debian
sudo apt-get install git

# Fedora
sudo dnf install git
```

---

### "Profile not found"

**Problem:** You get "profile myapp not found" when trying to access it.

**Solution:** List available profiles and create one if needed.

```bash
deadenv profile list
deadenv profile new myapp
```

---

### "Decryption failed"

**Problem:** Import fails with "decryption failed — wrong password or file is corrupted."

**Solution:**
1. Verify you're using the correct sharing password
2. Confirm the `.deadenv` file is intact (not truncated, not moved)
3. Try re-exporting from the source machine

---

### macOS: "denied" on first use

**Problem:** OS Keychain access is denied the first time you run `deadenv`.

**Solution:** This is normal. Grant permission by:
1. Clicking "Allow" in the OS prompt, or
2. Using `deadenv set` which will prompt you through Keychain setup

---

### Linux: "Secret service not available"

**Problem:** `deadenv` reports libsecret or KWallet is not available.

**Solution:** Install the keyring service:

```bash
# GNOME Keyring (Ubuntu/Debian/Fedora)
sudo apt-get install gnome-keyring

# Or KWallet (KDE)
sudo apt-get install kwalletmanager
```

Then restart your session or manually start the service:
```bash
dbus-daemon --session &
/usr/bin/gnome-keyring-daemon --start &
```

---

### Windows: Credential Manager integration

**Problem:** Credentials don't appear in Credential Manager GUI.

**Solution:** This is expected. `deadenv` stores items with a special prefix (`deadenv/<profile>`). They are accessible only through `deadenv` for security. To verify they're stored:

```bash
# List keys in a profile (masked)
deadenv profile show myapp

# Reveal values (requires Windows Hello or password prompt)
deadenv profile show myapp --reveal
```

---

## Security Considerations

### What's Protected

✅ Secrets are stored in OS-native encrypted storage  
✅ OS-native authentication gates all read access  
✅ Export files are encrypted with AES-256-GCM + Argon2id  
✅ History contains no plaintext values (only hashes)  
✅ Temp files used during editing are created with `0600` permissions  

### What's Not Protected

⚠️ **This tool is for development machines only.** Not for production infrastructure.  
⚠️ Secrets in process memory are accessible to system administrators.  
⚠️ If an attacker gains filesystem/system access, they can potentially extract secrets via OS APIs.  
⚠️ For production: use Vault, AWS Secrets Manager, GCP Secret Manager, etc.  

### Best Practices

1. **Never commit `.env` files or `.deadenv` exports to git**
2. **Share `.deadenv` files and passwords via separate channels** (file in email, password via Signal)
3. **Rotate secrets regularly**, especially if machines are shared
4. **Review audit history** with `deadenv history <profile>` to catch unauthorized changes
5. **Enable OS lock timeout** on your machine so keychain locks when you step away

---

## Roadmap (Future Phases)

- [ ] **Team sync:** Real-time secret sharing with team members (post-v1)
- [ ] **IDE plugins:** VS Code, JetBrains, etc. (post-v1)
- [ ] **GitHub Actions integration:** Automatic secret injection in CI workflows (post-v1)
- [ ] **Secret rotation:** Automatic or manual rotation with versioning (post-v2)
- [ ] **Access policies:** Fine-grained permissions for shared machines (post-v2)
- [ ] **Backup/restore:** Encrypted backups of all profiles (post-v2)

---

## Contributing

Contributions are welcome! To contribute:

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/your-feature`
3. Write tests for your changes (run `make test`)
4. Lint your code: `make lint`
5. Commit with clear messages
6. Push to your fork
7. Open a pull request with a description of your changes

### Development Environment Setup

```bash
# Install Go 1.26+
# Install golangci-lint for linting
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh

# Install dependencies
go mod download

# Build and test
make build test lint
```

### Reporting Issues

When reporting bugs, include:
- OS and version (macOS 14.1, Ubuntu 22.04, Windows 11, etc.)
- Go version: `go version`
- Error message and context
- Steps to reproduce
- Whether you're using a real keychain or running tests

Avoid sharing actual credentials in issue reports.

---

## License

MIT License — see `LICENSE` file for details.

---

## Support

- **Documentation:** See this README and `deadenv-prd.md` for detailed specs
- **Issues:** Report bugs on GitHub
- **Security issues:** Please disclose privately to the maintainers

---

## Acknowledgments

- Built with [urfave/cli](https://github.com/urfave/cli) for CLI framework
- Uses platform-specific keychain libraries for secure storage
- Inspired by best practices in secret management tools like Vault and 1Password

---

## See Also

- [PRD (Product Requirements Document)](deadenv-prd.md)
- [Architecture Documentation](docs/architecture.md) *(forthcoming)*
- [API Reference](docs/api.md) *(forthcoming)*

---

**Made with ❤️ for developers who value security.**
