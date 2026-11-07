package middleware

import (
	"context"
	"errors"
	"mycodemystyle_api/internal/apperr"
	"mycodemystyle_api/internal/constant"
	"mycodemystyle_api/internal/env"
	"mycodemystyle_api/internal/model"

	"github.com/gofiber/fiber/v2"
)

const UserContextKey = "UserContext"

// MemberFinder is implemented by the front auth repository (tbl_members).
// It returns apperr.ErrNotFound when the member does not exist.
type UserFinder interface {
	FindByUUID(ctx context.Context, userUUID string) (*model.UserContext, error)
}

func NewUserJWT(finder UserFinder) fiber.Handler {
	secret := []byte(env.MustString("JWT_SECRET_KEY"))

	return func(c *fiber.Ctx) error {
		id, err := authenticate(c, secret, constant.AudienceUser)
		if err != nil {
			return err
		}
		if err := handleUserContext(c, id, finder); err != nil {
			return err
		}
		return c.Next()
	}
}

func handleUserContext(c *fiber.Ctx, id *identity, finder UserFinder) error {
	u, err := finder.FindByUUID(c.UserContext(), id.UserUUID)
	if errors.Is(err, apperr.ErrNotFound) {
		return apperr.ErrSessionExpired // member was deleted
	}
	if err != nil {
		return err // a real failure (database down): the repo already logged it
	}

	// the member logged in somewhere else, so this token is no longer valid
	if id.Session != u.LoginSession {
		return apperr.ErrSessionExpired
	}
	if u.RoleID == 0 {

		return apperr.ErrForbidden
	}

	// disabled after the token was issued
	if u.StatusID != 1 {
		return apperr.ErrForbidden
	}

	u.Exp = id.Exp
	u.UserAgent = c.Get(fiber.HeaderUserAgent)
	u.IP = c.IP()

	c.Locals(UserContextKey, *u)
	return nil
}

// CurrentMember returns the member stored by NewMemberJWT for this request.
func CurrentUser(c *fiber.Ctx) (model.UserContext, bool) {
	m, ok := c.Locals(UserContextKey).(model.UserContext)
	return m, ok
}
