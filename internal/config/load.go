package config

import (
	"github.com/Mike-CZ/ftm-gas-monetization/cmd/gas-monetization-cli/flags"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
	"log"
)

func Load(ctx *cli.Context) *Config {
	var config Config

	cfg, err := readConfigFile(ctx.String(flags.Cfg.Name))
	if err != nil {
		log.Fatalf("can not read configuration file. Err: %v", err)
	}

	if err = cfg.Unmarshal(&config); err != nil {
		log.Fatalf("can not extract configuration. Err: %v", err)
	}

	return &config
}

func readConfigFile(path string) (*viper.Viper, error) {
	cfg := viper.New()
	cfg.SetConfigName("gas_monetization")
	cfg.AddConfigPath("/etc")
	cfg.AddConfigPath("/etc/gas_monetization")

	if path != "" {
		log.Println("loading config: ", path)
		cfg.SetConfigFile(path)
	}
	applyDefaults(cfg)

	if err := cfg.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Printf("configuration not found at %s", cfg.ConfigFileUsed())
			return nil, err
		}

		// config file not found; ignore the error, we may not need the config file
		log.Println("configuration file not found, using default values")
	}

	return cfg, nil
}
