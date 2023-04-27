# 👨🏻‍⚕️ `Bones` - An easy-peasy API squeezy. [WIP]

## ⚠️👷‍♂️🚧 Under construction! Please don't use this library until v1.0.0

👨🏻‍⚕️ `Bones` - Is a set of well-tested reusable components to speedup dat-to-day development of RESTful APIs.

#### Also

👨🏻‍⚕️ _`Bones` - Dr. Leonard H. McCoy, known as "Bones", is a character in science-fiction TV series Star Trek._

## Packages

- [`servekit`](servekit/listener.go) - Holds an HTTP server, which is just a thin wrapper around great router `github.com/go-chi/chi/v5`.
    - [`respond`](servekit/respond/respond.go) - Holds set of usefully functions to respond to an HTTP request with a proper status code and a body as JSON,
      plain text, and error.
    - [`middleware`](servekit/middleware/middleware.go) - 
- [`errkit`](errkit/errors.go) - Holds set of predefined sentinel errors for the common cases.
- [`idkit`](idkit/id.go) - Holds set of functions which generates and validates different kind of identifiers.
- `dbkit` - Holds database related utils and wrappers.
    - [`pgconn`](dbkit/pgconn/postgres.go) - Tiny wrapper around `github.com/jackc/pgx/v4` to work with Postgres.
    - [`pgmigrate`](dbkit/pgmigrate/migrator.go) - Tiny wrapper around `github.com/jackc/tern` to work with database schema migrations.

## On the shoulders of giants

- [github.com/go-chi/chi](https://github.com/go-chi/chi)
- [github.com/jackc/pgx](https://github.com/jackc/pgx)