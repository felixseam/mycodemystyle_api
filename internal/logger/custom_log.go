package logger

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type CustomLog struct {
	MessageID string
	LogReason string
	Function  string
	File      string
	Line      int
}

func (e *CustomLog) LogToString() string {
	return fmt.Sprintf("%s | %s | %s | %s:%d", e.MessageID, e.LogReason, e.Function, e.File, e.Line)
}

func NewCustomLog(messageID, logDesc string, logType ...string) *CustomLog {
	pc, file, line, ok := runtime.Caller(1)
	function := "unknown"
	if ok {
		if fn := runtime.FuncForPC(pc); fn != nil {
			function = fn.Name()
		}
	}
	// drop the module path: "mycodemystyle_api/internal/router.(*Router).x" -> "router.(*Router).x"
	if i := strings.LastIndex(function, "/"); i >= 0 {
		function = function[i+1:]
	}

	msg := &CustomLog{
		MessageID: messageID,
		LogReason: logDesc,
		Function:  function,
		File:      filepath.Base(file),
		Line:      line,
	}

	level := "info"
	if len(logType) > 0 {
		level = logType[0]
	}

	var ev *zerolog.Event
	switch level {
	case "fatal":
		ev = log.Fatal()
	case "error":
		ev = log.Error()
	case "warn":
		ev = log.Warn()
	case "debug":
		ev = log.Debug()
	default:
		ev = log.Info()
	}

	ev.Str("MessageID", msg.MessageID).
		Str("Function", msg.Function).
		Str("File", msg.File).
		Int("Line", msg.Line).
		Msg(msg.LogReason)

	return msg
}
