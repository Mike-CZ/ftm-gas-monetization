package apiserver

import (
	"github.com/Mike-CZ/ftm-gas-monetization/cmd/gas-monetization-cli/flags"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/api"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/config"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/logger"
	"github.com/urfave/cli/v2"
)

// CmdRun defines a CLI command for running the gas monetization api.
var CmdRun = cli.Command{
	Action: run,
	Name:   "run",
	Usage:  `Runs the gas monetization api.`,
	Flags: []cli.Flag{
		&flags.Cfg,
	},
}

func run(ctx *cli.Context) error {
	cfg := config.Load(ctx)
	log := logger.New(ctx.App.Writer, ctx.App.HelpName, cfg.Logger.LoggingLevel)
	api.New(cfg, log)

	return nil
}
