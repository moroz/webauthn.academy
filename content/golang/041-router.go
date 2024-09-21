package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/moroz/webauthn-academy-go/db/queries"
)

func NewRouter(db queries.DBTX) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	userRegistrations := UserRegistrationController(db)
	r.Get("/", userRegistrations.New)

	return r
}
