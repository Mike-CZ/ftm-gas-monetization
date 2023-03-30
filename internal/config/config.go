package config

import (
	"github.com/op/go-logging"
)

type Config struct {
	DB              DB
	Rpc             Rpc
	GasMonetization GasMonetization
	LoggingLevel    logging.Level
}

type Rpc struct {
	OperaRpcUrl string
}

type GasMonetization struct {
	StartFromBlock uint64
	// address of the gas monetization contract
	ContractAddress string
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
