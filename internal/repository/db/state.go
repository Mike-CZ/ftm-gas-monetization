package db

import (
	"context"
	"database/sql"
	"errors"
	"ftm-gas-monetization/internal/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/jmoiron/sqlx"
	"math/big"
)

const (
	// stateKeyLastProcessedBlock is the key of the last processed block in the state table.
	stateKeyLastProcessedBlock = "last_block"

	// stateCurrentEpoch is the key of the current epoch in the state table.
	stateCurrentEpoch = "current_epoch"

	// stateKeyTotalAmountWithdrawn is the key of the total amount collected in the state table.
	stateKeyTotalAmountCollected = "total_amount_collected"

	// stateKeyTotalAmountClaimed is the key of the total amount claimed in the state table.
	stateKeyTotalAmountClaimed = "total_amount_claimed"

	// stateKeyTotalTransactionsCount is the key of the total transactions count in the state table.
	stateKeyTotalTransactionsCount = "total_transactions_count"
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

// CurrentEpoch returns the current epoch.
func (db *Db) CurrentEpoch(ctx context.Context) (uint64, error) {
	db.log.Debugf("getting last epoch from the database")

	var currentEpoch uint64
	err := sqlx.GetContext(ctx, db.con, &currentEpoch, "SELECT value FROM state WHERE key = $1", stateCurrentEpoch)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			db.log.Warningf("no current epoch found, assuming 0")
			return 0, nil
		}
		db.log.Errorf("failed to get current epoch: %s", err)
		return 0, err
	}
	db.log.Debugf("current epoch is %d", currentEpoch)
	return currentEpoch, nil
}

// UpdateCurrentEpoch updates the current epoch.
func (db *Db) UpdateCurrentEpoch(ctx context.Context, epoch uint64) error {
	_, err := db.con.ExecContext(ctx,
		"INSERT INTO state (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = $2",
		stateCurrentEpoch, epoch)
	if err != nil {
		db.log.Errorf("failed to update current epoch: %s", err)
		return err
	}
	db.log.Noticef("setting current epoch to %d", epoch)
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

// SetTotalAmountCollected sets the total amount collected.
func (db *Db) SetTotalAmountCollected(ctx context.Context, amount *big.Int) error {
	_, err := db.con.ExecContext(ctx,
		"INSERT INTO state (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = $2",
		stateKeyTotalAmountCollected, &types.Big{Big: hexutil.Big(*amount)})
	if err != nil {
		db.log.Errorf("failed to set total amount collected: %s", err)
		return err
	}
	db.log.Noticef("total amount collected set to %d", amount)
	return nil
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

// SetTotalAmountClaimed sets the total amount claimed.
func (db *Db) SetTotalAmountClaimed(ctx context.Context, amount *big.Int) error {
	_, err := db.con.ExecContext(ctx,
		"INSERT INTO state (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = $2",
		stateKeyTotalAmountClaimed, &types.Big{Big: hexutil.Big(*amount)})
	if err != nil {
		db.log.Errorf("failed to set total amount claimed: %s", err)
		return err
	}
	db.log.Noticef("total amount claimed set to %d", amount)
	return nil
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

// TotalTransactionsCount returns the total number of transactions.
func (db *Db) TotalTransactionsCount(ctx context.Context) (uint64, error) {
	var totalTransactionsCount uint64
	err := sqlx.GetContext(ctx, db.con, &totalTransactionsCount, "SELECT value FROM state WHERE key = $1", stateKeyTotalTransactionsCount)
	if err != nil {
		if err == sql.ErrNoRows {
			db.log.Warningf("no total transactions count found, assuming 0")
			return 0, nil
		}
		db.log.Errorf("failed to get total transactions count: %s", err)
		return 0, err
	}
	db.log.Debugf("total transactions count is %d", totalTransactionsCount)
	return totalTransactionsCount, nil
}

// SetTotalTransactionsCount sets the total number of transactions.
func (db *Db) SetTotalTransactionsCount(ctx context.Context, count uint64) error {
	_, err := db.con.ExecContext(ctx,
		"INSERT INTO state (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = $2",
		stateKeyTotalTransactionsCount, count)
	if err != nil {
		db.log.Errorf("failed to set total transactions count: %s", err)
		return err
	}
	db.log.Noticef("total transactions count set to %d", count)
	return nil
}

// IncreaseTotalTransactionsCount increases the total number of transactions.
func (db *Db) IncreaseTotalTransactionsCount(ctx context.Context, count uint64) error {
	currentCount, err := db.TotalTransactionsCount(ctx)
	if err != nil {
		return err
	}
	newCount := currentCount + count
	_, err = db.con.ExecContext(ctx,
		"INSERT INTO state (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = $2",
		stateKeyTotalTransactionsCount, newCount)
	if err != nil {
		db.log.Errorf("failed to increase total transactions count: %s", err)
		return err
	}
	db.log.Noticef("total transactions count increased by %d", count)
	return nil
}
