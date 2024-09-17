type RegisterUserParams struct {
	Email                string `validate:"required|email"`
	DisplayName          string `validate:"required"`
	Password             string `validate:"required|min_len:8|max_len:64"`
	PasswordConfirmation string `validate:"eq_field:Password" message:"passwords do not match"`
}

func init() {
	validate.AddGlobalMessages(map[string]string{
		"required": "can't be blank",
		"min_len":  "must be at least %d characters long",
		"max_len":  "must be at most %d characters long",
	})
}

