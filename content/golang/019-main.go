// ...

func main() {
	db := sqlx.MustConnect("postgres", config.DatabaseURL)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	users := handler.UserHandler(db)
	r.Get("/", users.New)
	r.Post("/users/register", users.Create)

	// add these lines
	sessions := handler.SessionHandler(db)
	r.Get("/sign-in", sessions.New)

	log.Println("Listening on port 3000")
	log.Fatal(http.ListenAndServe(":3000", r))
}
