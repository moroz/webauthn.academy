func (s *ServiceTestSuite) TestRegisterUserWithMissingAttributes() {
	examples := []services.RegisterUserParams{
		{Email: "", DisplayName: "Example User", Password: "foobar123123", PasswordConfirmation: "foobar123123"},
		{Email: "registration@example.com", DisplayName: "", Password: "foobar123123", PasswordConfirmation: "foobar123123"},
		{Email: "registration@example.com", DisplayName: "Example User", Password: "", PasswordConfirmation: ""},
	}

	srv := services.NewUserService(s.db)

	for _, params := range examples {
		user, err := srv.RegisterUser(context.Background(), params)
		s.Error(err)
		s.Nil(user)
	}
}
