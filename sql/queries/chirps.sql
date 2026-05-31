-- name: CreateChirp :one
INSERT INTO chirps (body, user_id)
VALUES ($1, $2)
RETURNING *;

-- name: GetChirp :one
SELECT
    id,
    created_at,
    updated_at,
    body,
    user_id
FROM chirps
WHERE id = $1;

-- name: GetChirps :many
SELECT
    id,
    created_at,
    updated_at,
    body,
    user_id
FROM chirps
WHERE ($1::uuid IS NULL OR user_id = $1)
ORDER BY created_at;

-- name: DeleteChirp :exec
DELETE FROM chirps
WHERE id = $1;
