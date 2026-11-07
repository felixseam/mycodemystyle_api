package auth

import (
	"errors"
	"mycodemystyle_api/internal/apperr"
	"net/http"
)

var ErrMemberNotFound = errors.New("member not found")

var (
	ErrUserNameNotFound   = apperr.New(http.StatusUnauthorized, "username_not_found").OnField("user_name")
	ErrWrongPassword      = apperr.New(http.StatusUnauthorized, "wrong_password").OnField("password")
	ErrInvalidCredentials = apperr.New(http.StatusUnauthorized, "invalid_credentials")
	ErrAccountDisabled    = apperr.New(http.StatusForbidden, "account_disabled")
)

var (
	ErrUserNameTaken    = apperr.New(http.StatusConflict, "username_taken")
	ErrWrongOldPassword = apperr.New(http.StatusUnprocessableEntity, "wrong_old_password")
	ErrSamePassword     = apperr.New(http.StatusUnprocessableEntity, "same_password")
)
