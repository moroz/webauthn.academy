func (us *UserService) RegisterUser(ctx context.Context, params RegisterUserParams) (*queries.User, error) {
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
