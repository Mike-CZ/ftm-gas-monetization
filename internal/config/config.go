package config

import "github.com/op/go-logging"

type Config struct {
	LoggingLevel logging.Level
	OperaRpcUrl  string
}
