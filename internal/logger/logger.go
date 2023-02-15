package logger

import (
	"github.com/op/go-logging"
	"io"
)

// defaultLogFormat defines the format used for log output.
const defaultLogFormat = "%{color}%{level:-8s} %{shortpkg}/%{shortfunc}%{color:reset}: %{message}"

// New provides a new instance of the Logger based on output writer, logging level and module.
func New(out io.Writer, lvl logging.Level, module string) *logging.Logger {
	backend := logging.NewLogBackend(out, "", 0)

	fm := logging.MustStringFormatter(defaultLogFormat)
	fmtBackend := logging.NewBackendFormatter(backend, fm)

	lvlBackend := logging.AddModuleLevel(fmtBackend)
	lvlBackend.SetLevel(lvl, "")

	logging.SetBackend(lvlBackend)
	return logging.MustGetLogger(module)
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
