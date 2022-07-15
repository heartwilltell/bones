# 👨🏻‍⚕️ `Bones` - The set of reusable components to speedup dat-to-day development.

_`Bones` - Dr. Leonard H. McCoy, known as "Bones", is a character in science-fiction TV series Star Trek._

## Table of content

- `db` - Holds database related utils and wrappers.
    - [`pgconn`](db/pgconn/postgres.go) - Tiny wrapper around `github.com/jackc/pgx/v4` to work with Postgres.
    - `pgmigrate` - Tiny wrapper around `github.com/jackc/tern` to work with database schema migrations.
- `id` - Holds set of functions which generates and validates different kind of identifiers