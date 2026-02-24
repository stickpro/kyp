-- name: GetByName :one
SELECT * FROM vault_meta WHERE name=? LIMIT 1;

-- name: GetAll :one
SELECT * FROM vault_meta;

