package service

import (
	"github.com/alexedwards/argon2id"
	"github.com/gookit/validate"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/moroz/webauthn-academy-go/store"
	"github.com/moroz/webauthn-academy-go/types"
)

func init() {
	validate.Config(func(opt *validate.GlobalOption) {
		opt.StopOnError = false
	})
}

type UserService struct {
	store store.UserStore
}

func NewUserService(db *sqlx.DB) UserService {
	return UserService{store.NewUserStore(db)}
}

func (s *UserService) RegisterUser(params types.NewUserParams) (*types.User, error, validate.Errors) {
	v := validate.Struct(params)

	if !v.Validate() {
		return nil, nil, v.Errors
	}

	passwordHash, err := argon2id.CreateHash(params.Password, argon2id.DefaultParams)
	if err != nil {
		return nil, err, nil
	}

	user, err := s.store.InsertUser(&types.User{
		Email:        params.Email,
		PasswordHash: passwordHash,
		DisplayName:  params.DisplayName,
	})

	if err == nil {
		return user, nil, nil
	}

	// https://www.postgresql.org/docs/current/errcodes-appendix.html
	// Error 23505 `unique_violation` means that a unique constraint has
	// prevented us from inserting a duplicate value. Instead of returning
	// a raw error, we return a handcrafted validation error that we can
	// later display in a form.
	if err, ok := err.(*pq.Error); ok && err.Code == "23505" && err.Constraint == "users_email_key" {
		validationErrors := validate.Errors{}
		validationErrors.Add("Email", "unique", "has already been taken")
		return nil, nil, validationErrors
	}

	return nil, err, nil
}
