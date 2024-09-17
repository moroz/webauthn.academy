type RegisterUserParams struct {
	Email                string `validate:"required|email"`
	DisplayName          string `validate:"required"`
	Password             string `validate:"required|min_len:8|max_len:64"`
	PasswordConfirmation string `validate:"eq_field:Password"`
}

func (us *UserService) RegisterUser(ctx context.Context, params RegisterUserParams) (*queries.User, error) {
	if v := validate.Struct(&params); !v.Validate() {
		return nil, v.Errors
	}

	hash, err := argon2id.CreateHash(params.Password, argon2id.DefaultParams)
	if err != nil {
		return nil, err
	}

	user, err := us.queries.InsertUser(ctx, queries.InsertUserParams{
		Email:        params.Email,
		DisplayName:  params.DisplayName,
		PasswordHash: hash,
	})

	return user, err
}
