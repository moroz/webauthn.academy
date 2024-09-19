func (s *ServiceTestSuite) TestRegisterUserWithDuplicateEmail() {
	params := services.RegisterUserParams{
		Email: "uniqueness@example.com", DisplayName: "Uniqueness Test", Password: "foobar123123", PasswordConfirmation: "foobar123123",
	}

	srv := services.NewUserService(s.db)

	user, err := srv.RegisterUser(context.Background(), params)
	s.NotNil(user)
	s.Nil(err)

	emails := []string{params.Email, strings.ToUpper(params.Email), "Uniqueness@Example.Com"}

	for _, email := range emails {
		params.Email = email
		user, err := srv.RegisterUser(context.Background(), params)
		s.Nil(user)
		s.IsType(validate.Errors{}, err)

		actual := err.(validate.Errors).Field("Email")
		s.Equal(map[string]string{"unique": "has already been taken"}, actual)
	}
}
