// Package rpc provides high level access to the Fantom Opera blockchain
// node through RPC interface.
package rpc

import (
	"bytes"
	"embed"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/logger"
	"github.com/ethereum/go-ethereum/accounts/abi"
	client "github.com/ethereum/go-ethereum/rpc"
)

//go:embed contracts/abi/*.abi
var abiFiles embed.FS

// Rpc represents the implementation of the Blockchain interface for Fantom Opera node.
type Rpc struct {
	ftm *client.Client
	log *logger.AppLogger

	abiGasMonetization *abi.ABI
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

	// load and parse ABIs
	if err := loadABI(rpc); err != nil {
		rpcLogger.Criticalf("can not parse ABI files; %s", err.Error())
		return nil
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

// loadABI tries to load and parse expected ABI for contracts we need.
func loadABI(rpc *Rpc) (err error) {
	rpc.abiGasMonetization, err = loadABIFile("contracts/abi/gas_monetization.abi")
	if err != nil {
		return err
	}
	return nil
}

// loadABIFile reads specified ABI file and returns the decoded ABI.
func loadABIFile(path string) (*abi.ABI, error) {
	data, err := abiFiles.ReadFile(path)
	if err != nil {
		return nil, err
	}
	// parse ABI
	decoded, err := abi.JSON(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return &decoded, nil
}
