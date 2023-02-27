package gas_monetization

import (
	"github.com/Mike-CZ/ftm-gas-monetization/cmd/gas-monetization-cli/flags"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/app"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/config"
	"github.com/urfave/cli/v2"
)

// CmdRun defines a CLI command for running the gas monetization app.
var CmdRun = cli.Command{
	Action: run,
	Name:   "run",
	Usage:  `Runs the gas monetization app.`,
	Flags: []cli.Flag{
		&flags.Cfg,
	},
}

func run(ctx *cli.Context) error {
	cfg := config.Load(ctx)
	app.Bootstrap(ctx, cfg)
	app.Start()
	//res, err := app.Repository().BlockByNumber(nil)
	//if err != nil {
	//	fmt.Println(err)
	//	return nil
	//}
	//fmt.Println(res.Number.String())

	return nil
}
