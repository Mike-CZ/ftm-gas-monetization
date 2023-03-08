package db

import (
	"context"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/types"
	"github.com/jmoiron/sqlx"
)

// StoreTransaction stores a transaction reference in connected persistent storage.
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
