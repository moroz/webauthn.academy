package handler

import (
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/moroz/webauthn-academy-go/service"
	"github.com/moroz/webauthn-academy-go/templates/sessions"
)

type sessionHandler struct {
	us service.UserService
}

func SessionHandler(db *sqlx.DB) sessionHandler {
	return sessionHandler{service.NewUserService(db)}
}

func (h *sessionHandler) New(w http.ResponseWriter, r *http.Request) {
	err := sessions.New().Render(r.Context(), w)
	if err != nil {
		log.Print(err)
	}
}
