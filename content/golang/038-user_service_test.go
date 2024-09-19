package services_test

import (
	"context"
	"regexp"

	"github.com/gookit/validate"
	"github.com/moroz/webauthn-academy-go/services"
)

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

func (s *ServiceTestSuite) TestRegisterUserWithInvalidEmail() {
	params := services.RegisterUserParams{
		Email: "user@invalid", DisplayName: "Invalid User", Password: "foobar123123", PasswordConfirmation: "foobar123123",
	}

	srv := services.NewUserService(s.db)

	user, err := srv.RegisterUser(context.Background(), params)
	s.Nil(user)
	s.IsType(validate.Errors{}, err)

	actual := err.(validate.Errors).Field("Email")
	s.Equal(map[string]string{"email": "is not a valid email address"}, actual)
}

func (s *ServiceTestSuite) TestRegisterUserWithShortPassword() {
	params := services.RegisterUserParams{
		Email: "user@example.com", DisplayName: "Invalid User", Password: "foo", PasswordConfirmation: "foo",
	}

	srv := services.NewUserService(s.db)

	user, err := srv.RegisterUser(context.Background(), params)
	s.Nil(user)
	s.IsType(validate.Errors{}, err)

	actual := err.(validate.Errors).Field("Password")
	s.Equal(map[string]string{"min_len": "must be at least 8 characters long"}, actual)
}

func (s *ServiceTestSuite) TestRegisterUserWithTooLongPassword() {
	password := ""
	for _ = range 80 {
		password += "a"
	}

	params := services.RegisterUserParams{
		Email: "user@example.com", DisplayName: "Invalid User", Password: password, PasswordConfirmation: password,
	}

	srv := services.NewUserService(s.db)

	user, err := srv.RegisterUser(context.Background(), params)
	s.Nil(user)
	s.IsType(validate.Errors{}, err)

	actual := err.(validate.Errors).Field("Password")
	s.Equal(map[string]string{"max_len": "must be at most 64 characters long"}, actual)
}

func (s *ServiceTestSuite) TestRegisterUserWithInvalidPasswordConfirmation() {
	params := services.RegisterUserParams{
		Email: "user@example.com", DisplayName: "Invalid User", Password: "foobar123123", PasswordConfirmation: "different_password",
	}

	srv := services.NewUserService(s.db)

	user, err := srv.RegisterUser(context.Background(), params)
	s.Nil(user)
	s.IsType(validate.Errors{}, err)

	actual := err.(validate.Errors).Field("PasswordConfirmation")
	s.Equal(map[string]string{"eq_field": "passwords do not match"}, actual)
}
