// Package types implements different core types of the API.
package types

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// Block represents basic information provided by the API about block inside Opera blockchain.
type Block struct {
	// Number represents the block number. nil when its pending block.
	Number hexutil.Uint64 `json:"number"`

	// Epoch represents the block epoch.
	Epoch hexutil.Uint64 `json:"epoch"`

	// Hash represents hash of the block. nil when its pending block.
	Hash common.Hash `json:"hash"`

	// ParentHash represents hash of the parent block.
	ParentHash common.Hash `json:"parentHash"`

	// Size represents the size of this block in bytes.
	Size hexutil.Uint64 `json:"size"`

	// GasLimit represents the maximum gas allowed in this block.
	GasLimit hexutil.Uint64 `json:"gasLimit"`

	// GasUsed represents the actual total used gas by all transactions in this block.
	GasUsed hexutil.Uint64 `json:"gasUsed"`

	// TimeStamp represents the unix timestamp for when the block was collated.
	TimeStamp hexutil.Uint64 `json:"timestamp"`

	// Txs represents array of 32 bytes hashes of transactions included in the block.
	Txs []*common.Hash `json:"transactions"`
}
