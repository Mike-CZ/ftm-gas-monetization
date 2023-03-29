package db

import (
	"context"
	"fmt"
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

// WhereRequestEpoch adds a where clause to the query builder.
func (qb *WithdrawalRequestQueryBuilder) WhereRequestEpoch(epoch uint64) *WithdrawalRequestQueryBuilder {
	qb.where = append(qb.where, "request_epoch = :request_epoch")
	qb.parameters["request_epoch"] = epoch
	return qb
}

// WhereWithdrawEpoch adds a where clause to the query builder.
func (qb *WithdrawalRequestQueryBuilder) WhereWithdrawEpoch(epoch uint64) *WithdrawalRequestQueryBuilder {
	qb.where = append(qb.where, "withdraw_epoch = :withdraw_epoch")
	qb.parameters["withdraw_epoch"] = epoch
	return qb
}

// StoreWithdrawalRequest stores a new withdrawal request into the database.
func (db *Db) StoreWithdrawalRequest(ctx context.Context, request *types.WithdrawalRequest) error {
	query := `INSERT INTO withdrawal_request (project_id, request_epoch, withdraw_epoch, amount) 
				VALUES (:project_id, :request_epoch, :withdraw_epoch, :amount)`
	_, err := sqlx.NamedExecContext(ctx, db.con, query, request)
	if err != nil {
		db.log.Errorf("failed to store withdrawal request %d: %v", request.Id, err)
		return err
	}
	return nil
}

// UpdateWithdrawalRequest updates the withdrawal request in the database.
func (db *Db) UpdateWithdrawalRequest(ctx context.Context, request *types.WithdrawalRequest) error {
	if request.Id == 0 {
		return fmt.Errorf("failed to update withdrawal. request id is 0")
	}
	query := `UPDATE withdrawal_request SET project_id = :project_id, request_epoch = :request_epoch,
                              withdraw_epoch = :withdraw_epoch, amount = :amount WHERE id = :id`
	_, err := sqlx.NamedExecContext(ctx, db.con, query, request)
	if err != nil {
		db.log.Errorf("failed to update withdrawal request %d: %v", request.Id, err)
		return err
	}
	return nil
}
