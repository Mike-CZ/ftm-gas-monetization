package rpc

import (
	"github.com/Mike-CZ/ftm-gas-monetization/internal/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	eth "github.com/ethereum/go-ethereum/core/types"
)

// Transaction returns information about a blockchain transaction by hash.
func (rpc *Rpc) Transaction(hash *common.Hash) (*types.Transaction, error) {
	// keep track of the operation
	rpc.log.Debugf("loading transaction %s", hash.String())

	// call for data
	var trx types.Transaction
	err := rpc.ftm.Call(&trx, "eth_getTransactionByHash", hash)
	if err != nil {
		rpc.log.Error("transaction could not be extracted")
		return nil, err
	}

	// is there a block reference already?
	if trx.BlockNumber != nil {
		// get transaction receipt
		var rec struct {
			GasUsed hexutil.Uint64 `json:"gasUsed"`
			Logs    []eth.Log      `json:"logs"`
		}

		// call for the transaction receipt data
		err := rpc.ftm.Call(&rec, "eth_getTransactionReceipt", hash)
		if err != nil {
			rpc.log.Errorf("can not get receipt for transaction %s", hash)
			return nil, err
		}

		// copy some data
		trx.GasUsed = &rec.GasUsed
		trx.Logs = rec.Logs
	}

	// keep track of the operation
	rpc.log.Debugf("transaction %s loaded", hash.String())
	return &trx, nil
}
