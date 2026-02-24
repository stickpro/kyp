-- name: Create :exec
INSERT INTO entries (id, title, username, password, url, notes, totp_secret, totp_issuer, totp_algorithm, totp_digits, totp_period, created_at, updated_at, deleted_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: Delete :exec
DELETE FROM entries WHERE id=?;

-- name: Get :one
SELECT * FROM entries WHERE id=? LIMIT 1;

-- name: Update :one
UPDATE entries
	SET title=?, username=?, password=?, url=?, notes=?, totp_secret=?, 
		totp_issuer=?, totp_algorithm=?, totp_digits=?, totp_period=?, deleted_at=?
	WHERE id=?
	RETURNING *;

