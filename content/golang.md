---
title: Implementing Webauthn in Golang
---

This section is dedicated to a framework-agnostic implementation of a Webauthn workflow using the Go programming language.

## Who this text is for

This website is not meant as a complete learning resource for beginners, but rather a more or less comprehensive overview of Webauthn that I wished were available when I was learning about this technology.

This text assumes that you are an experienced Web developer, with reasonably good knowledge of back end development, the UNIX command line, SQL, and all three languages used in browser environments (HTML, CSS, and JavaScript).
Therefore I won't be stopping to explain code snippets that I believe should be readable without explanation.

If you have any suggestions for improvements to the tutorial, feel free to [reach out to me](https://github.com/moroz) or to submit a Pull Request or an issue in the [Github repository](https://github.com/moroz/webauthn.academy) of this website.

The source code of the application we are going to build (work in progress) is available on Github: [moroz/webauthn-academy-go](https://github.com/moroz/webauthn-academy-go).

## Technological stack

Whenever possible, I try to use just the standard library, so with enough knowledge of the Go ecosystem, you should be able to modify the solution to use your preferred libraries.

A few command-line tools we will be using in this walkthrough:

* [mise](https://mise.jdx.dev/) --- to manage different versions of programming languages, here Go and Node.js.
* [goose](https://github.com/pressly/goose) --- to generate and run database migrations,
* [direnv](https://direnv.net/) --- to manage settings and secrets in environment variables.
* [modd](https://github.com/cortesi/modd) --- to automatically rebuild and reload the application.

This website was developed and written on Debian 12, using Go 1.22.3 and Node 20.13.1, the latest LTS release as of this writing.
For persistence, I will be using PostgreSQL 16.2, but any reasonably modern version of PostgreSQL should work too.

A few notable Go libraries we will be using in the application:

* [github.com/alexedwards/argon2id](https://pkg.go.dev/github.com/alexedwards/argon2id) --- to hash passwords using the Argon2id password hashing algorithm.
* [github.com/jmoiron/sqlx](https://github.com/jmoiron/sqlx) --- a thin wrapper over `database/sql` that also allows us to read query results into structs. If you know enough SQL, you should be able to replace it with another solution, such as [sqlc](https://sqlc.dev/) or [gorm](https://gorm.io/) (which is not as fantastic as it claims to be).
* [github.com/go-webauthn/webauthn](https://github.com/go-webauthn/webauthn) --- the actual WebAuthn implementation. We will be using this library to generate and validate registration and attestation challenges.
* [templ](https://templ.guide/) --- a type-safe templating language that compiles to Go.
* [github.com/gorilla/schema](https://github.com/gorilla/schema) --- to parse URL-encoded data into structs.
* [github.com/gorilla/sessions](https://pkg.go.dev/github.com/gorilla/sessions) --- for persisting session state in cookies. We will be using session storage to display flash notifications, for CSRF protection, and to persist WebAuthn challenges across requests.
* [github.com/gookit/validate](https://github.com/gookit/validate/) --- for struct validation.

We will be compiling and bundling CSS and JavaScript using [Vite](https://vitejs.dev/), some [TypeScript](https://www.typescriptlang.org/), and [SASS](https://sass-lang.com/).

## Initial setup

The following walkthrough sets up a password authentication from scratch. Once this text is finalized, you will be able to skip to the section where I start implementing Webauthn. For now, you can just follow along.


Create a directory for the new project:

```plain
mkdir academy-go
```

Ensure Golang is installed (here using [mise](https://mise.jdx.dev/)):

```plain
cd academy-go
mise install go@1.22.3 node@lts
mise local go@1.22.3
mise local node@lts
```

Initialize a Go module in this directory.

```shell
# Swap "moroz" for your Github username
go mod init github.com/moroz/webauthn-academy-go
```

Initialize a Git repository in this directory:

```shell
git init
git branch -M main
git add .
git commit -m "Initial commit"
```

We will be writing database migrations using [goose](https://github.com/pressly/goose).
Install goose in your PATH using the following command:

```plain
go install github.com/pressly/goose/v3/cmd/goose@latest
```

Install [sqlx](https://jmoiron.github.io/sqlx/) and the PostgreSQL driver [pq](https://github.com/lib/pq):

```plain
go install github.com/lib/pq
go install github.com/jmoiron/sqlx
```

Now, let's set up a database. First, create a `.envrc` file. We will be using this file to set environment variables using [direnv](https://direnv.net/).

```shell
export PGDATABASE=academy_dev
export DATABASE_URL="postgres://postgres:postgres@localhost/${PGDATABASE}?sslmode=disable"
export GOOSE_MIGRATION_DIR=db/migrations
export GOOSE_DRIVER=postgres
export GOOSE_DBSTRING="$DATABASE_URL"
```

By setting the `GOOSE_MIGRATION_DIR` environment variable, we instruct Goose to look for migration files in the `db/migrations` directory.
The `GOOSE_DBSTRING` makes Goose run the migration scripts against our development database.
In the command line, source this script or run `direnv allow` to apply these settings:

```shell
# If you have configured direnv
direnv allow
# Otherwise just source this file
source .envrc
```

The `.envrc` file is likely to contain secrets that should not be committed to Git.
You can add the actual `.envrc` file to the local `.gitignore` and create a sample `.gitignore` instead:

```plain
echo .envrc >> .gitignore
cp .envrc .envrc.sample
```

### Create a `users` table

Generate a migration file for the `users` table:

```plain
goose create create_users sql
```

In the newly created migration file (called `db/migrations/20240511103916_create_users.sql` in my case, the timestamp part will be different for you), add instructions to create and tear down a `users` table:

```sql
-- +goose Up
-- +goose StatementBegin
create extension if not exists citext;

create table users (
  id bigint primary key generated by default as identity,
  email citext not null unique,
  display_name varchar(80) not null,
  password_hash varchar(100) not null,
  inserted_at timestamp(0) not null default (now() at time zone 'utc'),
  updated_at timestamp(0) not null default (now() at time zone 'utc')
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table users;
-- +goose StatementEnd
```

### Build a database interface for the `users` table

In `types/user.go`, define types representing records in the `users` table and new user registration params:

```go
package types

import (
	"time"

	"github.com/gookit/validate"
)

type User struct {
	ID           int       `db:"id"`
	Email        string    `db:"email"`
	DisplayName  string    `db:"display_name"`
	PasswordHash string    `db:"password_hash"`
	InsertedAt   time.Time `db:"inserted_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type NewUserParams struct {
	Email                string `schema:"email" validate:"required|email"`
	DisplayName          string `schema:"displayName" validate:"required"`
	Password             string `schema:"password" validate:"required|min_len:8|max_len:80"`
	PasswordConfirmation string `schema:"passwordConfirmation" validate:"required|eq_field:Password"`
}

func (p NewUserParams) Messages() map[string]string {
	return validate.MS{
		"required": "can't be blank",
		"email":    "is not a valid email address",
		"min_len":  "must be between 8 and 80 characters long",
		"max_len":  "must be between 8 and 80 characters long",
		"eq_field": "passwords do not match",
	}
}
```

On the `User` type, we define `db:` annotations, so that `sqlx` can map database columns to struct fields (these are not required if the struct field names match database columns).
On the `NewUserParams` struct type, we define annotations for [gorilla/schema](https://github.com/gorilla/schema) and [gookit/validate](https://github.com/gookit/validate). Later on, we will be using `gorilla/schema` to convert HTTP POST data to structs. `gookit/validate` is a simple validation library.

For reasons I cannot fathom, the Golang ecosystem has settled on the [go-playground/validator](https://pkg.go.dev/github.com/go-playground/validator) library as the state of the art in terms of struct validation.
I have found this library to be good for validation, but a pain in the neck whenever I had to customize error messages.
`gookit/validate` is much simpler, and customizing error messages is much simpler as well.

In `store/user_store.go`, define a `userStore` struct. We will be using this type to implement basic CRUD (**C**reate-**R**ead-**U**pdate-**D**elete) operations. For now, let's write an `InsertUser` method to insert pre-validated records into the database. Later on, we will be building on top of this method to implement a user registration workflow.

```go
package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/moroz/webauthn-academy-go/types"
)

type UserStore struct {
	db *sqlx.DB
}

func NewUserStore(db *sqlx.DB) UserStore {
	return UserStore{db}
}

const insertUserQuery = `insert into users (email, display_name, password_hash) values ($1, $2, $3) returning *`

func (s *UserStore) InsertUser(user *types.User) (*types.User, error) {
	var result types.User
	err := s.db.Get(&result, insertUserQuery, user.Email, user.DisplayName, user.PasswordHash)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
```

In `service/user_service.go`, define a `UserService` type. We will be using this type to implement higher-level database interactions.
While the `InsertUser` function in the previous example was a simple `INSERT` operation, the `RegisterUser` method on the `UserService` struct also handles data validation using `gookit/validate` and password hashing using [alexedwards/argon2id](https://github.com/alexedwards/argon2id).

```go
package service

import (
	"github.com/alexedwards/argon2id"
	"github.com/gookit/validate"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/moroz/webauthn-academy-go/store"
	"github.com/moroz/webauthn-academy-go/types"
)

func init() {
	validate.Config(func(opt *validate.GlobalOption) {
		opt.StopOnError = false
	})
}

type UserService struct {
	store store.UserStore
}

func NewUserService(db *sqlx.DB) UserService {
	return UserService{store.NewUserStore(db)}
}

func (s *UserService) RegisterUser(params types.NewUserParams) (*types.User, error, validate.Errors) {
	v := validate.Struct(params)

	if !v.Validate() {
		return nil, nil, v.Errors
	}

	passwordHash, err := argon2id.CreateHash(params.Password, argon2id.DefaultParams)
	if err != nil {
		return nil, err, nil
	}

	user, err := s.store.InsertUser(&types.User{
		Email:        params.Email,
		PasswordHash: passwordHash,
		DisplayName:  params.DisplayName,
	})

	if err == nil {
		return user, nil, nil
	}

	// https://www.postgresql.org/docs/current/errcodes-appendix.html
	// Error 23505 `unique_violation` means that a unique constraint has
    // prevented us from inserting a duplicate value. Instead of returning
    // a raw error, we return a handcrafted validation error that we can
    // later display in a form.
	if err, ok := err.(*pq.Error); ok && err.Code == "23505" && err.Constraint == "users_email_key" {
		validationErrors := validate.Errors{}
		validationErrors.Add("Email", "unique", "has already been taken")
		return nil, nil, validationErrors
	}

	return nil, err, nil
}
```

### Prepare a test suite

Next, we can test our data validation and the registration logic using unit tests.
Go comes with a built-in testing engine, but writing tests with just the standard library tooling is very tedious and repetitive.
Therefore we are going to install [stretchr/testify](https://pkg.go.dev/github.com/stretchr/testify).

Then, in `.envrc`, define two new environment variables: `TEST_DATABASE_NAME` and `TEST_DATABASE_URL`.
We will be using these variables to create and connect to the test database.
Then, define Makefile targets to prepare the test database and run the test suites:

```makefile
guard-%:
	@ test -n "${$*}" || (echo "FATAL: Environment variable $* is not set!"; exit 1)

db.test.prepare: guard-TEST_DATABASE_NAME guard-TEST_DATABASE_URL
	@ createdb ${TEST_DATABASE_NAME} 2>/dev/null || true
	@ env GOOSE_DBSTRING="${TEST_DATABASE_URL}" goose up

test: db.test.prepare
	go test -v ./...
```

This file utilizes GNU `make` syntax extensions to define a dynamic `guard-%` target, which ensures that each required environment variable is set and non-empty.
We then use these guards to validate the environment before running the `db.test.prepare` target, which creates a test database and runs migrations against this database.
Finally, the `test` target runs the test suites of all packages in the project. Since the `test` target lists `db.test.prepare` as a dependency, `make` will ensure that all the migrations are correctly applied against the test database before the test suites are executed.

In `service/service_test.go`, define a test suite using `stretchr/testify`. This file does not define any specific tests, only a scaffolding for the tests we are going to add in other files.

```go
package service_test

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	db *sqlx.DB
}

func (s *ServiceTestSuite) SetupTest() {
	conn := os.Getenv("TEST_DATABASE_URL")
	s.db = sqlx.MustConnect("postgres", conn)
	s.db.MustExec("truncate users cascade")
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
```

With this file in place, we can set up more specific tests for registration logic. In `service/user_registration.go`, add tests for the user service:

```go
package service_test

import (
	"github.com/alexedwards/argon2id"
	"github.com/moroz/webauthn-academy-go/service"
	"github.com/moroz/webauthn-academy-go/store"
	"github.com/moroz/webauthn-academy-go/types"
)

func (s *ServiceTestSuite) TestRegisterUser() {
	params := types.NewUserParams{
		Email:                "registration@example.com",
		DisplayName:          "Example User",
		Password:             "foobar123123",
		PasswordConfirmation: "foobar123123",
	}

	srv := service.NewUserService(s.db)
	user, err, _ := srv.RegisterUser(params)
	s.NoError(err)
	s.Equal(params.Email, user.Email)
	s.Equal(params.DisplayName, user.DisplayName)

	match, err := argon2id.ComparePasswordAndHash(params.Password, user.PasswordHash)
	s.True(match)
}

func (s *ServiceTestSuite) TestRegisterUserWithInvalidParams() {
	params := types.NewUserParams{
		Email:                "invalid",
		DisplayName:          "Example User",
		Password:             "short",
		PasswordConfirmation: "not matching",
	}

	srv := service.NewUserService(s.db)
	user, err, validationErrors := srv.RegisterUser(params)
	s.NoError(err)
	s.Nil(user)
	msg := validationErrors.FieldOne("Email")
	s.Equal("is not a valid email address", msg)
	msg = validationErrors.FieldOne("Password")
	s.Equal("must be between 8 and 80 characters long", msg)
	msg = validationErrors.FieldOne("PasswordConfirmation")
	s.Contains(msg, "do not match")
}

func (s *ServiceTestSuite) TestRegisterUserWithDuplicateEmail() {
	store := store.NewUserStore(s.db)
	user, err := store.InsertUser(&types.User{
		Email:        "duplicate@email.com",
		PasswordHash: "test",
		DisplayName:  "John Smith",
	})

	s.NoError(err)

	srv := service.NewUserService(s.db)

	params := types.NewUserParams{
		Email:                user.Email,
		DisplayName:          "Other User",
		Password:             "foobar123123",
		PasswordConfirmation: "foobar123123",
	}
	user, err, validationErrors := srv.RegisterUser(params)
	s.Nil(user)
	s.Nil(err)
	msg := validationErrors.FieldOne("Email")
	s.Equal("has already been taken", msg)
}
```

If you run the tests now, they should all pass:

```plain
$ make test
2024/05/16 00:01:45 goose: no migrations to run. current version: 20240511103916
go test -v ./...
?   	github.com/moroz/webauthn-academy-go	[no test files]
?   	github.com/moroz/webauthn-academy-go/handler	[no test files]
?   	github.com/moroz/webauthn-academy-go/store	[no test files]
?   	github.com/moroz/webauthn-academy-go/types	[no test files]
=== RUN   TestServiceTestSuite
=== RUN   TestServiceTestSuite/TestRegisterUser
=== RUN   TestServiceTestSuite/TestRegisterUserWithDuplicateEmail
=== RUN   TestServiceTestSuite/TestRegisterUserWithInvalidParams
--- PASS: TestServiceTestSuite (0.20s)
    --- PASS: TestServiceTestSuite/TestRegisterUser (0.10s)
    --- PASS: TestServiceTestSuite/TestRegisterUserWithDuplicateEmail (0.06s)
    --- PASS: TestServiceTestSuite/TestRegisterUserWithInvalidParams (0.03s)
PASS
ok  	github.com/moroz/webauthn-academy-go/service	(cached)
```

## Set up a router

Now that we have the database logic in place, we can try and build a sign up view using HTML and CSS.
First, install [go-chi/chi](https://github.com/go-chi/chi)---a router for use with `net/http`:

```plain
go get -u github.com/go-chi/chi/v5
```

Then, in `main.go`, we can set up a router like so:

```go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "<h1>Hello from the router!</h1>")
	})

	log.Println("Listening on port 3000")
	log.Fatal(http.ListenAndServe(":3000", r))
}
```

Run the server:

```plain
$ go run .
2024/05/19 14:46:01 Listening on port 3000
```

When you visit [localhost:3000](http://localhost:3000) now, you should be greeted by this view:

<figure class="bordered-figure">
<a href="/golang/01-router-hello-world.png" target="_blank" rel="noopener noreferrer"><img src="/golang/01-router-hello-world.png" alt="" /></a>
<figcaption>A &ldquo;Hello world&rdquo;-like message served using <code>chi-router</code>.</figcaption>
</figure>

### Set up `templ` for HTML templating

We will be building templates using [templ](https://templ.guide/) instead of Go's built-in `html/template` package.
This is because Templ makes it much easier to share common data between views (such as flash messages, authentication status, page title, etc.).
Install the templ CLI:

```plain
go install github.com/a-h/templ/cmd/templ@latest
```

Next, define the basic HTML layouts at `templates/layout/root.templ`:

```go{data-lang="templ"}
package layout

templ RootLayout(title string) {
	<!DOCTYPE html>
	<html lang="en">
		<head>
			<meta charset="UTF-8"/>
			<title>{ title } | Academy</title>
		</head>
		<body>
			{ children... }
		</body>
	</html>
}

templ Unauthenticated(title string) {
	@RootLayout(title) {
		<div class="layout unauthenticated">
			{ children... }
		</div>
	}
}
```

We define two layout templates: `RootLayout`, which is the base HTML layout for all context-specific layouts in the application, and `Unauthenticated`, a basic layout used for views shown to unauthenticated visitors, such as the login page or the registration page.

In `handler/templates/users/new.html.tmpl`, add the registration form template:

```go{data-lang="templ"}
package users 

import "github.com/gookit/validate"
import "github.com/moroz/webauthn-academy-go/types"
import "github.com/moroz/webauthn-academy-go/templates/layout"

func fieldClass(error string) string {
	if error != "" {
		return "field has-error"
	}
	return "field"
}

templ New(params types.NewUserParams, errors validate.Errors) {
	@layout.Unauthenticated("Register") {
		<form action="/users/register" method="POST" class="card">
			<header>
				<h1>Register</h1>
			</header>
			<div class={ fieldClass(errors.FieldOne("Email")) }>
				<label for="email">Email:</label>
				<input
					id="email"
					type="email"
					name="email"
					value={ params.Email }
					autocomplete="email"
					autofocus
				/>
				<p class="error-explanation">{ errors.FieldOne("Email") }</p>
			</div>
			<div class={ fieldClass(errors.FieldOne("DisplayName")) }>
				<label for="displayName">Display name:</label>
				<input
					id="displayName"
					type="text"
					name="displayName"
					value={ params.DisplayName }
					autocomplete="name"
				/>
				<p class="error-explanation">{ errors.FieldOne("DisplayName") }</p>
			</div>
			<div class={ fieldClass(errors.FieldOne("Password")) }>
				<label for="password">Password:</label>
				<input
					id="password"
					type="password"
					name="password"
					autocomplete="new-password"
				/>
				<p class="error-explanation">{ errors.FieldOne("Password") }</p>
			</div>
			<div class={ fieldClass(errors.FieldOne("PasswordConfirmation")) }>
				<label for="passwordConfirmation">Confirm password:</label>
				<input
					id="passwordConfirmation"
					type="password"
					name="passwordConfirmation"
					autocomplete="new-password"
				/>
				<p class="error-explanation">{ errors.FieldOne("PasswordConfirmation") }</p>
			</div>
			<div>
				<button type="submit" class="button is-fullwidth is-primary">
					Submit
				</button>
			</div>
			<footer>
				<p>Already have an account? <a href="/sign-in">Sign in</a></p>
			</footer>
		</form>
	}
}
```

You can generate Go code from `.templ` files using this command:

```plain
templ generate
```

Now we can write a handler that will render these templates in response to HTTP requests.

```go
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
```

Update `main.go` to serve requests to `GET /` with this handler:

```go
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/moroz/webauthn-academy-go/handler"
)

func main() {
	db := sqlx.MustConnect("postgres", os.Getenv("DATABASE_URL"))

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	users := handler.UserHandler(db)
	r.Get("/", users.New)

	log.Println("Listening on port 3000")
	log.Fatal(http.ListenAndServe(":3000", r))
}
```

If you re-run this project now (using `go run .` in the project's root directory) and navigate to [localhost:3000](http://localhost:3000), you should be greeted with an unstyled registration form like the one below:

<figure class="bordered-figure">
<a href="/golang/02-sign-up-without-css.png" target="_blank" rel="noopener noreferrer"><img src="/golang/02-sign-up-without-css.png" alt="" /></a>
<figcaption>The sign up page rendered without CSS at 200% zoom.</figcaption>
</figure>

### Set up Vite for asset bundling

We will be using [Vite](https://vitejs.dev/) to compile and bundle CSS and JavaScript assets.
First, install the [pnpm package manager](https://pnpm.io/) for node using `npm`:

```plain
npm i -g pnpm
```

Then, create a Vite project under `assets/`:

```plain
pnpm create vite@latest assets --template vanilla-ts
cd assets
pnpm install
```

### Code reloading with `modd`

With Vite added to the project, we will have to run the Vite development server in the background alongside the application.
At this point, running multiple commands (`templ generate` and `go run .`) just to rebuild the code could already become very tedious.
Let's set up [modd](https://github.com/cortesi/modd) to rebuild templates and application code.

Start by installing `modd`:

```plain
go install github.com/cortesi/modd/cmd/modd@latest
```

Then, in a file named `modd.conf` in the root directory of the project, add the following configuration:

```plain
{
  daemon +sigterm: cd assets/ && pnpm run dev --port=5173
}

**/*.templ {
  prep +onchange: templ generate
}

**/*.go !**/*_test.go {
  prep +onchange: go build -o server .
  daemon +sigterm: ./server
}
```

This file instructs `modd` to:

* always start the Vite development server in the background whenever we start the project,
* regenerate view code whenever a `.templ` file is modified,
* rebuild and restart the application whenever `.go` files are modified (including view code).

Update `.gitignore` to look like this:

```shell
# Compiled server executable
/server

# Go code generated by templ
**/*_templ.go

# Environment variables (machine-specific values and secrets)
.envrc
```

Now, terminate the application server if you still had it running, and run `modd` in the terminal.
With a correct setup, the tool should regenerate your views and start the Vite development server:

```plain
$ modd
20:06:47: skipping prep: templ generate
20:06:47: skipping prep: go build -o server .
20:06:47: daemon: cd assets/ && pnpm run dev --port=5173
>> starting...
20:06:47: daemon: ./server
>> starting...
2024/05/20 20:06:47 Listening on port 3000
```

### Style the page with CSS

Now we can add some CSS to make the page more presentable. We will be writing CSS by hand to show you how simple this can be.

Install [dart-sass](https://sass-lang.com/) to compile stylesheets:

```plain
pnpm add sass
```

Create an empty directory at `assets/src/css` and an empty file therein:

```plain
mkdir -p assets/src/css
touch assets/src/css/style.scss
```

Replace the contents of `assets/src/main.ts` with a single line, importing the SCSS entrypoint file:

```javascript
import "./css/style.scss";
```

In the `RootLayout` template in `templates/layout/root.templ`, add a `<script>` tag to load assets with Vite:

```go{data-lang="templ"}
// ...

templ RootLayout(title string) {
	<!DOCTYPE html>
	<html lang="en">
		<head>
			<meta charset="UTF-8"/>
			<title>{ title } | Academy</title>
			<script type="module" src="http://localhost:5173/src/main.ts"></script>
		</head>
		<body>
			{ children... }
		</body>
	</html>
}

// ...
```

In development, this change is enough to load the Vite project in the browser, and the script will automatically inject CSS into the DOM.
However, in production builds, the JavaScript files will be compiled and minified into separate JavaScript and CSS files, and we will need to load them separately.
This is a bit more involved than the above example, however we don't really need to think about this until we start preparing the project for production deployments.

In `assets/src/css/_palette.scss`, add a few colors (they are all borrowed from [a certain CSS toolkit that I otherwise don't want to use](https://tailwindcss.com/docs/customizing-colors), but it's okay since the aforementioned toolkit is MIT-licensed).

```scss
$green-50: #f0fdf4;
$green-100: #dcfce7;
$green-900: #14532d;

$red-100: #fee2e2;
$red-200: #fecaca;
$red-600: #dc2626;
$red-900: #7f1d1d;

$danger: $red-600;

$body-bg: $green-50;
$primary: $green-900;
$primary-darker: desaturate($primary, 20%);

$family-sans: Inter, Arial, Helvetica, sans-serif;
```

First, let's add some styles to center the form within the page:

```scss
@import "./palette";

*,
*::after,
*::before {
  box-sizing: border-box;
}

html,
body {
  margin: 0;
  padding: 0;
  font-size: 100%;
  font-family: $family-sans;
}

body {
  background: $body-bg;
  color: #000;
}

h1,
h2 {
  color: $primary-darker;
}

a,
a:visited {
  color: $primary;
}

.layout.unauthenticated {
  display: grid;
  place-items: center;
  height: 100vh;

  p:not([class]) {
    margin: 0;
  }

  footer,
  header {
    text-align: center;
  }

  header {
    margin-bottom: 0.75rem;
  }

  footer {
    margin-top: 1.75rem;
  }
}

.card {
  padding: 1.5rem;
  border: 1px solid $primary-darker;
  box-shadow: 0 0 10px transparentize($primary, 0.8);
  border-radius: 5px;
  background: white;

  h1,
  h2 {
    margin-top: 0;
    margin-bottom: 0.5rem;
    text-align: center;
  }
}
```

With these changes in place, the form should be displayed in a white card, centered on a light-green page:

<figure class="bordered-figure">
<a href="/golang/03-sign-up-with-layout.png" target="_blank" rel="noopener noreferrer"><img src="/golang/03-sign-up-with-layout.png" alt="" /></a>
<figcaption>Partly styled sign up page at 200% zoom.</figcaption>
</figure>

Then, let's style the form fields and the submit button:

```scss
.field {
  margin-bottom: 0.75rem;
  min-width: 350px;

  label {
    display: block;
    font-weight: bold;
    margin-bottom: 0.25rem;
    color: $primary-darker;
  }

  input {
    width: 100%;
  }

  &.has-error {
    label {
      color: $danger;
    }

    input {
      border-color: $danger;

      &:focus {
        outline-color: $danger;
      }
    }
  }
}

.error-explanation {
  margin-top: 0.25rem;
  color: $danger;

  &:empty {
    display: none;
  }
}

input[type="text"],
input[type="password"],
input[type="email"] {
  border: 1px solid #666;
  height: 40px;
  border-radius: 3px;
  font: inherit;
  padding-left: 0.75rem;
  padding-right: 0.75rem;
}

.button.is-primary {
  color: #fff;
  background: $primary;
  font-weight: bold;
  font-family: inherit;
  font-size: 1rem;
  height: 40px;
  outline: 0;
  border: 0;
  border-radius: 3px;
  margin-top: 0.5rem;
}

.button.is-fullwidth {
  width: 100%;
}
```

The sign up page should now begin to look like this:

<figure class="bordered-figure">
<a href="/golang/04-sign-up-styled.png" target="_blank" rel="noopener noreferrer"><img src="/golang/04-sign-up-styled.png" alt="" /></a>
<figcaption>Fully styled sign up page at 200% zoom.</figcaption>
</figure>
