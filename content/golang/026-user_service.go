package services

import (
	"github.com/moroz/webauthn-academy-go/db/queries"

	"context"
)

type UserService struct {
	queries *queries.Queries
}

func NewUserService(db queries.DBTX) UserService {
	return UserService{queries: queries.New(db)}
}

type RegisterUserParams struct {
	Email                string
	DisplayName          string
	Password             string
	PasswordConfirmation string
}

func (us *UserService) RegisterUser(ctx context.Context, params RegisterUserParams) (*queries.User, error) {
	return nil, nil
}
