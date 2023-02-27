// Package main defines the Gas Monetization CLI entry point
package main

import (
	"github.com/Mike-CZ/ftm-gas-monetization/cmd/gas-monetization-cli/gas-monetization"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func initApp() *cli.App {
	return &cli.App{
		Name:     "Gas Monetization App",
		HelpName: "gas-monetization",
		Usage:    "starts observing blocks and accumulating pending rewards for white-listed addresses",
		Commands: []*cli.Command{
			&gas_monetization.CmdRun,
			&gas_monetization.CmdConfig,
		},
	}
}

func main() {
	app := initApp()
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
