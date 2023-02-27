package repository

import (
	"context"
	"errors"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository/rpc"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	eth "github.com/ethereum/go-ethereum/rpc"
)

// ErrBlockNotFound represents an error returned if a block can not be found.
var ErrBlockNotFound = errors.New("requested block can not be found in Opera blockchain")

// BlockHeight returns the current height of the Opera blockchain in blocks.
func (repo *Repository) BlockHeight() (*hexutil.Big, error) {
	return repo.rpc.BlockHeight()
}

// LastProcessedBlock returns the last processed block number.
func (repo *Repository) LastProcessedBlock() (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbQueryTimeoutDuration)
	defer cancel()
	return repo.db.LastProcessedBlock(ctx)
}

// UpdateLastProcessedBlock updates the last observed block number.
func (repo *Repository) UpdateLastProcessedBlock(id uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbQueryTimeoutDuration)
	defer cancel()
	return repo.db.UpdateLastProcessedBlock(ctx, id)
}

// BlockByNumber returns a block at Opera blockchain represented by a number. Top block is returned if the number
// is not provided.
// If the block is not found, ErrBlockNotFound error is returned.
func (repo *Repository) BlockByNumber(num *hexutil.Uint64) (*types.Block, error) {
	// return the top block if block number is not provided
	if num == nil {
		tag := rpc.BlockTypeLatest
		return repo.blockByTag(&tag)
	}
	return repo.getBlock(num.String(), repo.blockByTag)
}

// BlockByHash returns a block at Opera blockchain represented by a hash. Top block is returned if the hash
// is not provided.
// If the block is not found, ErrBlockNotFound error is returned.
func (repo *Repository) BlockByHash(hash *common.Hash) (*types.Block, error) {
	// do we have a hash?
	if hash == nil {
		tag := rpc.BlockTypeLatest
		return repo.blockByTag(&tag)
	}
	return repo.getBlock(hash.String(), repo.rpc.BlockByHash)
}

// getBlock gets a block of given tag from cache, or from a repository pull function.
func (repo *Repository) getBlock(tag string, pull func(*string) (*types.Block, error)) (*types.Block, error) {
	// inform what we do
	repo.log.Debugf("block [%s] requested", tag)

	// extract the block from the chain
	blk, err := pull(&tag)
	if err != nil {
		// block simply not found?
		if err == eth.ErrNoResult {
			repo.log.Warning("block not found in the blockchain")
			return nil, ErrBlockNotFound
		}
		// something went wrong
		return nil, err
	}

	// inform what we do
	repo.log.Debugf("block [%s] loaded by pulling", tag)
	return blk, nil
}

// blockByTag returns a block at Opera blockchain represented by given tag.
// The tag could be an encoded block number, or a predefined string tag for "earliest", "latest" or "pending" block.
func (repo *Repository) blockByTag(tag *string) (*types.Block, error) {
	// inform what we do
	repo.log.Debugf("loading block [%s]", *tag)

	// extract the block
	block, err := repo.rpc.Block(tag)
	if err != nil {
		// block simply not found?
		if err == eth.ErrNoResult {
			repo.log.Warning("block not found in the blockchain")
			return nil, ErrBlockNotFound
		}

		// something went wrong
		return nil, err
	}

	return block, nil
}
