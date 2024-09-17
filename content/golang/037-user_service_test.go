func (s *ServiceTestSuite) TestRegisterUserWithMissingAttributes() {
	examples := []services.RegisterUserParams{
		{Email: "", DisplayName: "Example User", Password: "foobar123123", PasswordConfirmation: "foobar123123"},
		{Email: "registration@example.com", DisplayName: "", Password: "foobar123123", PasswordConfirmation: "foobar123123"},
		{Email: "registration@example.com", DisplayName: "Example User", Password: "", PasswordConfirmation: "foobar123123"},
	}

	srv := services.NewUserService(s.db)

	for _, params := range examples {
		user, err := srv.RegisterUser(context.Background(), params)
		s.Nil(user)
		s.IsType(validate.Errors{}, err)

		actual := err.(validate.Errors).OneError().Error()
		s.Equal("can't be blank", actual)
	}
}
