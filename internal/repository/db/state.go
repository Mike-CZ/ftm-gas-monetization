package db

import (
	"database/sql"
	"errors"
)

//goland:noinspection GoUnusedGlobalVariable,SqlDialectInspection,SqlNoDataSourceInspection
var stateSchema = `
CREATE TABLE IF NOT EXISTS state (
    last_block BIGINT
);
`

// LastBlock returns the last processed block.
//
//goland:noinspection SqlDialectInspection,SqlNoDataSourceInspection
func (db *Db) LastBlock() (uint64, error) {
	var lastBlock uint64
	err := db.con.Get(&lastBlock, "SELECT last_block FROM state")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			db.log.Warningf("no last block found, starting from 0")
			return 0, nil
		}
		db.log.Errorf("failed to get last block: %s", err)
		return 0, err
	}
	db.log.Noticef("last block is %d", lastBlock)
	return lastBlock, nil
}

// UpdateLastBlock updates the last processed block.
//
//goland:noinspection SqlDialectInspection,SqlNoDataSourceInspection
func (db *Db) UpdateLastBlock(block uint64) error {
	_, err := db.con.Exec("INSERT INTO state (last_block) VALUES ($1) ON CONFLICT (last_block) DO UPDATE SET last_block = $1", block)
	if err != nil {
		db.log.Errorf("failed to update last block: %s", err)
		return err
	}
	db.log.Noticef("last block updated to %d", block)
	return nil
}

// migrateStateTables migrates the state tables.
func (db *Db) migrateStateTables() {
	_, err := db.con.Exec(stateSchema)
	if err != nil {
		db.log.Panicf("failed to migrate state tables: %v", err)
	}
}
