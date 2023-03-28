package config

import (
	"github.com/op/go-logging"
)

type Config struct {
	DB           DB
	Rpc          Rpc
	LoggingLevel logging.Level
}

type Rpc struct {
	OperaRpcUrl string
	// address of the gas monetization contract
	GasMonetizationAddr string
	// private key of the account that will be used to provide data for gas monetization contract
	DataProviderPK string
}

type DB struct {
	User     string
	Password string
	Host     string
	Port     string
	Name     string
}
