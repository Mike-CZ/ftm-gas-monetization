package config

import (
	"github.com/op/go-logging"
)

type Config struct {
	DB           DB
	LoggingLevel logging.Level
	OperaRpcUrl  string
}

type DB struct {
	User     string
	Password string
	Host     string
	Port     int
	Name     string
}
