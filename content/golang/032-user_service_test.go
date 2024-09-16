func (s *ServiceTestSuite) TestRegisterUser() {
	params := services.RegisterUserParams{
		Email:                "registration@example.com",
		DisplayName:          "Example User",
		Password:             "foobar123123",
		PasswordConfirmation: "foobar123123",
	}

	srv := services.NewUserService(s.db)
	user, err := srv.RegisterUser(context.Background(), params)
	s.NoError(err)
	s.NotNil(user)

	s.Regexp(regexp.MustCompile(`^\$argon2id\$`), user.PasswordHash)
}
