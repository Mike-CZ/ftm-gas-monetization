// Package main defines the Gas Monetization CLI entry point
package main

import (
	"fmt"
	"github.com/Mike-CZ/ftm-gas-monetization/cmd/gas-monetization-cli/flags"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func initApp() *cli.App {
	return &cli.App{
		Name:     "FTM Gas Monetization",
		HelpName: "gas-monetization",
		Usage:    "starts observing blocks and accumulating pending rewards for white-listed addresses",
		Flags: []cli.Flag{
			&flags.OperaRpcUrl,
		},
		Action: func(*cli.Context) error {
			fmt.Println("boom! I say!")
			return nil
		},
	}
}

func main() {
	app := initApp()
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
