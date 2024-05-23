package handler

import (
	"net/http"

	"github.com/gorilla/schema"
)

var decoder = schema.NewDecoder()

func handleError(w http.ResponseWriter, status int) {
	w.Header().Add("content-type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	msg := http.StatusText(status)
	w.Write([]byte(msg))
}
