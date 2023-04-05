package logger

import (
	"github.com/op/go-logging"
	"io"
	"strings"
)

// defaultLogFormat defines the format used for log output.
const defaultLogFormat = "%{color}%{level:-8s} %{shortpkg}/%{shortfunc}%{color:reset}: %{message}"

// AppLogger defines extended logger with generic no-level logging option
type AppLogger struct {
	logging.Logger
}

// Printf implements default non-leveled output.
// We assume the information is low in importance if passed to this function so we relay it to Debug level.
func (a AppLogger) Printf(format string, args ...interface{}) {
	a.Debugf(format, args...)
}

// ModuleLogger provides a new instance of the Logger for a module.
func (l *AppLogger) ModuleLogger(module string) *AppLogger {
	var sb strings.Builder
	sb.WriteString(l.Module)
	sb.WriteString(".")
	sb.WriteString(module)
	log := logging.MustGetLogger(sb.String())
	return &AppLogger{Logger: *log}
}

// New provides a new instance of the Logger.
func New(out io.Writer, module string, lvl logging.Level) *AppLogger {
	backend := logging.NewLogBackend(out, "", 0)

	fm := logging.MustStringFormatter(defaultLogFormat)
	fmtBackend := logging.NewBackendFormatter(backend, fm)

	lvlBackend := logging.AddModuleLevel(fmtBackend)
	lvlBackend.SetLevel(lvl, "")

	logging.SetBackend(lvlBackend)
	l := logging.MustGetLogger(module)

	return &AppLogger{Logger: *l}
}
