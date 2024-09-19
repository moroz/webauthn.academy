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

	// intercept "unique_violation" errors on the email column
	if err, ok := err.(*pgconn.PgError); ok && err.Code == "23505" && err.ConstraintName == "users_email_key" {
		return nil, validate.Errors{"Email": {"unique": "has already been taken"}}
	}

	return user, err
}
