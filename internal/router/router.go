package router

import (
	"mycodemystyle_api/internal/cache/tarantool"
	admin_auth "mycodemystyle_api/internal/modules/admin/auth"
	admin_realtime "mycodemystyle_api/internal/modules/admin/realtime"
	"mycodemystyle_api/internal/modules/front/auth"
	"mycodemystyle_api/internal/modules/front/realtime"
	"mycodemystyle_api/internal/ws"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Router struct {
	app  *fiber.App
	deps *AppDependencies
}

type AppDependencies struct {
	PgDBPool       *pgxpool.Pool
	TarantoolCache *tarantool.Cache
	TimeZone       *time.Location
	FrontHub       *ws.Hub
	AdminHub       *ws.Hub
}

func New(app *fiber.App, deps *AppDependencies) *Router {
	return &Router{
		app:  app,
		deps: deps,
	}
}

func (r *Router) Register() {
	// API modules
	v1 := r.app.Group("/api/v1")

	//// Front API
	front := v1.Group("/front")
	// Auth
	auth.NewRoute(front, r.deps.PgDBPool, r.deps.TarantoolCache, r.deps.TimeZone).RegisterAuthRoute()
	// Realtime
	realtime.NewRealtimeRoute(front, r.deps.FrontHub, r.deps.PgDBPool).RegisterRealtime()

	//// Admin API
	admin := v1.Group("/admin")
	// Auth
	admin_auth.NewRoute(admin, r.deps.PgDBPool, r.deps.TarantoolCache, r.deps.TimeZone).RegisterAuthRoute()
	// Realtime
	admin_realtime.NewRealtimeRoute(admin, r.deps.AdminHub, r.deps.PgDBPool).RegisterRealtime()

}
