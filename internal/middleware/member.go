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

const MemberContextKey = "MemberContext"

// MemberFinder is implemented by the front auth repository (tbl_members).
// It returns apperr.ErrNotFound when the member does not exist.
type MemberFinder interface {
	FindByUUID(ctx context.Context, memberUUID string) (*model.MemberContext, error)
}

func NewMemberJWT(finder MemberFinder) fiber.Handler {
	secret := []byte(env.MustString("JWT_SECRET_KEY"))

	return func(c *fiber.Ctx) error {
		id, err := authenticate(c, secret, constant.AudienceMember)
		if err != nil {
			return err
		}
		if err := handleMemberContext(c, id, finder); err != nil {
			return err
		}
		return c.Next()
	}
}

func handleMemberContext(c *fiber.Ctx, id *identity, finder MemberFinder) error {
	m, err := finder.FindByUUID(c.UserContext(), id.UserUUID)
	if errors.Is(err, apperr.ErrNotFound) {
		return apperr.ErrSessionExpired // member was deleted
	}
	if err != nil {
		return err // a real failure (database down): the repo already logged it
	}

	// the member logged in somewhere else, so this token is no longer valid
	if id.Session != m.LoginSession {
		return apperr.ErrSessionExpired
	}

	// disabled after the token was issued
	if m.StatusID != 1 {
		return apperr.ErrForbidden
	}

	m.Exp = id.Exp
	m.UserAgent = c.Get(fiber.HeaderUserAgent)
	m.IP = c.IP()

	c.Locals(MemberContextKey, *m)
	return nil
}

// CurrentMember returns the member stored by NewMemberJWT for this request.
func CurrentMember(c *fiber.Ctx) (model.MemberContext, bool) {
	m, ok := c.Locals(MemberContextKey).(model.MemberContext)
	return m, ok
}
