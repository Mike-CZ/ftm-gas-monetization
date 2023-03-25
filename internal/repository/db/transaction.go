package db

import (
	"context"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/types"
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

// StoreTransaction stores a transaction reference in connected persistent storage.
func (db *Db) StoreTransaction(ctx context.Context, trx *types.Transaction) error {
	query := `INSERT INTO transaction (project_id, hash, block_hash, block_number, timestamp, from_address, to_address, gas_used, gas_price, reward_to_claim) 
		VALUES (:project_id, :hash, :block_hash, :block_number, :timestamp, :from_address, :to_address, :gas_used, :gas_price, :reward_to_claim)`

	_, err := sqlx.NamedExecContext(ctx, db.con, query, trx)
	if err != nil {
		db.log.Errorf("failed to store transaction %s: %v", trx.Hash.String(), err)
		return err
	}

	// add transaction to the db
	db.log.Debugf("transaction %s added to database", trx.Hash.String())
	return nil
}
