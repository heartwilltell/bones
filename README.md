# 👨🏻‍⚕️ `Bones` - An easy-peasy API squeezy.

👨🏻‍⚕️ `Bones` - Is a set of well-tested reusable components to speedup dat-to-day development of RESTful APIs.

#### Also

👨🏻‍⚕️ _`Bones` - Dr. Leonard H. McCoy, known as "Bones", is a character in science-fiction TV series Star Trek._

## Packages
The root package holds an HTTP server, which is just a thin wrapper around great router `github.com/go-chi/chi/v5`.

- [`middleware`](middleware/middleware.go) - Holds a set of HTTP server middlewares.
- `db` - Holds database related utils and wrappers.
    - [`pgconn`](db/pgconn/postgres.go) - Tiny wrapper around `github.com/jackc/pgx/v4` to work with Postgres.
    - [`pgmigrate`](db/pgmigrate/migrator.go) - Tiny wrapper around `github.com/jackc/tern` to work with database schema migrations.
- [`id`](id/id.go) - Holds set of functions which generates and validates different kind of identifiers.

## On the shoulders of giants
 - github.com/go-chi/chi
 - 