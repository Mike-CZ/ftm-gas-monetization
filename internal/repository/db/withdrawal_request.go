package db

import (
	"context"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/types"
	"github.com/jmoiron/sqlx"
)

type WithdrawalRequestQueryBuilder struct {
	queryBuilder[types.WithdrawalRequest]
}

// WithdrawalRequestQuery returns a new withdrawal request query builder.
func (db *Db) WithdrawalRequestQuery(ctx context.Context) WithdrawalRequestQueryBuilder {
	return WithdrawalRequestQueryBuilder{
		queryBuilder: newQueryBuilder[types.WithdrawalRequest](ctx, db.con, "withdrawal_request"),
	}
}

// WhereProjectId adds a where clause to the query builder.
func (qb *WithdrawalRequestQueryBuilder) WhereProjectId(projectId int64) *WithdrawalRequestQueryBuilder {
	qb.where = append(qb.where, "project_id = :project_id")
	qb.parameters["project_id"] = projectId
	return qb
}

// WhereEpoch adds a where clause to the query builder.
func (qb *WithdrawalRequestQueryBuilder) WhereEpoch(epoch uint64) *WithdrawalRequestQueryBuilder {
	qb.where = append(qb.where, "epoch = :epoch")
	qb.parameters["epoch"] = epoch
	return qb
}

// StoreWithdrawalRequest stores a new withdrawal request into the database.
func (db *Db) StoreWithdrawalRequest(ctx context.Context, request *types.WithdrawalRequest) error {
	query := `INSERT INTO withdrawal_request (project_id, epoch) VALUES (:project_id, :epoch)`
	_, err := sqlx.NamedExecContext(ctx, db.con, query, request)
	if err != nil {
		db.log.Errorf("failed to store withdrawal request %d: %v", request.Id, err)
		return err
	}
	return nil
}
