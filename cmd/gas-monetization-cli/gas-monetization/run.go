package gas_monetization

import (
	"github.com/Mike-CZ/ftm-gas-monetization/cmd/gas-monetization-cli/flags"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/app"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/config"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/logger"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository/tracing"
	"github.com/ethereum/go-ethereum/common"
	"github.com/op/go-logging"
	"github.com/urfave/cli/v2"
	"os"
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

	l := logger.New(ctx.App.Writer, "test", logging.DEBUG)
	tracer := tracing.New(&cfg.Rpc, l)

	test, _ := tracer.TraceTransaction(common.HexToHash("0x98330fe1e62f0c35dc7d4dee4178158939bb5128aa870b840626eb587cb2e34e"))
	l.Infof("data: %s", test[0].Action.To.Hex())
	os.Exit(0)

	app.Bootstrap(ctx, cfg)
	app.Start()

	return nil
}
