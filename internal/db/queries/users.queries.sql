-- name: CreateUser :one
INSERT INTO users (username, name, email, password, bio,location) 
VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserById :one
SELECT * FROM users WHERE id = $1;

-- name: UpdateUser :one
UPDATE users 
SET 
    username = COALESCE(sqlc.narg('username'), username),
    name = COALESCE(sqlc.narg('name'), name),
    email = COALESCE(sqlc.narg('email'), email),
    password = COALESCE(sqlc.narg('password'), password),
    bio = COALESCE(sqlc.narg('bio'), bio),
    location = COALESCE(sqlc.narg('location'), location),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users;
