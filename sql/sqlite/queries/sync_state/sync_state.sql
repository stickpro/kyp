-- name: Get :one
SELECT * FROM sync_state LIMIT 1;

-- name: Upsert :exec
INSERT INTO sync_state (device_id, server_url, last_synced_at)
    VALUES (?, ?, ?)
    ON CONFLICT (device_id) DO UPDATE SET
        server_url     = excluded.server_url,
        last_synced_at = excluded.last_synced_at;

-- name: UpdateServerURL :exec
UPDATE sync_state SET server_url = ? WHERE device_id = ?;

-- name: UpdateLastSyncedAt :exec
UPDATE sync_state SET last_synced_at = ? WHERE device_id = ?;
