package main

import (
	"context"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/moroz/webauthn-academy-go/config"
	"github.com/moroz/webauthn-academy-go/handlers"
)

func main() {
	db, err := pgxpool.New(context.Background(), config.DATABASE_URL)
	if err != nil {
		log.Fatal(err)
	}

	router := handlers.NewRouter(db)

	log.Println("Listening on port 3000")
	log.Fatal(http.ListenAndServe(":3000", router))
}
