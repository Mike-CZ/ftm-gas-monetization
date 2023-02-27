package repository

import (
	"github.com/Mike-CZ/ftm-gas-monetization/internal/types"
	"github.com/ethereum/go-ethereum/common"
)

// Transaction returns a transaction at Opera blockchain by a hash, nil if not found.
// If the transaction is not found, ErrTransactionNotFound error is returned.
func (repo *Repository) Transaction(hash *common.Hash) (*types.Transaction, error) {
	return repo.rpc.Transaction(hash)
}
