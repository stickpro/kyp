-- name: GetByName :one
SELECT * FROM vault_meta WHERE name=? LIMIT 1;

-- name: GetAll :many
SELECT * FROM vault_meta;

