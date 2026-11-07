package middleware

import (
	"errors"
	"mycodemystyle_api/internal/apperr"
	"mycodemystyle_api/internal/i18n"
	"mycodemystyle_api/internal/response"

	"net/http"

	"github.com/gofiber/fiber/v2"
)

func ErrorHandler(c *fiber.Ctx, err error) error {
	msg := i18n.Translate(c, "request_failed")

	var fe *fiber.Error
	if errors.As(err, &fe) {
		switch fe.Code {
		case http.StatusNotFound:
			return response.FailErr(c, msg, apperr.ErrRouteNotFound)
		case http.StatusMethodNotAllowed:
			return response.FailErr(c, msg, apperr.ErrMethodNotAllowed)
		case http.StatusRequestEntityTooLarge:
			return response.FailErr(c, msg, apperr.ErrPayloadTooLarge)
		case http.StatusBadRequest:
			return response.FailErr(c, msg, apperr.ErrBadRequest)
		}
	}
	return response.FailErr(c, msg, err)
}
