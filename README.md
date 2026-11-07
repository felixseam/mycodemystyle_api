# MycodeMystyle API

REST API for the Merl Rerng Cinema platform, built with Go and Fiber. One API serves two systems, the public **member** app (`front`) and the **admin** panel, each with its own users, tokens and middleware. Data is in PostgreSQL, cache in Tarantool, and every message is translated (English and Khmer).

## Tech stack

| Area           | Tool                                                                     |
| -------------- | ------------------------------------------------------------------------ |
| Language       | Go                                                                       |
| HTTP framework | [Fiber v2](https://github.com/gofiber/fiber)                             |
| Database       | PostgreSQL via [pgx v5](https://github.com/jackc/pgx) (`pgxpool`)        |
| Cache          | Tarantool                                                                |
| Auth           | JWT ([golang-jwt](https://github.com/golang-jwt/jwt)) + bcrypt passwords |
| Validation     | [go-playground/validator](https://github.com/go-playground/validator)    |
| i18n           | go-i18n through `fiberi18n` (English, Khmer)                             |
| Logging        | zerolog with a custom console format                                     |
| Migrations     | [goose](https://github.com/pressly/goose)                                |
| Live reload    | [air](https://github.com/air-verse/air)                                  |

## Requirements

- Go (see `go.mod` for the version)
- PostgreSQL
- Tarantool
- `make`

Install the dev tools (air and goose) once:

```bash
make tools
```

## Quick start

```bash
# 1. configure environment
cp .env.example .env        # then edit the values

# 2. start the cache
make tarantool-up           # terminal 1

# 3. prepare the database
make migrate-up

# 4. run the API with live reload
make dev                    # terminal 2
```

The API listens on `APP_HOST:APP_PORT` (default `0.0.0.0:8085`)

## Environment variables

Copy `.env.example` to `.env`. Never commit `.env`. Keep every line as `KEY=value`, with no spaces around `=`.

| Variable                                | Description                                                                              |
| --------------------------------------- | ---------------------------------------------------------------------------------------- |
| `APP_NAME`                              | Application name                                                                         |
| `APP_ENV`                               | `development` or `production`                                                            |
| `APP_HOST` / `APP_PORT`                 | Address the server binds to                                                              |
| `APP_TIMEZONE`                          | IANA timezone for timestamps written by the app (default `Asia/Phnom_Penh`)              |
| `JWT_SECRET_KEY`                        | Secret that signs **member** tokens. Use a long random value (`openssl rand -base64 48`) |
| `JWT_EXP_HOUR`                          | Member token lifetime in hours (default `7`)                                             |
| `JWT_SECRET_KEY_ADMIN`                  | Secret that signs **admin** tokens. Must differ from `JWT_SECRET_KEY`                    |
| `JWT_EXP_HOUR_ADMIN`                    | Admin token lifetime in hours (default `2`)                                              |
| `PG_DATABASE_URL`                       | PostgreSQL connection string                                                             |
| `PG_MAX_CONNS` / `PG_MIN_CONNS`         | Connection pool size                                                                     |
| `TARANTOOL_HOST` / `TARANTOOL_PORT`     | Tarantool address                                                                        |
| `TARANTOOL_USER` / `TARANTOOL_PASSWORD` | Tarantool credentials                                                                    |
| `TARANTOOL_POOL_SIZE`                   | Tarantool connection pool size                                                           |

## Make commands

Run `make help` to list everything.

| Command                               | Purpose                                        |
| ------------------------------------- | ---------------------------------------------- |
| `make dev`                            | Run with live reload (air)                     |
| `make run`                            | Run once                                       |
| `make build`                          | Build binary to `tmp/main`                     |
| `make test`                           | Run tests with the race detector               |
| `make cover`                          | Tests with coverage summary                    |
| `make fmt` / `make vet` / `make lint` | Format, static checks, lint                    |
| `make check`                          | fmt + vet + test (run before committing)       |
| `make tarantool-up`                   | Start Tarantool with data in `.data/tarantool` |
| `make tarantool-clean`                | Delete local Tarantool data                    |
| `make migrate-new name=xxx`           | Create a new migration                         |
| `make migrate-up` / `migrate-down`    | Apply / roll back migrations                   |
| `make migrate-status`                 | Show applied migrations                        |
| `make migrate-reset`                  | Roll back everything (dev only)                |
| `make docker-up` / `docker-down`      | Start / stop compose services                  |

## Project structure

```
.
├── main.go                     entry point: load .env, connect, build router, start
├── makefile
├── internal/
│   ├── apperr/                 typed errors with an HTTP status and a translation key
│   ├── cache/
│   │   ├── tarantool/          Tarantool connection and cache helpers
│   │   └── redis/              (reserved)
│   ├── config/                 app, postgres and tarantool settings
│   ├── constant/               status IDs, token audiences
│   ├── database/postgres/      pgx pool connection
│   ├── env/                    typed env helpers (String, Int, MustString, ...)
│   ├── i18n/                   translations (locales/en.json, km.json)
│   ├── logger/                 custom log format
│   ├── middleware/             JWT guards (member, admin) and the global error handler
│   ├── model/                  shared types (MemberContext, AdminContext)
│   ├── response/               the standard API response
│   ├── router/                 mounts every module and builds the middleware
│   ├── server/                 fiber app setup
│   ├── validator/              request validation
│   └── modules/                business features, grouped by audience
│       ├── front/              member system: auth, films, ...
│       └── admin/              admin system: auth, films, ...
├── migrations/                 SQL schema history (goose)
├── deploy/                     tarantool init script, postgres files, compose
└── .data/                      local runtime data (gitignored)
```

### Dependency rules

Imports only flow downward:

```
main.go -> router -> modules/* -> response, validator, apperr, i18n, logger, ...
```

1. Low-level packages (`apperr`, `config`, `constant`, `env`, `logger`, `model`) never import `modules/`, `router/` or `middleware/`.
2. `middleware` depends on interfaces (`MemberFinder`, `AdminFinder`), never on a specific module. The router passes in the repository.
3. A module never imports `router/`, and never imports another module's internals.
4. `front/` and `admin/` never import each other.

### Module layout

```
modules/front/films/
├── route.go           mounts the endpoints
├── model.go           DB structs, request/response types, bind()
├── errors.go          errors that only this module uses
├── handler.go         parse, validate, call service, respond
├── service.go         business rules
└── repository.go      SQL only
```

When a file passes about 300 lines or 10 methods, keep it as a skeleton (struct, constructor, interface) and move methods into per-operation files (`create_handler.go`, `create_service.go`, `create_repo.go`). If filenames start mixing several nouns, split them into separate modules.

Layer responsibilities:

- **Handler**: HTTP only. No SQL, no business rules. It picks the response message.
- **Service**: business rules. No `*fiber.Ctx`.
- **Repository**: SQL only, one method per query. It logs unexpected database errors once and returns `apperr.ErrInternal`.

Dependencies are passed in through constructors. A module takes only what it needs (`pool`, `cache`, time zone), never the whole app. Services depend on interfaces (`Authenticator`, `AuthRepo`), and constructors return the concrete pointer.

## Authentication

Two separate systems share the middleware code but nothing else:

| System | Table         | Secret                 | Token `aud` | Guard                     |
| ------ | ------------- | ---------------------- | ----------- | ------------------------- |
| Member | `tbl_members` | `JWT_SECRET_KEY`       | `member`    | `middleware.NewMemberJWT` |
| Admin  | `tbl_users`   | `JWT_SECRET_KEY_ADMIN` | `admin`     | `middleware.NewAdminJWT`  |

A member token is rejected on admin routes (and the reverse) because the audience and the signing secret differ.

Send the token on every protected request:

```
Authorization: Bearer <access_token>
```

On each request the guard checks signature, expiry and audience, then loads the user and checks that:

- the `login_session` in the token matches the database (a new login on another device invalidates the old token),
- the account is still active.

Handlers read the current user with `middleware.CurrentMember(c)` or `middleware.CurrentAdmin(c)`. Admin routes can add `middleware.RequireRole(...)`.

Passwords are stored as bcrypt hashes. Generate one for a seed user with:

```bash
htpasswd -bnBC 10 "" 'YourStrongPassword' | tr -d ':\n' | sed 's/^\$2y/$2a/'
```

### Login

```bash
curl -X POST localhost:8085/api/v1/front/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"user_name":"FELIX007","password":"YourStrongPassword"}'
```

## API response format

Every response uses the same envelope:

```json
{
  "success": false,
  "message": "Login failed",
  "data": null,
  "errors": [
    {
      "field": "password",
      "code": "wrong_password",
      "message": "Incorrect password"
    }
  ]
}
```

- `message` is the general result of the action, chosen by the handler.
- `errors[]` says what exactly went wrong. `code` is stable and safe to compare in the frontend, `message` is translated.
- `data` is `null` on errors and `errors` is `null` on success.

Three functions in `internal/response`:

| Function                            | Use it when                                                                |
| ----------------------------------- | -------------------------------------------------------------------------- |
| `Success(c, status, message, data)` | the request worked                                                         |
| `FailErr(c, message, err)`          | you have an `error`: the status and the `errors[]` array are built for you |
| `Fail(c, status, message, errs)`    | you want to write the `[]ErrorDetail` by hand                              |

```go
res, err := h.service.Login(c.UserContext(), req)
if err != nil {
	return response.FailErr(c, i18n.Translate(c, "login_failed"), err)
}
return response.Success(c, http.StatusOK, i18n.Translate(c, "login_success"), res)
```

Validation errors list every invalid field, and the message has no field name because `field` is its own property:

```json
{
  "success": false,
  "message": "Validation failed",
  "data": null,
  "errors": [
    {
      "field": "user_name",
      "code": "required",
      "message": "This field is required"
    },
    {
      "field": "password",
      "code": "min_length",
      "message": "Must be at least 6 characters"
    }
  ]
}
```

## Errors

Services and repositories return `apperr` errors, never plain `fmt.Errorf` text for something the client should understand. A plain error is treated as unknown: the client sees a translated `internal_error` and the real cause is written to the log.

```go
// shared errors live in internal/apperr
return nil, apperr.ErrNotFound

// module errors live in the module's errors.go
var ErrWrongPassword = apperr.New(http.StatusUnauthorized, "wrong_password").OnField("password")

// keep the real cause for the log, show only internal_error to the client
return nil, apperr.ErrInternal.Wrap(err)
```

To add a new error: create it with `apperr.New(status, "code")` and add `code` to **both** locale files.

Errors returned by handlers or middleware (unknown route, rate limit, token errors) go through the global `middleware.ErrorHandler`, so they use the same response shape.

## Validation

Request structs carry `validate` tags and a `bind` method that parses, cleans and validates:

```go
type LoginRequest struct {
	UserName string `json:"user_name" validate:"required,max=50"`
	Password string `json:"password"  validate:"required,min=6,max=100"`
}

func (lr *LoginRequest) bind(c *fiber.Ctx) error {
	if err := c.BodyParser(lr); err != nil {
		return apperr.ErrInvalidJSON
	}
	lr.UserName = strings.TrimSpace(lr.UserName)
	return validator.Validate(lr)
}
```

`validator.Validate` returns an `apperr` error listing every invalid field. Message keys are `validation_<code>` (for example `validation_required`, `validation_min_length`).

## Translations

Messages live in `internal/i18n/locales/en.json` and `km.json`. The language comes from the `Accept-Language` header, with English as the default.

- Every message ID must exist in **both** files.
- `i18n.Translate(c, "key")` and `i18n.TranslateWith(c, "key", data)` never panic: a missing key returns the key itself and logs one warning.
- `i18n.Lookup` tells you whether a key exists.
- `i18n.New()` must be registered before any route that translates.

## Configuration rules

- `.env` is loaded **once**, at the top of `main`.
- Read values with the `env` package (`env.String`, `env.Int`, `env.MustString`) in constructors or at startup, never inside a request handler.
- `env.MustString` stops the app at startup when a required value is missing.
- Timestamps written by the app use the `APP_TIMEZONE` location: the service creates `now` and passes it to the repository.

## Database migrations

```bash
make migrate-new name=create_films     # creates migrations/0000X_create_films.sql
make migrate-up
make migrate-status
```

Current tables: `tbl_users` (admin) and `tbl_members` (member).

Rules:

1. Never edit a migration that has already been applied on a shared environment. Add a new one.
2. One change per file, with a clear name.
3. Always write the `-- +goose Down` part.
4. Run migrations as a separate step, not automatically on API startup.
5. Seed users need a bcrypt hash in `password`, never plain text.

## Conventions

- Package names are short, lowercase and singular (`film`, not `films`), except where a folder name is already established.
- `ctx context.Context` is the first parameter of every service and repository method.
- Services return `apperr` errors. Handlers pass them to `response.FailErr`.
- The repository logs unexpected database errors once, with the method name, so each failure appears once in the log.
- Never log passwords or whole request structs.
- Put a helper in its own small package only once a second module needs it.

## Git hygiene

Make sure `.gitignore` contains:

```
.env
tmp/
.data/
*.snap
*.xlog
coverage.out
```

Commit `.env.example` with fake values so new machines know which variables exist.
