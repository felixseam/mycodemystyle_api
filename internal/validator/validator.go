package validator

import (
	"errors"
	"mycodemystyle_api/internal/apperr"
	"reflect"
	"regexp"
	"strings"
	"sync"

	gv "github.com/go-playground/validator/v10"
)

var (
	once sync.Once
	v    *gv.Validate

	usernameRe = regexp.MustCompile(`^[A-Za-z0-9_]{3,50}$`)
	phoneRe    = regexp.MustCompile(`^\+?[0-9]{8,15}$`)
)

func instance() *gv.Validate {
	once.Do(func() {
		v = gv.New(gv.WithRequiredStructEnabled())

		// report field names the way the client sent them
		v.RegisterTagNameFunc(func(f reflect.StructField) string {
			for _, tag := range []string{"json", "query"} {
				name := strings.SplitN(f.Tag.Get(tag), ",", 2)[0]
				if name != "" && name != "-" {
					return name
				}
			}
			return f.Name
		})

		_ = v.RegisterValidation("username", func(fl gv.FieldLevel) bool {
			return usernameRe.MatchString(fl.Field().String())
		})
		_ = v.RegisterValidation("phone", func(fl gv.FieldLevel) bool {
			return phoneRe.MatchString(fl.Field().String())
		})
	})
	return v
}

// Validate returns nil when s is valid, otherwise an *apperr.Error
// listing every invalid field.
func Validate(s any) error {
	err := instance().Struct(s)
	if err == nil {
		return nil // a plain nil, never a typed nil pointer
	}

	var verrs gv.ValidationErrors
	if !errors.As(err, &verrs) {
		return apperr.ErrInternal.Wrap(err) // programmer error, e.g. a non-struct was passed
	}

	fields := make([]apperr.FieldError, 0, len(verrs))
	for _, fe := range verrs {
		fields = append(fields, apperr.FieldError{
			Field: fieldPath(fe),
			Code:  ruleCode(fe),
			Param: fe.Param(),
		})
	}
	return apperr.Validation(fields)
}

// "LoginRequest.items[0].name" -> "items[0].name"
func fieldPath(fe gv.FieldError) string {
	ns := fe.Namespace()
	if i := strings.Index(ns, "."); i >= 0 {
		return ns[i+1:]
	}
	return ns
}

// min/max/len mean different things for text and numbers, so text gets its own codes.
func ruleCode(fe gv.FieldError) string {
	if fe.Kind() == reflect.String {
		switch fe.Tag() {
		case "min":
			return "min_length"
		case "max":
			return "max_length"
		case "len":
			return "exact_length"
		}
	}
	return fe.Tag()
}
