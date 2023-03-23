package repository

import (
	"context"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository/db"
)

// WithdrawalRequestQuery returns a new withdrawal request query builder.
func (repo *Repository) WithdrawalRequestQuery() db.WithdrawalRequestQueryBuilder {
	return repo.db.WithdrawalRequestQuery(context.Background())
}
