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
WHERE ($1::uuid = '00000000-0000-0000-0000-000000000000'::uuid OR user_id = $1)
ORDER BY created_at;

-- name: DeleteChirp :exec
DELETE FROM chirps
WHERE id = $1;
