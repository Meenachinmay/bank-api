-- name: CreateUser :one
INSERT INTO users (user_name, email, created_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
    WHERE email = $1
LIMIT 1;

-- name: GetUserById :one
SELECT * FROM users
WHERE id = $1
LIMIT 1;

-- name: UpdateUserInterest :one
UPDATE users
SET extra_interest = $2
WHERE id = $1
RETURNING *;