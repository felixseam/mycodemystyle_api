package auth

import (
	"mycodemystyle_api/internal/cache/tarantool"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthRoute struct {
	app      fiber.Router
	pool     *pgxpool.Pool
	cache    *tarantool.Cache
	timeZone *time.Location
	handler  *AuthHandler
}

func NewRoute(app fiber.Router, pool *pgxpool.Pool, cache *tarantool.Cache, timeZone *time.Location) *AuthRoute {
	return &AuthRoute{
		app:     app,
		handler: NewHandler(pool, cache, timeZone),
	}
}

func (r *AuthRoute) RegisterAuthRoute() *AuthRoute {
	auth := r.app.Group("/auth")
	auth.Post("/login", r.handler.Login)
	return r
}
