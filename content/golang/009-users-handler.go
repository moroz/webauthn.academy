package handler

import (
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/moroz/webauthn-academy-go/service"
	"github.com/moroz/webauthn-academy-go/templates/users"
	"github.com/moroz/webauthn-academy-go/types"
)

type userHandler struct {
	us service.UserService
}

func UserHandler(db *sqlx.DB) userHandler {
	return userHandler{service.NewUserService(db)}
}

func (h *userHandler) New(w http.ResponseWriter, r *http.Request) {
	err := users.New(types.NewUserParams{}, nil).Render(r.Context(), w)
	if err != nil {
		log.Printf("Rendering error: %s", err)
	}
}
