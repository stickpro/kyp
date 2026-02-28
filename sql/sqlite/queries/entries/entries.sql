-- name: GetWithPaginate :many
SELECT * FROM entries WHERE deleted_at IS NULL ORDER BY title ASC LIMIT ? OFFSET ?;
