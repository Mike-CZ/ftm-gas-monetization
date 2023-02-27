// Package types implements different core types of the API.
package types

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"time"
)

// Transaction represents basic information provided by the API about transaction inside Opera blockchain.
type Transaction struct {
	// Hash represents 32 bytes hash of the transaction.
	Hash *Hash `json:"hash" db:"hash"`

	// BlockHash represents hash of the block where this transaction was in. nil when it's pending.
	BlockHash *Hash `json:"blockHash" db:"block_hash"`

	// BlockNumber represents number of the block where this transaction was in. nil when it's pending.
	BlockNumber *hexutil.Uint64 `json:"blockNumber" db:"block_number"`

	// TimeStamp represents the time stamp of the transaction.
	TimeStamp time.Time `json:"timestamp" db:"timestamp"`

	// From represents address of the sender.
	From *Address `json:"from" db:"from_address"`

	// To represents the address of the receiver. nil when it's a contract creation transaction.
	To *Address `json:"to,omitempty" db:"to_address"`

	// Gas represents gas provided by the sender.
	Gas hexutil.Uint64 `json:"gas" db:"gas_limit"`

	// Gas represents gas provided by the sender.
	GasUsed *hexutil.Uint64 `json:"gasUsed" db:"gas_used"`

	// GasPrice represents gas price provided by the sender in Wei.
	GasPrice *Big `json:"gasPrice" db:"gas_price"`
}
