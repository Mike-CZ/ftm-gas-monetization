package db

import (
	"context"
	"database/sql"
	"errors"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/jmoiron/sqlx"
	"math/big"
)

const (
	// stateKeyLastProcessedBlock is the key of the last processed block in the state table.
	stateKeyLastProcessedBlock = "last_block"

	// stateKeyLastProcessedEpoch is the key of the last processed epoch in the state table.
	stateKeyLastProcessedEpoch = "last_epoch"

	// stateKeyTotalAmountWithdrawn is the key of the total amount collected in the state table.
	stateKeyTotalAmountCollected = "total_amount_collected"

	// stateKeyTotalAmountClaimed is the key of the total amount claimed in the state table.
	stateKeyTotalAmountClaimed = "total_amount_claimed"
)

// LastProcessedBlock returns the last processed block.
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

// TotalAmountCollected returns the total amount collected.
func (db *Db) TotalAmountCollected(ctx context.Context) (*big.Int, error) {
	var totalAmountCollected *types.Big
	err := sqlx.GetContext(ctx, db.con, &totalAmountCollected, "SELECT value FROM state WHERE key = $1", stateKeyTotalAmountCollected)
	if err != nil {
		if err == sql.ErrNoRows {
			db.log.Warningf("no total amount collected found, assuming 0")
			return &big.Int{}, nil
		}
		db.log.Errorf("failed to get total amount collected: %s", err)
		return nil, err
	}

	db.log.Debugf("total amount collected is %d", totalAmountCollected)
	return totalAmountCollected.ToInt(), nil
}

// IncreaseTotalAmountCollected increases the total amount collected.
func (db *Db) IncreaseTotalAmountCollected(ctx context.Context, amount *big.Int) error {
	currentAmount, err := db.TotalAmountCollected(ctx)
	if err != nil {
		return err
	}
	newAmount := new(big.Int)
	newAmount.Add(currentAmount, amount)
	_, err = db.con.ExecContext(ctx,
		"INSERT INTO state (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = $2",
		stateKeyTotalAmountCollected, &types.Big{Big: hexutil.Big(*newAmount)})
	if err != nil {
		db.log.Errorf("failed to increase total amount collected: %s", err)
		return err
	}

	db.log.Noticef("total amount collected increased by %d", amount)
	return nil
}

// TotalAmountClaimed returns the total amount claimed.
func (db *Db) TotalAmountClaimed(ctx context.Context) (*big.Int, error) {
	var totalAmountClaimed *types.Big
	err := sqlx.GetContext(ctx, db.con, &totalAmountClaimed, "SELECT value FROM state WHERE key = $1", stateKeyTotalAmountClaimed)
	if err != nil {
		if err == sql.ErrNoRows {
			db.log.Warningf("no total amount claimed found, assuming 0")
			return &big.Int{}, nil
		}
		db.log.Errorf("failed to get total amount claimed: %s", err)
		return nil, err
	}

	db.log.Debugf("total amount claimed is %d", stateKeyTotalAmountClaimed)
	return totalAmountClaimed.ToInt(), nil
}

// IncreaseTotalAmountClaimed increases the total amount claimed.
func (db *Db) IncreaseTotalAmountClaimed(ctx context.Context, amount *big.Int) error {
	currentAmount, err := db.TotalAmountClaimed(ctx)
	if err != nil {
		return err
	}
	newAmount := new(big.Int)
	newAmount.Add(currentAmount, amount)
	_, err = db.con.ExecContext(ctx,
		"INSERT INTO state (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = $2",
		stateKeyTotalAmountClaimed, &types.Big{Big: hexutil.Big(*newAmount)})
	if err != nil {
		db.log.Errorf("failed to increase total amount claimed: %s", err)
		return err
	}

	db.log.Noticef("total amount claimed increased by %d", amount)
	return nil
}
