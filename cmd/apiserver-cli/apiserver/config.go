// Package apiserver implements executable entry points to the gas monetization api.
package apiserver

import (
	"encoding/json"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/config"
	"github.com/urfave/cli/v2"
)

var CmdConfig = cli.Command{
	Name:  "config",
	Usage: "Prints default config",
	Action: func(ctx *cli.Context) error {
		cfg := config.Load(ctx)
		enc := json.NewEncoder(ctx.App.Writer)
		enc.SetIndent("", "    ")
		if err := enc.Encode(cfg); err != nil {
			return err
		}
		return nil
	},
}
