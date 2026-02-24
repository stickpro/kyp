CREATE TABLE vault_meta (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    salt        BLOB NOT NULL,
    verifier    BLOB NOT NULL,
    created_at  INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_at  INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE TABLE entries (
    id              TEXT PRIMARY KEY,
    title           TEXT    NOT NULL,
    username        BLOB,
    password        BLOB,
    url             BLOB,
    notes           BLOB,

    -- TOTP
    totp_secret     BLOB,
    totp_issuer     TEXT,
    totp_algorithm  TEXT    NOT NULL DEFAULT 'SHA1',
    totp_digits     INTEGER NOT NULL DEFAULT 6,
    totp_period     INTEGER NOT NULL DEFAULT 30,

    created_at      INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_at      INTEGER NOT NULL DEFAULT (unixepoch()),
    deleted_at      INTEGER
);

CREATE INDEX idx_entries_deleted_at ON entries (deleted_at);
CREATE INDEX idx_entries_updated_at ON entries (updated_at);

CREATE TABLE sync_state (
    device_id       TEXT    PRIMARY KEY,
    server_url      TEXT,
    last_synced_at  INTEGER
);
