// Package rpc provides high level access to the Fantom Opera blockchain
// node through RPC interface.
package rpc

import (
	"github.com/Mike-CZ/ftm-gas-monetization/internal/logger"
	client "github.com/ethereum/go-ethereum/rpc"
)

// Rpc represents the implementation of the Blockchain interface for Fantom Opera node.
type Rpc struct {
	ftm *client.Client
	log *logger.AppLogger
}

// New creates a new instance of the RPC client.
func New(url string, log *logger.AppLogger) *Rpc {
	rpcLogger := log.ModuleLogger("rpc")

	c, err := connect(url)
	if err != nil {
		rpcLogger.Criticalf("can not connect to the Opera node; %s", err.Error())
		return nil
	}

	rpc := &Rpc{
		ftm: c,
		log: rpcLogger,
	}

	return rpc
}

// connect opens RPC connection to the Opera node.
func connect(url string) (*client.Client, error) {
	c, err := client.Dial(url)
	if err != nil {
		return nil, err
	}
	return c, nil
}
