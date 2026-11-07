package auth

import (
	"mycodemystyle_api/internal/cache/tarantool"
	"mycodemystyle_api/internal/i18n"
	"mycodemystyle_api/internal/response"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthHandler struct {
	service Authenticator
}

func NewHandler(pool *pgxpool.Pool, cache *tarantool.Cache, timeZone *time.Location) *AuthHandler {
	return &AuthHandler{
		service: NewService(pool, cache, timeZone),
	}
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := req.bind(c); err != nil {
		return response.FailErr(c, i18n.Translate(c, "validation_failed"), err)
	}
	// apple := []int{}
	// appleone := apple[0]
	// fmt.Println(appleone)

	res, err := h.service.Login(c.UserContext(), req)
	if err != nil {
		return response.FailErr(c, i18n.Translate(c, "login_failed"), err)
	}

	return response.Success(c, http.StatusOK, i18n.Translate(c, "login_success"), res)
}
