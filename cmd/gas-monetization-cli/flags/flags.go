// Package flags defines all the flags used by the gas monetization app.
package flags

import "github.com/urfave/cli/v2"

var (
	// OperaRpcUrl defines the opera rpc url
	OperaRpcUrl = cli.StringFlag{
		Name:     "rpc-url",
		Usage:    "opera rpc url",
		Required: true,
	}
	// LogLevel defines the level of logging of the app
	LogLevel = cli.StringFlag{
		Name:    "log",
		Aliases: []string{"l"},
		Usage:   "Level of the logging of the app action (\"critical\", \"error\", \"warning\", \"notice\", \"info\", \"debug\")",
		Value:   "info",
	}
)
