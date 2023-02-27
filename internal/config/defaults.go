package config

import (
	"github.com/op/go-logging"
	"github.com/spf13/viper"
)

func applyDefaults(cfg *viper.Viper) {
	// db
	cfg.SetDefault("db.user", "root")
	cfg.SetDefault("db.password", "root")
	cfg.SetDefault("db.host", "localhost")
	cfg.SetDefault("db.port", 5432)
	cfg.SetDefault("db.name", "gas_monetization")

	// logger
	cfg.SetDefault("loggingLevel", logging.INFO)

	// rpc api
	cfg.SetDefault("operaRpcUrl", "https://rpcapi.fantom.network")
}
