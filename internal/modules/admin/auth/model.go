package auth

import (
	"mycodemystyle_api/internal/apperr"
	"mycodemystyle_api/internal/validator"

	"github.com/gofiber/fiber/v2"
)

type LoginRequest struct {
	UserName string `json:"user_name" validate:"required,max=50"`
	Password string `json:"password"  validate:"required,min=6,max=100"`
}

func (lr *LoginRequest) bind(c *fiber.Ctx) error {
	if err := c.BodyParser(lr); err != nil {
		return apperr.ErrInvalidJSON
	}

	return validator.Validate(lr)
}

type LoginResponse struct {
	Auth Auth `json:"auth"`
}

type Auth struct {
	TokenType   string `json:"token_type"`
	AccessToken string `json:"access_token"`
}

type User struct {
	ID           int    `db:"id"`
	UserUUID     string `db:"user_uuid"`
	UserName     string `db:"user_name"`
	Password     string `db:"password"`
	LoginSession string `db:"login_session"`
	StatusID     int16  `db:"status_id"`
	RoleID       int    `db:"role_id"`
}
