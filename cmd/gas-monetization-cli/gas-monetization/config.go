// Package gas_monetization implements executable entry points to the gas monetization app.
package gas_monetization

import (
	"github.com/urfave/cli/v2"
	"os"
)

var configCommand = cli.Command{
	Name:  "config",
	Usage: "output default config",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:   "node-id",
			Usage:  "node id",
			Value:  getHostname(),
			EnvVar: "STELLAR_CONFIG_NODE_ID",
		},
		cli.StringFlag{
			Name:   "nic, n",
			Usage:  "network interface to use for detecting IP (default: first non-local)",
			EnvVar: "STELLAR_CONFIG_NIC",
		},
		cli.StringSliceFlag{
			Name:   "peer, p",
			Usage:  "peer(s) to configure for joining",
			Value:  &cli.StringSlice{},
			EnvVar: "STELLAR_CONFIG_PEERS",
		},
		cli.StringFlag{
			Name:   "namespace",
			Usage:  "containerd namespace",
			Value:  "default",
			EnvVar: "CONTAINERD_NAMESPACE",
		},
	},
	Action: func(ctx *cli.Context) error {
		cfg, err := defaultConfig(ctx)
		if err != nil {
			return err
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "    ")
		if err := enc.Encode(cfg); err != nil {
			return err
		}
		return nil
	},
}
