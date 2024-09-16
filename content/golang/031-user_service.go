func (us *UserService) RegisterUser(ctx context.Context, params RegisterUserParams) (*queries.User, error) {
	user, err := us.queries.InsertUser(ctx, queries.InsertUserParams{
		Email:        params.Email,
		DisplayName:  params.DisplayName,
		PasswordHash: params.Password,
	})

	return user, err
}
