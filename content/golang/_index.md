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

In the newly created migration file, add instructions to create and tear down the `users` table:

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
Other settings of interest include `emit_pointers_for_null_types: true` (instructing `sqlc` to generate structs with pointer types for nullable columns, e. g. `*string` instead of `sql.NullString`).

{{< gist "golang/024-users.sql" "sql" "db/sql/users.sql" >}}

Run the generator:

```shell
sqlc generate
```

If there is no output, it means that the generation has completed successfully.
Now, in `db/queries`, you should find three files, `db.go`, `models.go`, and `users.sql.go`.

These files contain types, interfaces, and methods that we will use to interact with the database further in the project.
Based on the short snippet of SQL code in `db/sql/users.sql`, `sqlc` managed to generate the following code:

{{< gist "golang/025-users.sql.go" "go" "db/queries/users.sql.go" >}}

Even though the query we wrote ended with `returning *`, `sqlc` expanded the asterisk into all the corresponding columns, defined all the necessary data structures, and generated type-safe code for this operation. Impressive!

## Implementing user registration logic

In this section, we will implement the business logic for the user registration workflow.
The logic will be implemented within the `services` package.
Define a `UserService` type in `services/user_service.go`:

{{< gist "golang/026-user_service.go" "go" "services/user_service.go" >}}

At this point, the `RegisterUser` method does nothing.
However, we have a clearly defined API contract: we have defined a type for the input parameters, and know that the method should return a `(*queries.User, error)` tuple.
This is enough to write some unit tests for the `RegisterUser` method and implement the logic later to get the test suite to pass.

According to [this StackOverflow thread](https://stackoverflow.com/questions/2381910/is-it-reasonable-to-enforce-that-unit-tests-should-never-talk-to-a-live-database), what we are going to write is, by definition, an _integration test_, rather than a _unit test_, by merit of talking to an actual database.
However, I find this distinction to be quite useless, because almost all tests that we are going to be writing for this project will be running against a real database.
So, if you really care about correctness, please just keep in mind that everything below is actually an _integration test_.

### Preparing a test database

In order to run our integration tests against a dedicated database, we need to prepare the database in roughly the following steps:

1. **Create** a test database.
2. Run all schema **migrations** against the newly created database.
3. On subsequent runs, **clean** database tables.

First, let us define new environment variables in `.envrc`.
We will be using these variables to create and connect to the test database.

{{< gist "golang/027-envrc" "shell" ".envrc" `{"linenostart":5}` >}}

Make sure to run `direnv allow` to approve these changes and to update `.envrc.sample` accordingly.

In `Makefile`, define the following targets to prepare a test database and run test suites:

{{< gist "golang/028-Makefile" "makefile" "Makefile" `{"linenostart":8}` >}}

This file utilizes GNU `make` syntax extensions to define a dynamic `guard-%` target, which ensures that each required environment variable is set and non-empty.
If you are developing on a system that defaults to BSD `make` (such as FreeBSD), you may need to run all `make` commands as `gmake`.

The `db.test.prepare` uses this dynamic target to validate that both `TEST_DATABASE_NAME` and `TEST_DATABASE_URL` are non-empty before creating a tast database and running schema migrations against it using `goose`. If the database already exists, we ignore all errors and proceed with the schema migrations.

Finally, the `test` target runs the test suites of all packages in the project. Since the `test` target lists `db.test.prepare` as a dependency, `make` will ensure that all the migrations are correctly applied against the test database before the test suites are executed.

With these changes, you should be able to execute both targets and end up with a properly migrated test database, and the testing engine should inform you that it did not manage to find any test files in the project:

```shell
$ make db.test.prepare
2024/09/16 01:00:28 OK   20240913000048_create_users.sql (13.56ms)
2024/09/16 01:00:28 goose: successfully migrated database to version: 20240913000048

$ make test
2024/09/16 01:00:31 goose: no migrations to run. current version: 20240913000048
go test -v ./...
?       github.com/moroz/webauthn-academy-go    [no test files]
?       github.com/moroz/webauthn-academy-go/db/queries [no test files]
?       github.com/moroz/webauthn-academy-go/services   [no test files]
```

### Setting up a test suite

Even though Go comes with a built-in testing engine, writing tests with just the standard library tooling is very tedious and repetitive.
Therefore we are going to install [stretchr/testify](https://pkg.go.dev/github.com/stretchr/testify).

First, install `testify`:

```shell
$ go get github.com/stretchr/testify
$ go get github.com/stretchr/testify/suite
```

In `services/service_test.go`, set up a `services_test` package. In this file, we are going to define the main test suite, which will later be shared by unit tests for all service types.

{{< gist "golang/029-service_test.go" "go" "services/service_test.go" >}}

This file makes use of the [github.com/stretchr/testify/suite](https://pkg.go.dev/github.com/stretchr/testify/suite) package, providing convenience functions to run tests in _test suites_ (pronounced like _test sweets_).

Using a test suite, our test examples can use setup and teardown callbacks.
In this example, we define a `SetupTestSuite()` method on the `ServiceTestSuite` type, and within that method, we connect to the database and clean the `users` table to ensure that we are starting with an empty database. The calls to `s.NoError(err)` use _test assertions_, which are convenience methods defined in the [github.com/stretchr/testify/assert](https://pkg.go.dev/github.com/stretchr/testify@v1.9.0/assert#Assertions) package. For instance, when calling the method [`NoError`](https://pkg.go.dev/github.com/stretchr/testify@v1.9.0/assert#Assertions.NoError) with an error value, we assert that no error was returned from a function. If an error is indeed returned, the test will fail, hopefully with a descriptive error message.

Note that we also need to define a regular Go test example, here named `TestServiceTestSuite`, serving as an entry point for the Go test runner.

Next, we can test our data validation and registration logic in a new file:

{{< gist "golang/030-user_service_test.go" "go" "services/user_service_test.go" >}}

### Making the tests pass

If you run the tests now, this test example is going to fail, because we have not implemented the registration logic yet:

```shell
$ make test
2024/09/17 00:31:16 goose: no migrations to run. current version: 20240913000048
go test -v ./...
?       github.com/moroz/webauthn-academy-go    [no test files]
?       github.com/moroz/webauthn-academy-go/db/queries [no test files]
=== RUN   TestServiceTestSuite
=== RUN   TestServiceTestSuite/TestRegisterUser
    user_service_test.go:20: 
                Error Trace:    /home/karol/working/webauthn/wip/services/user_service_test.go:20
                Error:          Expected value not to be nil.
                Test:           TestServiceTestSuite/TestRegisterUser
--- FAIL: TestServiceTestSuite (0.02s)
    --- FAIL: TestServiceTestSuite/TestRegisterUser (0.02s)
FAIL
FAIL    github.com/moroz/webauthn-academy-go/services   0.024s
FAIL
make: *** [Makefile:16: test] Error 1
```

We can make our initial test pass with the following implementation of `RegisterUser`:

{{< gist "golang/031-user_service.go" "go" "services/user_service.go" `{"linenostart":23}` >}}

This implementation satisfies our initial test conditions:

```shell
$ make test
2024/09/17 00:57:05 goose: no migrations to run. current version: 20240913000048
go test -v ./...
?       github.com/moroz/webauthn-academy-go    [no test files]
?       github.com/moroz/webauthn-academy-go/db/queries [no test files]
=== RUN   TestServiceTestSuite
=== RUN   TestServiceTestSuite/TestRegisterUser
--- PASS: TestServiceTestSuite (0.03s)
    --- PASS: TestServiceTestSuite/TestRegisterUser (0.03s)
PASS
ok      github.com/moroz/webauthn-academy-go/services   0.030s
```

However, there is a problem with this implementation: we are storing passwords in plain text. The [recommended way to store passwords](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html) in a database is to hash them using a dedicated password hashing algorithm. 
In the next subsection, we are going to modify our implementation to use [Argon2id](https://en.wikipedia.org/wiki/Argon2) instead.

### Hashing passwords using Argon2id

Before implementing password hashing, let us first add a test example.
Using the [`Regexp`](https://pkg.go.dev/github.com/stretchr/testify@v1.9.0/assert#Assertions.Regexp) matcher and a [regular expression](https://en.wikipedia.org/wiki/Regular_expression), we test if the `PasswordHash` field on the returned `queries.User` struct matches the [PHC string format](https://github.com/P-H-C/phc-string-format/blob/master/phc-sf-spec.md) for Argon2id:

{{< gist "golang/032-user_service_test.go" "go" "services/user_service_test.go" `{"linenostart":9}` >}}

This regular expression will only match if the string _starts with_ the substring `$argon2id$`.
We need to escape dollar signs as `\$` so that they are interpreted as literal dollar signs and not as "end of string."
Do note that we are passing the source for this regular expression as a raw string, surrounded with backticks (<code>&#96;</code>) rather than double quotes (<code>&#34;</code>). This way we do not need to escape backslashes (we can write <code>&#92;</code> instead of <code>&#92;&#92;</code>).

This test successfully fails:

```shell
$ make test
2024/09/17 01:36:39 goose: no migrations to run. current version: 20240913000048
go test -v ./...
?       github.com/moroz/webauthn-academy-go    [no test files]
?       github.com/moroz/webauthn-academy-go/db/queries [no test files]
=== RUN   TestServiceTestSuite
=== RUN   TestServiceTestSuite/TestRegisterUser
    user_service_test.go:23: 
                Error Trace:    /home/karol/working/webauthn/wip/services/user_service_test.go:23
                Error:          Expect "foobar123123" to match "^\$argon2id\$"
                Test:           TestServiceTestSuite/TestRegisterUser
--- FAIL: TestServiceTestSuite (0.02s)
    --- FAIL: TestServiceTestSuite/TestRegisterUser (0.02s)
FAIL
FAIL    github.com/moroz/webauthn-academy-go/services   0.026s
FAIL
make: *** [Makefile:16: test] Error 1
```

The first step to make it pass is to install `argon2id`:

```shell
$ go get github.com/alexedwards/argon2id
go: added github.com/alexedwards/argon2id v1.0.0
```

Then, in `RegisterUser`, hash the password before inserting it into the database:

{{< gist "golang/033-user_service.go" "go" "services/user_service.go" `{"linenostart":25}` >}}

The test should now be passing:

```shell
$ make test
2024/09/17 01:41:04 goose: no migrations to run. current version: 20240913000048
go test -v ./...
?       github.com/moroz/webauthn-academy-go    [no test files]
?       github.com/moroz/webauthn-academy-go/db/queries [no test files]
=== RUN   TestServiceTestSuite
=== RUN   TestServiceTestSuite/TestRegisterUser
--- PASS: TestServiceTestSuite (0.04s)
    --- PASS: TestServiceTestSuite/TestRegisterUser (0.04s)
PASS
ok      github.com/moroz/webauthn-academy-go/services   0.043s
```

### Validating inputs using `gookit/validate`

[`gookit/validate`](https://github.com/gookit/validate/) is a simple and customizable struct validation library. As is the case with many open source Go libraries, its documentation is written in [Chinglish](https://gookit.github.io/validate/#/) and it is easier to read if you understand [Chinese](https://gookit.github.io/validate/#/README.zh-CN).

Why not [`go-playground/validator`](https://pkg.go.dev/github.com/go-playground/validator), you may ask?

For reasons I cannot fathom, the Golang ecosystem has settled on `go-playground/validator` as the state of the art for struct validation.
I admit, this library has a considerable set of built-in validators, but even at version 10, its documentation and codebase is dreadful, full of cryptic logic, [tight coupling](https://github.com/go-playground/validator/blob/master/_examples/translations/main.go) to equally poorly documented libraries, and single-letter variable names. `gookit/validate`, on the other hand, is much simpler, and I do _not_ need to use a mysterious ["universal translator"](github.com/go-playground/universal-translator) library just to display a custom error message.

Start by installing `gookit/validate`:

```shell
$ go get github.com/gookit/validate
go: added github.com/gookit/filter v1.2.1
go: added github.com/gookit/goutil v0.6.15
go: added github.com/gookit/validate v1.5.2
```

Let's build a failing test to ensure that user registration fails if any of the parameters is an empty string:

{{< gist "golang/034-user_service_test.go" "go" "services/user_service_test.go" >}}

On the `NewUserParams` struct type, define annotations for [gookit/validate](https://github.com/gookit/validate):

{{< gist "golang/035-user_service.go" "go" "services/user_service.go" `{"linenostart":19}` >}}

With these changes, our tests for `required` validations should again be passing:

```shell
$ make test
2024/09/17 16:36:00 goose: no migrations to run. current version: 20240913000048
go test -v ./...
?       github.com/moroz/webauthn-academy-go    [no test files]
?       github.com/moroz/webauthn-academy-go/db/queries [no test files]
=== RUN   TestServiceTestSuite
=== RUN   TestServiceTestSuite/TestRegisterUser
=== RUN   TestServiceTestSuite/TestRegisterUserWithMissingAttributes
--- PASS: TestServiceTestSuite (0.07s)
    --- PASS: TestServiceTestSuite/TestRegisterUser (0.05s)
    --- PASS: TestServiceTestSuite/TestRegisterUserWithMissingAttributes (0.02s)
PASS
ok      github.com/moroz/webauthn-academy-go/services   (cached)
```

When running the test in [dlv](https://github.com/go-delve/delve), we can see the actual error and the corresponding error messages (see my [dotfiles](https://github.com/moroz/dotfiles/blob/master/nvim/after/plugin/keymappings.lua) for Neovim integration):

```shell
Type 'help' for list of commands.
Breakpoint 1 set at 0xd3cf7c for github.com/moroz/webauthn-academy-go/services_test.(*ServiceTestSuite).TestRegisterUserWithMissingAttributes() ./user_service_test.go:38
> [Breakpoint 1] github.com/moroz/webauthn-academy-go/services_test.(*ServiceTestSuite).TestRegisterUserWithMissingAttributes() ./user_service_test.go:38 (hits goroutine(146):1 total:1) (PC: 0xd3cf7c)
    33:
    34:         srv := services.NewUserService(s.db)
    35:
    36:         for _, params := range examples {
    37:                 user, err := srv.RegisterUser(context.Background(), params)
=>  38:                 s.Error(err)
    39:                 s.Nil(user)
    40:         }
    41: }
(dlv) p err
error(github.com/gookit/validate.Errors) [
        "Email": [
                "required": "Email is required to not be empty", 
        ], 
]
(dlv) 
```

The error message is technically correct, but the wording is strange. Let's update the tests to ensure that the error message is equal to `"can't be blank"`:

{{< gist "golang/037-user_service_test.go" "go" "services/user_service_test.go" `{"linenostart":27}` >}}



```plain
$ make test                                                                                                                          
2024/09/17 18:57:05 goose: no migrations to run. current version: 20240913000048                                                                              
go test -v ./...                                                                                                                                              
?       github.com/moroz/webauthn-academy-go    [no test files]                                                                                               
?       github.com/moroz/webauthn-academy-go/db/queries [no test files]                                                                                       
=== RUN   TestServiceTestSuite                                                                                                                                
=== RUN   TestServiceTestSuite/TestRegisterUser                                                                                                               
=== RUN   TestServiceTestSuite/TestRegisterUserWithMissingAttributes                                                                                          
    user_service_test.go:42:                                                                                                                                  
                Error Trace:    /home/karol/working/webauthn/wip/services/user_service_test.go:42                                                             
                Error:          Not equal: 
                                expected: "can't be blank"
                                actual  : "Email is required to not be empty"

# ... many similar errors below ...
```

{{< gist "golang/036-user_service.go" "go" "services/user_service.go" `{"linenostart":19}` >}}

Again, we can verify the changes in the debugger:

```shell
Type 'help' for list of commands.
Breakpoint 1 set at 0xd3d39c for github.com/moroz/webauthn-academy-go/services_test.(*ServiceTestSuite).TestRegisterUserWithMissingAttributes() ./user_service_test.go:38
> [Breakpoint 1] github.com/moroz/webauthn-academy-go/services_test.(*ServiceTestSuite).TestRegisterUserWithMissingAttributes() ./user_service_test.go:38 (hits goroutine(86):1 total:1) (PC: 0xd3d39c)
    33:
    34:         srv := services.NewUserService(s.db)
    35:
    36:         for _, params := range examples {
    37:                 user, err := srv.RegisterUser(context.Background(), params)
=>  38:                 s.Error(err)
    39:                 s.Nil(user)
    40:         }
    41: }
(dlv) p err
error(github.com/gookit/validate.Errors) [
        "Email": [
                "required": "can't be blank", 
        ], 
]
(dlv) 

```
