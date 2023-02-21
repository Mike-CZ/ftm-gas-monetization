package config

import (
	"github.com/Mike-CZ/ftm-gas-monetization/cmd/gas-monetization-cli/flags"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/logger"
	"github.com/op/go-logging"
	"github.com/urfave/cli/v2"
)

type Config struct {
	LoggingLevel logging.Level
	OperaRpcUrl  string
	DbUser       string
	DbPassword   string
	DbHost       string
	DbPort       string
	DbName       string
}

// LoadFromCli loads the config from the given cli context.
func LoadFromCli(ctx *cli.Context) *Config {
	return &Config{
		LoggingLevel: logger.ParseLevel(ctx.String(flags.LogLevel.Name)),
		OperaRpcUrl:  ctx.String(flags.OperaRpcUrl.Name),
		DbUser:       "root",
		DbPassword:   "root",
		DbHost:       "localhost",
		DbPort:       "5432",
		DbName:       "gas_monetization",
	}
}
