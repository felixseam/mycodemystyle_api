package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"mycodemystyle_api/internal/apperr"
	"mycodemystyle_api/internal/cache/tarantool"
	"mycodemystyle_api/internal/constant"
	"mycodemystyle_api/internal/env"
	"mycodemystyle_api/internal/logger"
)

type Authenticator interface {
	Login(ctx context.Context, lr LoginRequest) (*LoginResponse, error)
}

type AuthService struct {
	pool      *pgxpool.Pool
	cache     *tarantool.Cache
	timeZone  *time.Location
	repo      AuthRepo
	tz        *time.Location
	jwtSecret []byte
	jwtTTL    time.Duration
}

func NewService(pool *pgxpool.Pool, cache *tarantool.Cache, timeZone *time.Location) *AuthService {
	tz := timeZone
	if tz == nil {
		tz = time.UTC
	}
	return &AuthService{
		repo:      NewRepoImpl(pool),
		cache:     cache,
		tz:        tz,
		jwtSecret: []byte(env.MustString("JWT_SECRET_KEY")),
		jwtTTL:    time.Duration(env.Int("JWT_EXP_HOUR", 7)) * time.Hour,
	}
}

func (s *AuthService) Login(ctx context.Context, lr LoginRequest) (*LoginResponse, error) {
	user, err := s.repo.GetByUserName(ctx, lr.UserName)
	if errors.Is(err, ErrUserNotFound) {
		logger.NewCustomLog("login_failed", "unknown user: "+lr.UserName, "warn")
		return nil, ErrUserNameNotFound
	}
	if err != nil {
		return nil, err
	}

	if user.Password != lr.Password {
		logger.NewCustomLog("login_failed", "wrong password for: "+lr.UserName, "warn")
		return nil, ErrWrongPassword
	}

	if user.StatusID != 1 {
		logger.NewCustomLog("login_failed", "disabled account: "+lr.UserName, "warn")
		return nil, ErrAccountDisabled
	}

	now := time.Now().In(s.tz) // one "now", in the app timezone

	session := uuid.NewString()
	if err := s.repo.UpdateLogin(ctx, user.ID, session, now); err != nil {
		return nil, err
	}

	claims := jwt.MapClaims{
		"id":            user.ID,
		"user_uuid":     user.UserUUID,
		"username":      user.UserName,
		"role_id":       user.RoleID,
		"login_session": session,
		"aud":           constant.AudienceUser,
		"exp":           now.Add(s.jwtTTL).Unix(),
	}

	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.jwtSecret)
	if err != nil {
		logger.NewCustomLog("login_error", "sign token: "+err.Error(), "error")
		return nil, apperr.ErrInternal
	}

	return &LoginResponse{Auth: Auth{
		AccessToken: tokenString,
		TokenType:   "jwt",
	}}, nil
}
