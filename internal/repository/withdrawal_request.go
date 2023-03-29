package repository

import (
	"context"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository/db"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/types"
)

// WithdrawalRequestQuery returns a new withdrawal request query builder.
func (repo *Repository) WithdrawalRequestQuery() db.WithdrawalRequestQueryBuilder {
	return repo.db.WithdrawalRequestQuery(context.Background())
}

// StoreWithdrawalRequest stores a new withdrawal request into the database.
func (repo *Repository) StoreWithdrawalRequest(request *types.WithdrawalRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbQueryTimeoutDuration)
	defer cancel()
	return repo.db.StoreWithdrawalRequest(ctx, request)
}

// UpdateWithdrawalRequest updates the withdrawal request in the database.
func (repo *Repository) UpdateWithdrawalRequest(request *types.WithdrawalRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbQueryTimeoutDuration)
	defer cancel()
	return repo.db.UpdateWithdrawalRequest(ctx, request)
}
