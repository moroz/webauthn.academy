---
title: Implementing Webauthn in Golang
layout: single
---

This section is dedicated to an implementation of a WebAuthn ([Web Authentication](https://developer.mozilla.org/en-US/docs/Web/API/Web_Authentication_API)) workflow using the Go programming language.

This website is not meant as a complete learning resource for beginners, but rather a reference implementation of a complete Webauthn workflow.
It is the text that I wish were available when I first started implementing Webauthn in my applications.

## Who this text is for

In general, I have written this text with an experienced audience in mind.
On the other hand, even if you are a beginner, building a Web application as described in this walkthrough, although painful, is likely going to benefit you more than watching another learn-Python-in-3-hours video or trying to wish a job into existence.
You are going to need a reasonably good knowledge of back end development, the UNIX command line, SQL, and all three languages used in browser environments (HTML, CSS, and JavaScript).
I may leave several difficult code snippets unexplained, which I believe should be readable without explanation.

If you have any suggestions for improvements to the tutorial, feel free to [reach out to me](https://github.com/moroz) or to submit a Pull Request in the [Github repository](https://github.com/moroz/webauthn.academy) of this website.

The source code of the application we are going to build is available on Github: [moroz/webauthn-academy-go](https://github.com/moroz/webauthn-academy-go).

## Technological stack

This website was developed on a variety of Linux-powered machines, using Go 1.23.1 and Node 20.17.0, and PostgreSQL 16.4.
The back end application should work exactly the same on any UNIX-like operating system, and possibly even on Windows.
However, the Web Authentication API in the browser is only supported on Linux, macOS, and Windows.

There are several command-line tools we will be using in this walkthrough:

* [mise](https://mise.jdx.dev/) --- to manage different versions of programming languages, here Go and Node.js.
* [goose](https://github.com/pressly/goose) --- to generate and run database migrations,
* [direnv](https://direnv.net/) --- to manage settings and secrets in environment variables.
* [modd](https://github.com/cortesi/modd) --- to automatically rebuild and reload the application upon changes to the source code.
* [sqlc](https://sqlc.dev/) --- to generate type-safe code for database operations based on the database schema.

Whenever possible, I try to use just the standard library, so with enough knowledge of the Go ecosystem, you should be able to modify the solution to use your preferred libraries.
However, there are several libraries that handle everyday tasks much more elegantly than the standard library.
A few notable Go examples:

* [github.com/alexedwards/argon2id](https://pkg.go.dev/github.com/alexedwards/argon2id) --- to hash passwords using the Argon2id password hashing algorithm.
* [github.com/go-webauthn/webauthn](https://github.com/go-webauthn/webauthn) --- the actual WebAuthn implementation. We will be using this library to generate and validate registration and attestation challenges.
* [templ](https://templ.guide/) --- a type-safe templating language that compiles to Go.
* [github.com/gorilla/schema](https://github.com/gorilla/schema) --- to parse URL-encoded data into structs.
* [github.com/gorilla/sessions](https://pkg.go.dev/github.com/gorilla/sessions) --- for persisting session state in cookies. We will be using session storage to display flash notifications, for CSRF protection, and to persist WebAuthn challenges across requests.
* [github.com/gookit/validate](https://github.com/gookit/validate/) --- for struct validation.

We will be bundling CSS and JavaScript code using [Vite](https://vitejs.dev/), [TypeScript](https://www.typescriptlang.org/), and [SASS](https://sass-lang.com/).

## Initial setup

The following walkthrough sets up a password authentication from scratch. Once this text is finalized, you will be able to skip to the section where I start implementing Webauthn. For now, you can just follow along.

Create a directory for the new project:

```plain
mkdir academy-go
```

Ensure Golang is installed (here using [mise](https://mise.jdx.dev/)):

```shell
$ cd academy-go
$ mise install go@1.23.1 node@lts

# Save preferred versions of the Go toolchain and Node.js to .tool-versions
$ mise local go@1.23.1
$ mise local node@lts

# Check that Go and Node.js are installed with the correct versions
$ go version
go version go1.23.1 linux/amd64
$ node --version
v20.17.0
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

In the remaining part of this article, I will not include any git commands or commit messages, as this would make the article too verbose.
I encourage you to commit and push often, even if you think there is "nothing to commit," or if you haven't finished your tasks.
It is also good practice to write [informative commit messages](https://cbea.ms/git-commit/).

## Simple HTTP router using `chi-router`

Let's build a simple HTTP handler to process incoming requests.
First, install [go-chi/chi](https://github.com/go-chi/chi)---a routing library built on top of the `net/http` package from the Go standard library:

```plain
go get -u github.com/go-chi/chi/v5
```

Then create a file named `main.go` in the project's root directory:

{{< gist "golang/006-main.go" "go" "main.go" >}}

This file uses a common Go idiom: the blocking call to `http.ListenAndServe` is wrapped in a call to `log.Fatal`.
The blocking call will only return a value if the operation fails, for instance if a port is already in use.
`log.Fatal` will then log the error message and terminate the program with a non-zero exit code.

Run the server:

```shell
$ go run .
2024/05/19 14:46:01 Listening on port 3000
```

When you visit [localhost:3000](http://localhost:3000) now, you should be greeted by this view:

{{< figure "/golang/01-router-hello-world.png" "A &ldquo;Hello world&rdquo;-like message served using <code>chi-router</code>." >}}

## Database schema migrations using `goose`

In this section, we are going to set up `goose`, a command-line tool for database [schema migrations](https://en.wikipedia.org/wiki/Schema_migration).

### Installing `goose` using `tools.go`

Since `goose` is a command-line tool separate from our program logic, we want to install it as a standalone application.
The Go toolchain allows us to track versions of CLI tools in `go.mod` using a technique called `tools.go`.

In the root directory of the project, run this command to download `goose` and add it to the project's dependencies:

```shell
go get github.com/pressly/goose/v3/cmd/goose@latest
```

In the same directory, create a file called `tools.go` with the following contents:

{{< gist "golang/021-tools.go" "go" "tools.go" >}}

Then, create a `Makefile` with the following contents:

{{< gist "golang/020-Makefile" "makefile" "Makefile" >}}

If you run `make install.tools` now, you should end up with `goose` correctly installed in `PATH`:

```shell
$ make install.tools
go mod download
Installing tools from tools.go
go install github.com/pressly/goose/v3/cmd/goose
$ which goose
/home/karol/.local/share/mise/installs/go/1.23.1/bin/goose
```

Now, let's set up a database. First, create a `.envrc` file. We will be using this file to set environment variables using [direnv](https://direnv.net/).

{{< gist "golang/022-envrc" "shell" ".envrc" >}}

By setting a `PGDATABASE` variable, we instruct the PostgreSQL CLI tools to connect to the project's database by default.
`DATABASE_URL` is a database connection string in URL format.

`GOOSE_MIGRATION_DIR` instructs Goose to look for migration files in the `db/migrations` directory.
The `GOOSE_DBSTRING` makes Goose run the migration scripts against our development database.
In the command line, source this script or run `direnv allow` to apply these settings and create a database:

```shell
# If you have configured direnv
$ direnv allow

# Otherwise just source this file
$ source .envrc

# Create the database 
$ createdb
```

The `.envrc` file should not be committed to Git. There are two main reasons for that: Firstly, in the future, the `.envrc` will likely contain some secrets that should not be exposed to the outside world, such as passwords, API tokens, etc.
Secondly, every time you set up the project for local work, you may want to apply some changes to these variables that do not need to be propagated to the upstream Git repository.
Therefore, we tell Git to ignore this file by adding its filename to `.gitignore` and commit a safe `.envrc.sample` file instead:

```shell
# Create a file with the contents of ".envrc"
# or append a line if the file exists
$ echo .envrc >> .gitignore

# Copy the content to the file we want to commit
$ cp .envrc .envrc.sample
```

If you add new required environment variables to your local `.envrc` file, make sure to also update `.envrc.sample`.

### `create_users` migration

Create a directory for database migrations. `goose` will not create one automatically and will fail with an error message if the directory does not exist when creating a migration file.

```shell
mkdir -p db/migrations
```

Generate a migration file for the `users` table. Do note that the file name contains a timestamp and will therefore be different each time you run this command.

```shell
$ goose create create_users sql
2024/09/13 08:00:48 Created new file: db/migrations/20240913000048_create_users.sql
```

In the newly created migration file, add instructions to create and tear down a `users` table:

{{< gist "golang/000-create-users.sql" "sql" "db/migrations/20240913000048_create_users.sql" >}}

You can execute this migration using `goose up`:

```shell
$ goose up
2024/09/13 10:48:01 OK   20240913000048_create_users.sql (9.66ms)
2024/09/13 10:48:01 goose: successfully migrated database to version: 20240913000048
```

You can check whether the migration was successful by connecting to the database using `psql` and requesting information about the `users` table using `\d+ users`.

This migration should create a table with the following columns:

* `id`: an automatically generated primary key of type `bigint` (equivalent of `int64`),
* `email`: case-insensitive string column with a unique index,
* `display_name`,
* `password_hash`: string column to store an password hashed using Argon2 in [PHC string format](https://github.com/P-H-C/phc-string-format/blob/master/phc-sf-spec.md),
* `inserted_at` and `updated_at`: to store creation and modification times, respectively. The timestamps are stored without milliseconds (hence the type name `timestamp(0)`). We will store the times in the UTC time zone, regardless of your geographical location.

## Type-safe SQL using `sqlc`

Add the `sqlc` dependency to `go.mod`:

```shell
go get github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

Add the dependency to `tools.go`:

{{< gist "golang/022-tools.go" "go" "tools.go" "hl_lines=[8]" >}}

Install tools using `make install.tools`:

```shell
$ make install.tools 
go mod download
Installing tools from tools.go
go install github.com/pressly/goose/v3/cmd/goose
go install github.com/sqlc-dev/sqlc/cmd/sqlc

$ which sqlc
/home/karol/.local/share/mise/installs/go/1.23.1/bin/sqlc
```

Create a configuration file at `sqlc.yml`:

{{< gist "golang/023-sqlc.yml" "yaml" "sqlc.yml" >}}

This config file tells `sqlc` to generate code for all queries defined in `db/sql/*.sql`.
`sqlc` will infer data types for database columns based on the schema migrations defined with `goose`.

{{< gist "golang/024-users.sql" "sql" "db/sql/users.sql" >}}

Run the generator:

```shell
sqlc generate
```

If there is no output, it means that the generation has completed successfully.

-- Unrevised content below -- 

### Build a database interface for the `users` table

In `types/user.go`, define types representing records in the `users` table and new user registration params:

{{< file "golang/001-users.go" "go" >}}

On the `NewUserParams` struct type, we define annotations for [gorilla/schema](https://github.com/gorilla/schema) and [gookit/validate](https://github.com/gookit/validate). Later on, we will be using `gorilla/schema` to convert HTTP POST data to structs. `gookit/validate` is a simple validation library.

For reasons I cannot fathom, the Golang ecosystem has settled on the [go-playground/validator](https://pkg.go.dev/github.com/go-playground/validator) library as the state of the art in terms of struct validation.
I have found this library to be good for validation, but a pain in the neck whenever I had to customize error messages.
`gookit/validate` is much simpler, and customizing error messages is much simpler as well.

In `store/user_store.go`, define a `userStore` struct. We will be using this type to implement basic CRUD (**C**reate-**R**ead-**U**pdate-**D**elete) operations. For now, let's write an `InsertUser` method to insert pre-validated records into the database. Later on, we will be building on top of this method to implement a user registration workflow.

{{< file "golang/002-user-store.go" "go" >}}

In `service/user_service.go`, define a `UserService` type. We will be using this type to implement higher-level database interactions.
While the `InsertUser` function in the previous example was a simple `INSERT` operation, the `RegisterUser` method on the `UserService` struct also handles data validation using `gookit/validate` and password hashing using [alexedwards/argon2id](https://github.com/alexedwards/argon2id).

{{< file "golang/003-user-service.go" "go" >}}

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

{{< file "golang/004-service-test.go" "go" >}}

With this file in place, we can set up more specific tests for registration logic. In `service/user_service_test.go`, add tests for the user service:

{{< file "golang/005-user-service-test.go" "go" >}}

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

## Create a configuration package

In `config/config.go`, add a module to encapsulate the logic for reading and validating application configuration from environment variables.

{{< file "golang/005-config-helper.go" "go" >}}

The helper function `MustGetenv` wraps [`os.Getenv`](https://pkg.go.dev/os#Getenv) so that if any required environment variable is unset or empty, the function will log an error message and terminate the program. Failing early helps identify configuration errors early on, and putting configuration in a single, independent package allows us to import this package anywhere in the program, without having to worry about circular dependency errors.

For now, we only 


### Set up `templ` for HTML templating

We will be building templates using [templ](https://templ.guide/) instead of Go's built-in `html/template` package.
This is because Templ makes it much easier to share common data between views (such as flash messages, authentication status, page title, etc.).
Install the templ CLI:

```plain
go install github.com/a-h/templ/cmd/templ@latest
```

Next, define the basic HTML layouts at `templates/layout/root.templ`:

{{< file "golang/007-root.templ" "templ" >}}

We define two layout templates: `RootLayout`, which is the base HTML layout for all context-specific layouts in the application, and `Unauthenticated`, a basic layout used for views shown to unauthenticated visitors, such as the login page or the registration page.

In `templates/users/users.templ`, add the registration form template:

{{< file "golang/008-users.templ" "templ" >}}

You can generate Go code from `.templ` files using this command:

```plain
templ generate
```

Now we can write a handler that will render these templates in response to HTTP requests.
In `handler/user_handler.go`, add the following:

{{< file "golang/009-users-handler.go" "go" >}}

Update `main.go` to serve requests to `GET /` with this handler:

{{< file "golang/010-main.go" "go" >}}

Do note that in line 10 we need to import the [`github.com/lib/pq`](https://github.com/lib/pq) library using an `import` statement with the blank identifier `_` as an [explicit package name](https://go.dev/ref/spec#Import_declarations). This package is never called directly, but this import statement is required for its side effects. If you forget to add this import, the call to [`sqlx.MustConnect`](https://pkg.go.dev/github.com/jmoiron/sqlx#MustConnect) will result in a panic.

If you re-run this project now (using `go run .` in the project's root directory) and navigate to [localhost:3000](http://localhost:3000), you should be greeted with an unstyled registration form like the one below:

{{< figure "/golang/02-sign-up-without-css.png" "The sign up page rendered without CSS at 200% zoom." >}}

Now, let's set up some styling.

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

{{< file "golang/011-root.templ" "templ" >}}

In development, this change is enough to load the Vite project in the browser, and the script will automatically inject CSS into the DOM.
However, in production builds, the JavaScript files will be compiled and minified into separate JavaScript and CSS files, and we will need to load them separately.
This is a bit more involved than the above example, however we don't really need to think about this until we start preparing the project for production deployments.

In `assets/src/css/_palette.scss`, add a few colors (they are all borrowed from [a certain CSS toolkit that I otherwise don't want to use](https://tailwindcss.com/docs/customizing-colors), but it's okay since the aforementioned toolkit is MIT-licensed).

{{< file "golang/012-palette.scss" "scss" >}}

First, let's add some styles to center the form within the page:

{{< file "golang/013-style.scss" "scss" >}}

With these changes in place, the form should be displayed in a white card, centered on a light-green page:

{{< figure "/golang/03-sign-up-with-layout.png" "Partly styled sign up page at 200% zoom." >}}

Then, let's style the form fields and the submit button:

{{< file "golang/014-style.scss" "scss" >}}

The sign up page should now begin to look like this:

{{< figure "/golang/04-sign-up-styled.png" "Fully styled sign up page at 200% zoom." >}}

## Sign up handler

Now that the registration form is rendering correctly, we can implement the handler that will process the data submitted by that form.
Since the form is using the POST HTTP method and is not marked as `multipart` (which is only necessary if you want to upload files together with other data in a single request), the request body will be submitted in URL-encoded format ([`application/x-www-form-urlencoded`](https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/POST)).
Once the request reaches the handler, we can parse the data using [`net/http.Request.ParseForm`](https://pkg.go.dev/net/http#Request.ParseForm), which will populate the [`PostForm`](https://pkg.go.dev/net/http#Request) field on the Request struct.
In order to convert the data to a `types.NewUserParams` struct, we could do something like this:

```go {hl_lines=[1 2 3]}
if err := r.ParseForm(); err != nil {
    // handle bad request
    return
}

var params types.NewUserParams
params.Email = r.PostForm.Get("email")
params.DisplayName = r.PostForm.Get("displayName")
params.Password = r.PostForm.Get("password")
params.PasswordConfirmation = r.PostForm.Get("passwordConfirmation")

// actually try to create a User
```

As you can imagine, this approach could become extremely tedious, especially if at some point we decided to submit multiple values per form field (e. g. multiple checkboxes in a fieldset).
Therefore, we are going to use [github.com/gorilla/schema](https://github.com/gorilla/schema) to handle this task for us.

First, install the library:

```plain
go get github.com/gorilla/schema
```

Then, in a new file called `handler/helpers.go`, add the following:

{{< file "golang/015-helpers.go" "go" >}}

The `decoder` variable is a shared instance of the schema decoder that we can use to decode the data submitted by multiple requests.
The `handleError` function is a helper that will help us quickly terminate unprocessable requests with a simple response based on a HTTP status code.

In `handler/user_handler.go`, add a `Create` method that will handle the user creation action.

{{< file "golang/016-create-user.go" "go" >}}

In this method, we decode the submitted HTTP POST data into a `types.NewUserParams` struct, and if the data cannot be parsed, we return a simple [`400 Bad Request`](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/400) error response.
Then we validate the params and attempt to insert them into the database. If the validation fails, we re-render the registration form with error messages. Finally, if everything goes smooth, we redirect the user to the `/sign-in` path, which we have not implemented yet.

## Sign in page

In `templates/sessions/sessions.templ`, add a template for the sign in form:

{{< file "golang/017-sessions.templ" "templ" >}}

Then, in `handler/session_handler.go`, add the handler that will render this form:

{{< file "golang/018-session-handler.go" "go" >}}

In `main.go`, add the new route at `GET /sign-in`:

{{< file "golang/019-main.go" "go {hl_lines=[13 14 15] linenostart=13}" >}}

If you navigate to [http://localhost:3000/sign-in](http://localhost:3000/sign-in), you should see a login page like this:

{{< figure "/golang/05-sign-in.png" "The sign in view at 150% zoom." >}}
