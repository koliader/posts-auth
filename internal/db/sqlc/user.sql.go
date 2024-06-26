// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: user.sql

package db

import (
	"context"
)

const createUser = `-- name: CreateUser :one
INSERT INTO users (
  email,
  username,
  password
) VALUES (
  $1, $2, $3
) RETURNING email, username, password
`

type CreateUserParams struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRow(ctx, createUser, arg.Email, arg.Username, arg.Password)
	var i User
	err := row.Scan(&i.Email, &i.Username, &i.Password)
	return i, err
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT email, username, password FROM users
WHERE email = $1
`

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRow(ctx, getUserByEmail, email)
	var i User
	err := row.Scan(&i.Email, &i.Username, &i.Password)
	return i, err
}

const listUsers = `-- name: ListUsers :many
SELECT email, username, password FROM users
`

func (q *Queries) ListUsers(ctx context.Context) ([]User, error) {
	rows, err := q.db.Query(ctx, listUsers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []User{}
	for rows.Next() {
		var i User
		if err := rows.Scan(&i.Email, &i.Username, &i.Password); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateUserEmail = `-- name: UpdateUserEmail :one
UPDATE users
SET email = $2
WHERE email = $1
RETURNING email, username, password
`

type UpdateUserEmailParams struct {
	Email   string `json:"email"`
	Email_2 string `json:"email_2"`
}

func (q *Queries) UpdateUserEmail(ctx context.Context, arg UpdateUserEmailParams) (User, error) {
	row := q.db.QueryRow(ctx, updateUserEmail, arg.Email, arg.Email_2)
	var i User
	err := row.Scan(&i.Email, &i.Username, &i.Password)
	return i, err
}
