func init() {
	validate.AddGlobalMessages(map[string]string{
		"required":                      "can't be blank",
		"min_len":                       "must be at least %d characters long",
		"max_len":                       "must be at most %d characters long",
		"PasswordConfirmation.eq_field": "passwords do not match",
	})
}
