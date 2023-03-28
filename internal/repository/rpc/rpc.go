// Package rpc provides high level access to the Fantom Opera blockchain
// node through RPC interface.
package rpc

import (
	"bytes"
	"context"
	"embed"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/config"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/logger"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository/rpc/contracts"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	client "github.com/ethereum/go-ethereum/rpc"
)

//go:embed contracts/abi/*.abi
var abiFiles embed.FS

// Rpc represents the implementation of the Blockchain interface for Fantom Opera node.
type Rpc struct {
	ftm *client.Client
	log *logger.AppLogger

	abiGasMonetization  *abi.ABI
	dataProviderSession *contracts.GasMonetizationSession
}

// New creates a new instance of the RPC client.
func New(cfg *config.Rpc, log *logger.AppLogger) *Rpc {
	rpcLogger := log.ModuleLogger("rpc")

	c, err := connect(cfg.OperaRpcUrl)
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

	// initialize data provider session
	if err = loadDataProviderSession(rpc, cfg); err != nil {
		rpcLogger.Criticalf("can not initialize data provider session; %s", err.Error())
		return nil
	}

	return rpc
}

// SetDataProviderSession sets the data provider session.
// This is intended to be used only for testing purposes.
func (rpc *Rpc) SetDataProviderSession(session *contracts.GasMonetizationSession) {
	rpc.dataProviderSession = session
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

// initializeDataProviderSession initializes the data provider session.
func loadDataProviderSession(rpc *Rpc, cfg *config.Rpc) error {
	key, err := crypto.HexToECDSA(cfg.DataProviderPK)
	if err != nil {
		return err
	}
	// create gas monetization instance
	ethClient := ethclient.NewClient(rpc.ftm)
	gm, err := contracts.NewGasMonetization(common.HexToAddress(cfg.GasMonetizationAddr), ethClient)
	if err != nil {
		return err
	}
	// get chain id
	chainId, err := ethClient.ChainID(context.Background())
	if err != nil {
		return err
	}
	// create data provider session
	auth, err := bind.NewKeyedTransactorWithChainID(key, chainId)
	rpc.dataProviderSession = &contracts.GasMonetizationSession{
		Contract: gm,
		CallOpts: bind.CallOpts{},
		TransactOpts: bind.TransactOpts{
			From:     auth.From,
			Signer:   auth.Signer,
			GasLimit: 0,
		},
	}
	return nil
}
