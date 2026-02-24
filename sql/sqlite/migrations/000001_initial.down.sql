DROP INDEX IF EXISTS idx_entries_updated_at;
DROP INDEX IF EXISTS idx_entries_deleted_at;

DROP TABLE IF EXISTS sync_state;
DROP TABLE IF EXISTS entries;
DROP TABLE IF EXISTS vault_meta;