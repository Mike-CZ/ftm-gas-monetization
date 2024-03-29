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
	cfg.SetDefault("logger.loggingLevel", logging.INFO)
	cfg.SetDefault("logger.logFormat", "%{color}%{level:-8s} %{shortpkg}/%{shortfunc}%{color:reset}: %{message}")

	// rpc
	cfg.SetDefault("rpc.operaRpcUrl", "https://rpcapi.fantom.network")
	cfg.SetDefault("rpc.tracingRpcUrl", "https://rpcapi-tracing.fantom.network")

	// gas monetization
	cfg.SetDefault("gasMonetization.contractAddress", "0x9f6089633272C23cFD6E9C146b6E87cc9f065718")
	cfg.SetDefault("gasMonetization.dataProviderPK", "904d5dea0bdffb09d78a81c15f0b3b893f504679eb8cd1de585309cad58e6285")
	cfg.SetDefault("gasMonetization.startFromBlock", 0)

	// apiserver server
	cfg.SetDefault("api.readTimeout", 2)
	cfg.SetDefault("api.writeTimeout", 15)
	cfg.SetDefault("api.idleTimeout", 1)
	cfg.SetDefault("api.headerTimeout", 1)
	cfg.SetDefault("api.resolverTimeout", 30)
	cfg.SetDefault("api.bindAddress", "localhost:16761")
	cfg.SetDefault("api.domainAddress", "localhost:16761")
	cfg.SetDefault("api.corsOrigin", []string{"*"})
}
