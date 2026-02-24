-- name: Create :one
INSERT INTO vault_meta (id, name, salt, verifier, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?)
	RETURNING *;

-- name: Delete :exec
DELETE FROM vault_meta WHERE id=?;

-- name: Get :one
SELECT * FROM vault_meta WHERE id=? LIMIT 1;

-- name: Update :one
UPDATE vault_meta
	SET name=?, salt=?, verifier=?
	WHERE id=?
	RETURNING *;

