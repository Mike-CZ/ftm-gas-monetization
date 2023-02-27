// Package flags defines all the flags used by the gas monetization app.
package flags

import "github.com/urfave/cli/v2"

var (
	// Cfg defines path to config, if not set, config is searched in default paths
	Cfg = cli.StringFlag{
		Name:  "cfg",
		Usage: "path to config",
	}
)
