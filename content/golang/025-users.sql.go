// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: users.sql

package queries

import (
	"context"
)

const insertUser = `-- name: InsertUser :one
insert into users (email, display_name, password_hash) values ($1, $2, $3) returning id, email, display_name, password_hash, inserted_at, updated_at
`

type InsertUserParams struct {
	Email        string
	DisplayName  string
	PasswordHash string
}

func (q *Queries) InsertUser(ctx context.Context, arg InsertUserParams) (*User, error) {
	row := q.db.QueryRow(ctx, insertUser, arg.Email, arg.DisplayName, arg.PasswordHash)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.DisplayName,
		&i.PasswordHash,
		&i.InsertedAt,
		&i.UpdatedAt,
	)
	return &i, err
}

