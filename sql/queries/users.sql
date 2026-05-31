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
    hashed_password,
    is_chirpy_red
FROM users
WHERE email = $1;

-- name: UpdateUser :one
UPDATE users
SET email = $1, hashed_password = $2, updated_at = now()
WHERE id = $3
RETURNING *;

-- name: MakeUserChirpyRed :one
UPDATE users
SET is_chirpy_red = TRUE, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteUsers :exec
DELETE FROM users;
