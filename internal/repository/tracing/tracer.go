package tracing

import (
	"ftm-gas-monetization/internal/config"
	"ftm-gas-monetization/internal/logger"
	"ftm-gas-monetization/internal/types"
	"github.com/ethereum/go-ethereum/common"
	client "github.com/ethereum/go-ethereum/rpc"
)

// TracerInterface defines the interface of the TracingRpc client.
type TracerInterface interface {
	TraceTransaction(hash common.Hash) ([]types.TransactionTrace, error)
}

// Tracer represents the implementation of the Blockchain tracing interface for Fantom Opera node.
type Tracer struct {
	ftm *client.Client
	log *logger.AppLogger
}

// New creates a new instance of the TracingRpc client.
func New(rpcCfg *config.Rpc, log *logger.AppLogger) *Tracer {
	rpcLogger := log.ModuleLogger("tracing_rpc")

	c, err := connect(rpcCfg.TracingRpcUrl)
	if err != nil {
		rpcLogger.Criticalf("can not connect to the Opera tracing node; %s", err.Error())
		return nil
	}

	rpc := &Tracer{
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
