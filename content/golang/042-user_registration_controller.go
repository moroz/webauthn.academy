package handlers

import (
	"fmt"
	"net/http"

	"github.com/moroz/webauthn-academy-go/db/queries"
)

type userRegistrationController struct {
	queries *queries.Queries
}

func UserRegistrationController(db queries.DBTX) userRegistrationController {
	return userRegistrationController{queries.New(db)}
}

func (uc *userRegistrationController) New(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Hello from UserRegistrationController!</h1>")
}
