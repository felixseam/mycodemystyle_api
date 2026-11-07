package realtime

import (
	"mycodemystyle_api/internal/middleware"
	"mycodemystyle_api/internal/modules/front/auth"
	"mycodemystyle_api/internal/ws"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RealtimeRoute struct {
	app     fiber.Router
	pool    *pgxpool.Pool
	handler *RealtimeHandler
}

func NewRealtimeRoute(app fiber.Router, hub *ws.Hub, pool *pgxpool.Pool) *RealtimeRoute {
	return &RealtimeRoute{
		app:     app,
		pool:    pool,
		handler: NewRealtimeHandler(hub),
	}
}

func (r *RealtimeRoute) RegisterRealtime() *RealtimeRoute {
	r.app.Get(
		"/ws",
		middleware.NewMemberJWT(auth.NewRepoImpl(r.pool)),
		r.handler.RequireUpgrade,
		websocket.New(r.handler.Serve,
			websocket.Config{
				Subprotocols: []string{"Bearer"},
			},
		),
	)

	return r

}
