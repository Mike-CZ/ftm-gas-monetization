package db

import (
	"context"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/types"
	"github.com/jmoiron/sqlx"
)

//goland:noinspection SqlDialectInspection,SqlNoDataSourceInspection
var transactionSchema = `
CREATE TABLE IF NOT EXISTS transaction (
    hash VARCHAR(64) PRIMARY KEY,
    block_hash VARCHAR(64),
    block_number BIGINT,
	timestamp TIMESTAMP NOT NULL,
	from_address VARCHAR(40) NOT NULL,
	to_address VARCHAR(40),
	gas_limit BIGINT NOT NULL,
    gas_used BIGINT,
    gas_price TEXT NOT NULL
);
`

// StoreTransaction stores a transaction reference in connected persistent storage.
//
//goland:noinspection SqlDialectInspection,SqlNoDataSourceInspection
func (db *Db) StoreTransaction(ctx context.Context, trx *types.Transaction) error {
	query := `INSERT INTO transaction (hash, block_hash, block_number, timestamp, from_address, to_address, gas_limit, gas_used, gas_price) 
		VALUES (:hash, :block_hash, :block_number, :timestamp, :from_address, :to_address, :gas_limit, :gas_used, :gas_price)`

	_, err := sqlx.NamedExecContext(ctx, db.con, query, trx)
	if err != nil {
		db.log.Errorf("failed to store transaction %s: %v", trx.Hash.String(), err)
		return err
	}

	// add transaction to the db
	db.log.Debugf("transaction %s added to database", trx.Hash.String())
	return nil
}

// migrateTransactionTables migrates the transaction tables.
func (db *Db) migrateTransactionTables() {
	_, err := db.db.Exec(transactionSchema)
	if err != nil {
		db.log.Panicf("failed to migrate state tables: %v", err)
	}
}
