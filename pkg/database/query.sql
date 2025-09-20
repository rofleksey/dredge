-- name: CreateMessage :exec
INSERT INTO messages (id, created, channel, username, text)
VALUES ($1, $2, $3, $4, $5);

-- name: GetMigrations :many
SELECT *
FROM migration
ORDER BY id;

-- name: CreateMigration :one
INSERT INTO migration (id, applied)
VALUES ($1, $2) RETURNING id;
