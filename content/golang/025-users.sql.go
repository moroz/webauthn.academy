package queries

import (
	"context"
)

const getUserByEmail = `-- name: GetUserByEmail :one
select id, email, display_name, password_hash, inserted_at, updated_at from users where email = $1
`

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	row := q.db.QueryRow(ctx, getUserByEmail, email)
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

// ...more code below...
