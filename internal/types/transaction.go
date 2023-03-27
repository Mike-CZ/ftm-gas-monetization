// Package types implements different core types of the API.
package types

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"time"
)

// Transaction represents basic information provided by the API about transaction inside Opera blockchain.
type Transaction struct {
	Id int64 `db:"id"`
	// ProjectId represents the project ID this transaction belongs to.
	ProjectId int64 `db:"project_id"`

	// Hash represents 32 bytes hash of the transaction.
	Hash *Hash `json:"hash" db:"hash"`

	// BlockHash represents hash of the block where this transaction was in. nil when it's pending.
	BlockHash *Hash `json:"blockHash" db:"block_hash"`

	// BlockNumber represents number of the block where this transaction was in. nil when it's pending.
	BlockNumber *hexutil.Uint64 `json:"blockNumber" db:"block_number"`

	// Epoch represents the epoch that transaction belongs to.
	Epoch hexutil.Uint64 `db:"epoch_number"`

	// Timestamp represents the time stamp of the transaction.
	Timestamp time.Time `json:"timestamp" db:"timestamp"`

	// From represents address of the sender.
	From *Address `json:"from" db:"from_address"`

	// To represents the address of the receiver. nil when it's a contract creation transaction.
	To *Address `json:"to,omitempty" db:"to_address"`

	// GasUsed represents the amount of gas used by this specific transaction alone.
	GasUsed *hexutil.Uint64 `json:"gasUsed" db:"gas_used"`

	// GasPrice represents gas price provided by the sender in Wei.
	GasPrice *Big `json:"gasPrice" db:"gas_price"`

	// RewardToClaim represents the amount of reward to claim in Wei.
	RewardToClaim *Big `db:"reward_to_claim"`

	// Logs represents a list of log records created along with the transaction
	Logs []types.Log `json:"logs"`
}
