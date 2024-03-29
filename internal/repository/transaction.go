package repository

import (
	"context"
	"ftm-gas-monetization/internal/repository/db"
	"ftm-gas-monetization/internal/types"
	"github.com/ethereum/go-ethereum/common"
)

// TransactionQuery returns a new transaction query builder.
func (repo *Repository) TransactionQuery() db.TransactionQueryBuilder {
	return repo.db.TransactionQuery(context.Background())
}

// Transaction returns a transaction at Opera blockchain by a hash, nil if not found.
// If the transaction is not found, ErrTransactionNotFound error is returned.
func (repo *Repository) Transaction(hash *common.Hash) (*types.Transaction, error) {
	return repo.rpc.Transaction(hash)
}

// TraceTransaction returns the structured transaction traces.
func (repo *Repository) TraceTransaction(hash common.Hash) ([]types.TransactionTrace, error) {
	return repo.tracer.TraceTransaction(hash)
}
