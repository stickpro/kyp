# kyp

A local-first password manager with TOTP support. Stores everything in an encrypted SQLite database on your machine. A sync server (`kypd`) is planned for optional cross-device synchronization.

## How it works

The vault is a single SQLite file. All sensitive fields (username, password, URL, notes, TOTP secret) are encrypted with AES-256-GCM before being written to disk. The encryption key is never stored — it is derived from your master password at runtime using Argon2id and discarded when the application exits.

The master password is verified through a small encrypted token stored alongside the vault metadata. If decryption of that token succeeds, the password is correct and the derived key is kept in memory for the session.

## Project structure

```
cmd/
  kyp/     TUI client
  kypd/    sync server (in progress)

internal/
  crypto/  key derivation, encryption, password verification
  vault/   vault lifecycle and entry management
  storage/ SQLite layer with generated queries

sql/
  sqlite/
    migrations/  goose migration files
    queries/     sqlc query definitions
```

## Security

- Master password is never stored anywhere
- Key derivation: Argon2id (time=1, memory=64MB, threads=4, key=32 bytes)
- Encryption: AES-256-GCM with a random nonce per operation
- Soft deletes (`deleted_at`) preserve entry history for future sync conflict resolution
- `Close()` zeroes the key in memory before releasing it

## Requirements

- Go 1.25+
- [pgxgen](https://github.com/stickpro/pgxgen) and [sqlc](https://sqlc.dev) for code generation (development only)

## Building

```bash
make build        # build kyp (TUI client)
make build-server # build kypd (sync server)
make build-all    # build both
```

## Development

```bash
make gen-sql  # regenerate repository code from SQL queries
make fmt      # format code with gofumpt
make lint     # run golangci-lint
```

## Configuration

The application looks for `config.yaml` in the working directory by default. The config file is optional — all settings can be provided via environment variables or flags.

```bash
kyp start --configs config.yaml,config.local.yaml
```

## Status

The project is under active development. The core crypto and vault layers are complete. The TUI and sync server are next.
