-- name: CreateUser :one
INSERT INTO users (email, hashed_password)
VALUES ($1, $2)
RETURNING *;

-- name: GetUser :one
SELECT
    id,
    created_at,
    updated_at,
    email,
    hashed_password
FROM users
WHERE email = $1;

-- name: DeleteUsers :exec
DELETE FROM users;
