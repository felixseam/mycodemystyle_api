package logger

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// ANSI colors
const (
	reset  = "\033[0m"
	gray   = "\033[90m"
	red    = "\033[31m"
	yellow = "\033[33m"
	green  = "\033[32m"
	cyan   = "\033[36m"
	bgRed  = "\033[41;97m"
)

var levelColor = map[string]string{
	"debug": cyan,
	"info":  cyan,
	"warn":  yellow,
	"error": red,
	"fatal": bgRed,
}

// "2006-01-02 15:04:05" (19) + space + "| INFO  |" (9) + space = 30
var detailIndent = strings.Repeat(" ", 0)

func Init(level string) {
	lvl, err := zerolog.ParseLevel(strings.ToLower(level))
	if err != nil || level == "" {
		lvl = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(lvl)

	log.Logger = zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05",
		PartsOrder: []string{
			zerolog.TimestampFieldName,
			zerolog.LevelFieldName,
			zerolog.MessageFieldName,
		},

		FormatTimestamp: func(i interface{}) string {
			s := fmt.Sprint(i)
			if t, err := time.Parse(time.RFC3339, s); err == nil {
				s = t.Format("2006-01-02 15:04:05")
			}
			return "\n" + gray + s + reset
		},

		FormatLevel: func(i interface{}) string {
			l := strings.ToLower(fmt.Sprint(i))
			c, ok := levelColor[l]
			if !ok {
				c = reset
			}
			return fmt.Sprintf("\n%s%-5s%s", c, strings.ToUpper(l), reset)
		},

		// Turns the custom fields into an aligned detail block under the message.
		// Plain log lines (no MessageID) are left untouched.
		FormatPrepare: func(evt map[string]interface{}) error {
			id, ok := evt["MessageID"].(string)
			if !ok {
				return nil
			}
			file, _ := evt["File"].(string)
			fn, _ := evt["Function"].(string)
			line := evt["Line"]
			errStr, hasErr := evt[zerolog.ErrorFieldName].(string)
			lvl, _ := evt[zerolog.LevelFieldName].(string)

			delete(evt, "MessageID")
			delete(evt, "File")
			delete(evt, "Function")
			delete(evt, "Line")
			delete(evt, zerolog.ErrorFieldName)

			c, found := levelColor[lvl]
			if !found {
				c = reset
			}
			label := func(s string) string { return gray + s + reset }

			var b strings.Builder
			b.WriteString(fmt.Sprint(evt[zerolog.MessageFieldName]))
			fmt.Fprintf(&b, "\n%s%s %s%s%s", detailIndent, label("id   :"), c, id, reset)
			fmt.Fprintf(&b, "\n%s%s %s:%v", detailIndent, label("at   :"), file, line)
			fmt.Fprintf(&b, "\n%s%s %s", detailIndent, label("func :"), fn)
			if hasErr {
				fmt.Fprintf(&b, "\n%s%s %s%s%s", detailIndent, label("error:"), red, errStr, reset)
			}
			b.WriteString("\n") // blank line between entries

			evt[zerolog.MessageFieldName] = b.String()
			return nil
		},
	}).With().Timestamp().Logger()
}
