package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"mycodemystyle_api/internal/apperr"
)

// identity is what a valid token tells us. The database check comes next.
type identity struct {
	UserUUID string
	Session  string
	Exp      time.Time
}

// authenticate reads the token and verifies signature, expiry and audience.
func authenticate(c *fiber.Ctx, secret []byte, audience string) (*identity, error) {
	isWS := strings.EqualFold(c.Get("Upgrade"), "websocket")

	tokenString, err := extractToken(c, isWS)
	if err != nil {
		return nil, err
	}

	claims, err := parseToken(tokenString, secret, audience)
	if err != nil {
		return nil, err
	}

	userUUID, _ := claims["user_uuid"].(string)
	if userUUID == "" {
		return nil, apperr.MissingKey("user_uuid")
	}
	session, _ := claims["login_session"].(string)
	if session == "" {
		return nil, apperr.MissingKey("login_session")
	}
	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, apperr.MissingKey("exp")
	}

	return &identity{UserUUID: userUUID, Session: session, Exp: time.Unix(int64(exp), 0)}, nil
}

func extractToken(c *fiber.Ctx, isWS bool) (string, error) {
	if isWS {
		// browsers cannot set headers on WebSocket: "Bearer, <token>" in the protocol header
		parts := strings.Split(c.Get("Sec-WebSocket-Protocol"), ",")
		if len(parts) != 2 || strings.TrimSpace(parts[0]) != "Bearer" {
			return "", apperr.ErrMissingToken
		}
		return strings.TrimSpace(parts[1]), nil
	}

	parts := strings.Fields(c.Get(fiber.HeaderAuthorization))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", apperr.ErrMissingToken
	}
	return parts[1], nil
}

func parseToken(tokenString string, secret []byte, audience string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(
		tokenString,
		func(t *jwt.Token) (interface{}, error) { return secret, nil },
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithExpirationRequired(),
		jwt.WithAudience(audience), // a member token fails on admin routes, and the reverse
	)
	if err != nil || !token.Valid {
		return nil, apperr.ErrInvalidToken
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, apperr.ErrInvalidToken
	}
	return claims, nil
}
