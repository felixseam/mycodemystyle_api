package realtime

import (
	"mycodemystyle_api/internal/apperr"
	"mycodemystyle_api/internal/i18n"
	"mycodemystyle_api/internal/middleware"
	"mycodemystyle_api/internal/model"
	"mycodemystyle_api/internal/response"
	"mycodemystyle_api/internal/ws"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

type RealtimeHandler struct {
	hub *ws.Hub
}

func NewRealtimeHandler(hub *ws.Hub) *RealtimeHandler {
	return &RealtimeHandler{hub: hub}
}

func (h *RealtimeHandler) RequireUpgrade(c *fiber.Ctx) error {
	if !websocket.IsWebSocketUpgrade(c) {
		return response.FailErr(c, i18n.Translate(c, "request_failed"), apperr.ErrUpgradeRequired)
	}

	return c.Next()
}

func (h *RealtimeHandler) Serve(conn *websocket.Conn) {
	m, ok := conn.Locals(middleware.UserContextKey).(model.UserContext)
	if !ok {
		_ = conn.Close()
		return
	}

	h.hub.Serve(conn, ws.User{
		ID:       m.ID,
		UUID:     m.UserUUID,
		UserName: m.UserName,
	})
}
