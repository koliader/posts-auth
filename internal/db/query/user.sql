-- name: CreateUser :one
INSERT INTO users (
  email,
  username,
  password
) VALUES (
  $1, $2, $3
) RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: ListUsers :many
SELECT * FROM users;