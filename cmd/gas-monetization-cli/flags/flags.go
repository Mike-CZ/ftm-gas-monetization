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
)
