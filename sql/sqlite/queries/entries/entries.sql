-- name: GetWithPaginate :many
SELECT * FROM entries WHERE deleted_at IS NULL ORDER BY name DESC LIMIT ? OFFSET ?;
