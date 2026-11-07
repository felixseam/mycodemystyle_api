package i18n

import (
	"encoding/json"

	fiberi18n "github.com/gofiber/contrib/fiberi18n/v2"
	"github.com/gofiber/fiber/v2"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

func New() fiber.Handler {
	return fiberi18n.New(&fiberi18n.Config{
		RootPath:         "./internal/i18n/locales",
		AcceptLanguages:  []language.Tag{language.English, language.Khmer},
		DefaultLanguage:  language.English,
		FormatBundleFile: "json",
		UnmarshalFunc:    json.Unmarshal,
	})
}

// Lookup translates a message ID using the request language.
// ok is false when the key is missing; msg is then the key itself.
// It never panics, even if the i18n middleware was not registered.
func Lookup(c *fiber.Ctx, id string, data map[string]interface{}) (msg string, ok bool) {
	defer func() {
		if r := recover(); r != nil {
			msg, ok = id, false
		}
	}()

	m, err := fiberi18n.Localize(c, &goi18n.LocalizeConfig{
		MessageID:    id,
		TemplateData: data,
	})
	if err != nil || m == "" {
		return id, false
	}
	return m, true
}

// Translate returns the translated message, or the key itself when missing.
func Translate(c *fiber.Ctx, id string) string {
	return TranslateWith(c, id, nil)
}

// TranslateWith is Translate with template data, e.g. {"Key": "user_uuid"}.
func TranslateWith(c *fiber.Ctx, id string, data map[string]interface{}) string {
	msg, _ := Lookup(c, id, data)

	return msg
}
