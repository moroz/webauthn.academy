package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/moroz/webauthn-academy-go/types"
)

type UserStore struct {
	db *sqlx.DB
}

func NewUserStore(db *sqlx.DB) UserStore {
	return UserStore{db}
}

const insertUserQuery = `insert into users (email, display_name, password_hash) values ($1, $2, $3) returning *`

func (s *UserStore) InsertUser(user *types.User) (*types.User, error) {
	var result types.User
	err := s.db.Get(&result, insertUserQuery, user.Email, user.DisplayName, user.PasswordHash)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
