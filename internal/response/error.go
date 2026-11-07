package response

import (
	"errors"
	"mycodemystyle_api/internal/apperr"
	"mycodemystyle_api/internal/i18n"
	"mycodemystyle_api/internal/logger"

	"github.com/gofiber/fiber/v2"
)

// FailErr builds the ErrorDetail array from an error and sends it.
// The status comes from the error, and the message is yours.
func FailErr(c *fiber.Ctx, message string, err error) error {
	ae := toAppError(err)

	if ae.Err != nil {
		logger.NewCustomLog("api-error", ae.Err.Error(), "error")
	}

	return Fail(c, ae.Status, message, buildDetails(c, ae))
}

func toAppError(err error) *apperr.Error {
	if err == nil {
		return apperr.ErrInternal
	}
	var ae *apperr.Error
	if errors.As(err, &ae) {
		return ae
	}
	// plain fmt.Errorf, pgx, a bug: the client sees internal_error
	return apperr.ErrInternal.Wrap(err)
}

func buildDetails(c *fiber.Ctx, ae *apperr.Error) []ErrorDetail {
	// a normal error: one entry
	if len(ae.Fields) == 0 {
		return []ErrorDetail{{
			Field:   ae.Field,
			Code:    ae.Code,
			Message: i18n.TranslateWith(c, ae.Code, ae.Data),
		}}
	}

	// validation: one entry per invalid field
	out := make([]ErrorDetail, 0, len(ae.Fields))
	for _, f := range ae.Fields {
		msg, ok := i18n.Lookup(c, "validation_"+f.Code, map[string]interface{}{"Param": f.Param})
		if !ok { // no message for this rule: use the generic one
			msg = i18n.Translate(c, "validation_invalid")
		}
		out = append(out, ErrorDetail{Field: f.Field, Code: f.Code, Message: msg})
	}
	return out
}
