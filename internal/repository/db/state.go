package db

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
)

const (
	// stateKeyLastProcessedBlock is the key of the last processed block in the state table.
	stateKeyLastProcessedBlock = "last_block"

	// stateKeyLastProcessedEpoch is the key of the last processed epoch in the state table.
	stateKeyLastProcessedEpoch = "last_epoch"
)

//goland:noinspection GoUnusedGlobalVariable,SqlDialectInspection,SqlNoDataSourceInspection
var stateSchema = `
CREATE TABLE IF NOT EXISTS state (
    key VARCHAR PRIMARY KEY,
    value TEXT
);
`

// LastProcessedBlock returns the last processed block.
//
//goland:noinspection SqlDialectInspection,SqlNoDataSourceInspection
func (db *Db) LastProcessedBlock(ctx context.Context) (uint64, error) {
	db.log.Debugf("getting last block from the database")

	var lastBlock uint64
	err := sqlx.GetContext(ctx, db.con, &lastBlock, "SELECT value FROM state WHERE key = $1", stateKeyLastProcessedBlock)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			db.log.Warningf("no last block found, assuming 0")
			return 0, nil
		}
		db.log.Errorf("failed to get last block: %s", err)
		return 0, err
	}

	db.log.Debugf("last block is %d", lastBlock)
	return lastBlock, nil
}

// UpdateLastProcessedBlock updates the last processed block.
//
//goland:noinspection SqlDialectInspection,SqlNoDataSourceInspection
func (db *Db) UpdateLastProcessedBlock(ctx context.Context, block uint64) error {
	_, err := db.con.ExecContext(ctx,
		"INSERT INTO state (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = $2",
		stateKeyLastProcessedBlock, block)
	if err != nil {
		db.log.Errorf("failed to update last block: %s", err)
		return err
	}

	db.log.Noticef("last block updated to %d", block)
	return nil
}

// LastProcessedEpoch returns the last processed epoch.
//
//goland:noinspection SqlDialectInspection,SqlNoDataSourceInspection
func (db *Db) LastProcessedEpoch(ctx context.Context) (uint64, error) {
	db.log.Debugf("getting last epoch from the database")

	var lastEpoch uint64
	err := sqlx.GetContext(ctx, db.con, &lastEpoch, "SELECT value FROM state WHERE key = $1", stateKeyLastProcessedEpoch)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			db.log.Warningf("no last epoch found, assuming 0")
			return 0, nil
		}
		db.log.Errorf("failed to get last epoch: %s", err)
		return 0, err
	}

	db.log.Debugf("last epoch is %d", lastEpoch)
	return lastEpoch, nil
}

// UpdateLastProcessedEpoch updates the last processed epoch.
//
//goland:noinspection SqlDialectInspection,SqlNoDataSourceInspection
func (db *Db) UpdateLastProcessedEpoch(ctx context.Context, epoch uint64) error {
	_, err := db.con.ExecContext(ctx,
		"INSERT INTO state (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = $2",
		stateKeyLastProcessedEpoch, epoch)
	if err != nil {
		db.log.Errorf("failed to update last epoch: %s", err)
		return err
	}

	db.log.Noticef("last epoch updated to %d", epoch)
	return nil
}

// migrateStateTables migrates the state tables.
func (db *Db) migrateStateTables() {
	_, err := db.db.Exec(stateSchema)
	if err != nil {
		db.log.Panicf("failed to migrate state tables: %v", err)
	}
}
