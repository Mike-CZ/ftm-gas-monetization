package db

import (
	"context"
	"ftm-gas-monetization/internal/types"
	"github.com/jmoiron/sqlx"
)

type TransactionQueryBuilder struct {
	queryBuilder[types.Transaction]
}

// TransactionQuery returns a new transaction query builder.
func (db *Db) TransactionQuery(ctx context.Context) TransactionQueryBuilder {
	return TransactionQueryBuilder{
		queryBuilder: newQueryBuilder[types.Transaction](ctx, db.con, "transaction"),
	}
}

// WhereEpoch adds a where clause to the query builder.
func (qb *TransactionQueryBuilder) WhereEpoch(epoch uint64) *TransactionQueryBuilder {
	qb.where = append(qb.where, "epoch_number = :epoch_number")
	qb.parameters["epoch_number"] = epoch
	return qb
}

// WhereEpochLt adds a where clause to the query builder.
func (qb *TransactionQueryBuilder) WhereEpochLt(epoch uint64) *TransactionQueryBuilder {
	qb.where = append(qb.where, "epoch_number < :epoch_number")
	qb.parameters["epoch_number"] = epoch
	return qb
}

// WhereProjectId adds a where clause to the query builder.
func (qb *TransactionQueryBuilder) WhereProjectId(id int64) *TransactionQueryBuilder {
	qb.where = append(qb.where, "project_id = :project_id")
	qb.parameters["project_id"] = id
	return qb
}

// StoreTransaction stores a transaction reference in connected persistent storage.
func (db *Db) StoreTransaction(ctx context.Context, trx *types.Transaction) error {
	query := `INSERT INTO transaction (project_id, hash, block_hash, block_number, epoch_number, timestamp, from_address, to_address, gas_used, gas_price, reward_to_claim) 
		VALUES (:project_id, :hash, :block_hash, :block_number, :epoch_number, :timestamp, :from_address, :to_address, :gas_used, :gas_price, :reward_to_claim)`

	_, err := sqlx.NamedExecContext(ctx, db.con, query, trx)
	if err != nil {
		db.log.Errorf("failed to store transaction %s: %v", trx.Hash.String(), err)
		return err
	}

	// add transaction to the db
	db.log.Debugf("transaction %s added to database", trx.Hash.String())
	return nil
}
