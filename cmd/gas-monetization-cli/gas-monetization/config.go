// Package gas_monetization implements executable entry points to the gas monetization app.
package gas_monetization

import (
	"encoding/json"
	"github.com/Mike-CZ/ftm-gas-monetization/cmd/gas-monetization-cli/flags"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/config"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/logger"
	"github.com/urfave/cli/v2"
)

var CmdConfig = cli.Command{
	Name:  "config",
	Usage: "Prints default config",
	Action: func(ctx *cli.Context) error {
		cfg := loadConfig(ctx)
		enc := json.NewEncoder(ctx.App.Writer)
		enc.SetIndent("", "    ")
		if err := enc.Encode(cfg); err != nil {
			return err
		}
		return nil
	},
}

func loadConfig(ctx *cli.Context) *config.Config {
	return &config.Config{
		LoggingLevel: logger.ParseLevel(ctx.String(flags.LogLevel.Name)),
		OperaRpcUrl:  ctx.String(flags.OperaRpcUrl.Name),
	}
}
