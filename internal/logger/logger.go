package logger

import (
	"github.com/Mike-CZ/ftm-gas-monetization/internal/config"
	"github.com/op/go-logging"
	"github.com/urfave/cli/v2"
	"strings"
)

// defaultLogFormat defines the format used for log output.
const defaultLogFormat = "%{color}%{level:-8s} %{shortpkg}/%{shortfunc}%{color:reset}: %{message}"

// AppLogger defines extended logger with generic no-level logging option
type AppLogger struct {
	logging.Logger
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

// New provides a new instance of the Logger based on output writer, logging level and module.
func New(ctx *cli.Context, cfg *config.Config) *AppLogger {
	backend := logging.NewLogBackend(ctx.App.Writer, "", 0)

	fm := logging.MustStringFormatter(defaultLogFormat)
	fmtBackend := logging.NewBackendFormatter(backend, fm)

	lvlBackend := logging.AddModuleLevel(fmtBackend)
	lvlBackend.SetLevel(cfg.LoggingLevel, "")

	logging.SetBackend(lvlBackend)
	l := logging.MustGetLogger(ctx.App.Name)

	return &AppLogger{Logger: *l}
}

// ParseLevel parses a string into a logging.Level. If the string is not a valid
// logging level, logging.INFO is returned.
func ParseLevel(level string) logging.Level {
	lvl, err := logging.LogLevel(level)
	if err == nil {
		return lvl
	}
	return logging.INFO
}
