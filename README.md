![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)
![License](https://img.shields.io/github/license/stickpro/kyp)
![Build](https://img.shields.io/github/actions/workflow/status/stickpro/kyp/ci.yml)

# kyp — Keep Your Passwords

> No cloud. No account. No trust required.

A terminal password manager that stores everything in an **encrypted SQLite file on your machine**. Your vault never touches a server.


[![asciicast](https://asciinema.org/a/os0yIFyfkI7z7Brk.svg)](https://asciinema.org/a/os0yIFyfkI7z7Brk)

## Why kyp

| | kyp | pass | gopass | Bitwarden |
|--|-----|------|--------|--------------|
| Encrypted vault | AES-256-GCM | GPG | GPG | AES-256 |
| TOTP built-in | ✅ | ❌ | ✅ | ❌ |
| No cloud required | ✅ | ✅ | ✅ | ❌ |
| Single binary | ✅ | ❌ | ❌ | ✅ |
| No GPG setup | ✅ | ❌ | ❌ | ✅ |
| TUI interface | ✅ | ❌ | ❌ | ❌ |

## Features

- **Fully local** - vault is a single SQLite file, no cloud required
- **AES-256-GCM encryption** - every sensitive field (username, password, URL, notes, TOTP secret) is encrypted individually before being written to disk
- **Argon2id key derivation** - master password is never stored; the key is derived at runtime and zeroed from memory on exit
- **TOTP support** - store TOTP secrets, view live codes with countdown timer, copy to clipboard with one key
- **Clipboard integration** - copy login, password or TOTP code without revealing it on screen
- **Password visibility toggle** - show/hide password in the detail view
- **Bitwarden CSV import** - import your existing vault with optional folder filter
- **Fuzzy search** - built-in filtering across all entries
- **Tab navigation** - keyboard-only, no mouse required
- **Adaptive colors** - UI works correctly on both light and dark terminals

> **Sync server (`kypd`) and GUI client are under development.**

## Install

**macOS — Homebrew**
```bash
brew tap stickpro/kyp
brew install kyp
```

**Debian / Ubuntu**
```bash
# Download the .deb from the latest release, then:
sudo dpkg -i kyp_*_linux_amd64.deb
```

**RHEL / Fedora / CentOS**
```bash
sudo rpm -i kyp_*_linux_amd64.rpm
```

**Alpine**
```bash
apk add --allow-untrusted kyp_*_linux_amd64.apk
```

**Windows — Scoop**
```powershell
scoop bucket add kyp https://github.com/stickpro/scoop-kyp
scoop install kyp
```

**Go**
```bash
go install github.com/stickpro/kyp/cmd/kyp@latest
```

**Manual** — download the archive for your OS/arch from the [releases page](https://github.com/stickpro/kyp/releases), extract, and put `kyp` in your `$PATH`.

## How it works

The vault is a single SQLite file. All sensitive fields are encrypted with AES-256-GCM before being written to disk. The encryption key is never stored - it is derived from your master password at runtime using Argon2id and discarded when the application exits.

The master password is verified through a small encrypted token stored alongside the vault metadata. If decryption of that token succeeds, the password is correct and the derived key is kept in memory for the session.

## Keyboard shortcuts

| Key | Action |
|-----|--------|
| `tab` / `shift+tab` | Next / previous field |
| `enter` | Confirm / select |
| `esc` | Back |
| `n` | New entry (from list) |
| `e` | Edit entry (from detail) |
| `u` | Copy username |
| `c` | Copy password |
| `t` | Copy TOTP code |
| `p` / `space` | Show / hide password |
| `q` / `ctrl+c` | Quit |

## Import from Bitwarden

```bash
# Build the import tool
make build-import

# Import all entries
./.bin/kyp-import bitwarden_export.csv kyp.db mypassword

# Import only entries from a specific folder
./.bin/kyp-import bitwarden_export.csv kyp.db mypassword "Work"
```

## Project structure

```
cmd/
  kyp/     TUI client
  kypd/    sync server (in development)
  import/  Bitwarden CSV import tool

internal/
  crypto/  key derivation, AES-256-GCM, password generator
  totp/    RFC 6238 TOTP code generation
  vault/   vault lifecycle and entry CRUD
  storage/ SQLite layer with generated queries
  tui/     Bubbletea UI screens (list, detail, form, unlock, create)

sql/
  sqlite/
    migrations/  goose migration files
    queries/     sqlc query definitions
```

## Security

- Master password is never stored anywhere
- Key derivation: Argon2id (time=1, memory=64 MB, threads=4, key=32 bytes)
- Encryption: AES-256-GCM with a random nonce per field per write
- Soft deletes (`deleted_at`) preserve entry history for future sync conflict resolution
- `vault.Close()` zeroes the master key in memory before releasing it

## Building

```bash
make build         # TUI client  →  .bin/kyp
make build-server  # sync server →  .bin/kypd
make build-import  # import tool →  .bin/kyp-import
make build-all     # all three
```

## Running

```bash
make run start
```

## Development

```bash
make gen-sql  # regenerate repository code from SQL queries
make fmt      # format code with gofumpt
make lint     # run golangci-lint
```

## Requirements

- Go 1.22+
- [pgxgen](https://github.com/stickpro/pgxgen) and [sqlc](https://sqlc.dev) for code generation (development only)

## Roadmap

- [x] Encrypted SQLite vault
- [x] Argon2id key derivation
- [x] AES-256-GCM per-field encryption
- [x] TOTP code generation (RFC 6238)
- [x] Terminal UI (list / detail / form / unlock / create)
- [x] Clipboard integration
- [x] Bitwarden CSV import
- [ ] Sync server (`kypd`)
- [ ] GUI client
- [ ] Browser extension
