// Package types implements different core types of the API.
package types

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"math/big"
	"time"
)

// TransactionDecimalsCorrection is used to manipulate precision of a transaction amount value,
// so it can be stored in database as INT64 without loosing too much data
var TransactionDecimalsCorrection = new(big.Int).SetUint64(1000000000)

// TransactionGasCorrection is used to restore the precision on the transaction gas value calculations.
var TransactionGasCorrection = new(big.Int).SetUint64(10000000)

// Transaction represents basic information provided by the API about transaction inside Opera blockchain.
type Transaction struct {
	// Hash represents 32 bytes hash of the transaction.
	Hash common.Hash `json:"hash" db:"hash"`

	// BlockHash represents hash of the block where this transaction was in. nil when its pending.
	BlockHash *common.Hash `json:"blockHash" db:"block_hash"`

	// BlockNumber represents number of the block where this transaction was in. nil when its pending.
	BlockNumber *hexutil.Uint64 `json:"blockNumber" db:"block_number"`

	// TimeStamp represents the time stamp of the transaction.
	TimeStamp time.Time `json:"timestamp" db:"timestamp"`

	// From represents address of the sender.
	From common.Address `json:"from" db:"from"`

	// Gas represents gas provided by the sender.
	Gas hexutil.Uint64 `json:"gas" db:"gas_limit"`

	// Gas represents gas provided by the sender.
	GasUsed *hexutil.Uint64 `json:"gasUsed" db:"gas_used"`

	// GasPrice represents gas price provided by the sender in Wei.
	GasPrice hexutil.Big `json:"gasPrice" db:"gas_price"`

	// To represents the address of the receiver. nil when its a contract creation transaction.
	To *common.Address `json:"to,omitempty" db:"to"`

	// Status represents transaction status; value is either 1 (success) or 0 (failure)
	Status *hexutil.Uint64 `json:"status,omitempty"`

	// GasGWei represents gas price in gWei
	GasGWei int64 `json:"-" bson:"gwx100"`
}

// ComputedGWei calculates gas price in gWei.
func (trx *Transaction) ComputedGWei() int64 {
	return new(big.Int).Div(trx.GasPrice.ToInt(), TransactionGasCorrection).Int64()
}
