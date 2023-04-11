package config

import (
	"github.com/op/go-logging"
)

type Config struct {
	DB              DB
	Rpc             Rpc
	Api             ApiServer
	Logger          Logging
	GasMonetization GasMonetization
	Slack           Slack
	AppName         string
}

type Rpc struct {
	OperaRpcUrl   string
	TracingRpcUrl string
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

type ApiServer struct {
	BindAddress     string
	DomainAddress   string
	ReadTimeout     int
	WriteTimeout    int
	IdleTimeout     int
	HeaderTimeout   int
	ResolverTimeout int
	CorsOrigin      []string
}

type Logging struct {
	LoggingLevel logging.Level
	LogFormat    string
}

// Slack is a configuration for Slack notifications.
type Slack struct {
	Token     string
	ChannelId string
}
