// Package main defines the Gas Monetization API CLI entry point
package main

import (
	"github.com/Mike-CZ/ftm-gas-monetization/cmd/apiserver-cli/apiserver"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func initApp() *cli.App {
	return &cli.App{
		Name:     "Gas Monetization Api",
		HelpName: "gas-monetization-api",
		Usage:    "graphql api for gas monetization",
		Commands: []*cli.Command{
			&apiserver.CmdRun,
			&apiserver.CmdConfig,
		},
	}
}

func main() {
	app := initApp()
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
