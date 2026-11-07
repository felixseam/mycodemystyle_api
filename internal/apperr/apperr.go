package apperr

import "net/http"

// FieldError is one invalid input field. It is translated later, in the
// response package, so validation code needs no i18n or HTTP.
type FieldError struct {
	Field string // "user_name"
	Code  string // "required", "min_length", ... (message key is "validation_" + Code)
	Param string // rule parameter, e.g. "6" for min=6
}

// Error is the one error type the app returns to clients.
// The response package turns it into []ErrorDetail.
type Error struct {
	Status int                    // HTTP status
	Code   string                 // machine code AND i18n message ID
	Field  string                 // single form field this error belongs to (optional)
	Fields []FieldError           // all invalid fields (validation only)
	Data   map[string]interface{} // template data for the message (optional)
	Err    error                  // underlying cause: logged, never shown to the client
}

func New(status int, code string) *Error {
	return &Error{Status: status, Code: code}
}

func (e *Error) Error() string {
	if e.Err != nil {
		return e.Code + ": " + e.Err.Error()
	}
	return e.Code
}

func (e *Error) Unwrap() error { return e.Err }

// Is matches by code, so errors.Is still works on copies made by With/OnField/Wrap.
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	return ok && t.Code == e.Code
}

// The three methods below return a copy, so the shared errors below are never changed.

// OnField ties the error to a form field, e.g. "user_name".
func (e *Error) OnField(field string) *Error {
	c := *e
	c.Field = field
	return &c
}

// With adds template data for the message, e.g. {"Key": "user_uuid"}.
func (e *Error) With(data map[string]interface{}) *Error {
	c := *e
	c.Data = data
	return &c
}

// Wrap remembers the real cause for the logs.
func (e *Error) Wrap(err error) *Error {
	c := *e
	c.Err = err
	return &c
}

// Validation builds the 422 error that carries every invalid field.
func Validation(fields []FieldError) *Error {
	return &Error{
		Status: http.StatusUnprocessableEntity,
		Code:   "validation_failed",
		Fields: fields,
	}
}

// general
var (
	ErrBadRequest      = New(http.StatusBadRequest, "bad_request")
	ErrInvalidJSON     = New(http.StatusBadRequest, "invalid_json")
	ErrUnauthorized    = New(http.StatusUnauthorized, "unauthorized")
	ErrForbidden       = New(http.StatusForbidden, "forbidden")
	ErrNotFound        = New(http.StatusNotFound, "not_found")
	ErrConflict        = New(http.StatusConflict, "conflict")
	ErrTooMany         = New(http.StatusTooManyRequests, "too_many_requests")
	ErrInternal        = New(http.StatusInternalServerError, "internal_error")
	ErrUpgradeRequired = New(http.StatusUpgradeRequired, "upgrade_required")
)

// router level (used by the global ErrorHandler)
var (
	ErrRouteNotFound    = New(http.StatusNotFound, "route_not_found")
	ErrMethodNotAllowed = New(http.StatusMethodNotAllowed, "method_not_allowed")
	ErrPayloadTooLarge  = New(http.StatusRequestEntityTooLarge, "payload_too_large")
)

// token and session (used by the JWT middleware)
var (
	ErrMissingToken   = New(http.StatusUnauthorized, "missing_or_malformed_jwt")
	ErrInvalidToken   = New(http.StatusUnauthorized, "invalid_jwt_token")
	ErrSessionExpired = New(http.StatusUnauthorized, "session_expired")
	ErrUserLookup     = New(http.StatusUnauthorized, "get_userinfo_failed")
	errMissingKey     = New(http.StatusUnauthorized, "missing_or_invalid_key")
)

// MissingKey reports a required JWT claim that is absent or invalid.
func MissingKey(key string) *Error {
	return errMissingKey.With(map[string]interface{}{"Key": key})
}
