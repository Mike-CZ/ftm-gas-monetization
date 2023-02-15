package gas_monetization

import (
	"fmt"
	"github.com/Mike-CZ/ftm-gas-monetization/cmd/gas-monetization-cli/flags"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/app"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/config"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/logger"
	"github.com/urfave/cli/v2"
)

// CmdRun defines a CLI command for running the gas monetization app.
var CmdRun = cli.Command{
	Action: run,
	Name:   "run",
	Usage:  `Runs the gas monetization app.`,
	Flags: []cli.Flag{
		&flags.OperaRpcUrl,
		&flags.LogLevel,
	},
}

func run(ctx *cli.Context) error {
	cfg := config.Config{
		LoggingLevel: logger.ParseLevel(ctx.String(flags.LogLevel.Name)),
		OperaRpcUrl:  ctx.String(flags.OperaRpcUrl.Name),
	}

	app.Bootstrap(ctx, &cfg)

	res, err := app.Repository().GetHeader(1)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	fmt.Println(res)

	return nil
}
