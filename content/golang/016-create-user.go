package handler

import (
	"log"
	"net/http"

	"github.com/moroz/webauthn-academy-go/templates/users"
	"github.com/moroz/webauthn-academy-go/types"
)

// ...

func (h *userHandler) Create(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() // Ignore the error as it will be handled later

	var params types.NewUserParams
	if err := decoder.Decode(&params, r.PostForm); err != nil {
		handleError(w, http.StatusBadRequest)
		return
	}

	_, err, validationErrors := h.us.RegisterUser(params)

	if err != nil || validationErrors != nil {
		w.Header().Add("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusUnprocessableEntity)
		err := users.New(params, validationErrors).Render(r.Context(), w)
		if err != nil {
			log.Print(err)
		}
		return
	}

	http.Redirect(w, r, "/sign-in", http.StatusMovedPermanently)
}
