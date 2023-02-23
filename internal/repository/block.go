package repository

import (
	"github.com/ethereum/go-ethereum/core/types"
)

// CurrentHead provides the ID of the latest known block.
func (repo *Repository) CurrentHead() (uint64, error) {
	return repo.rpc.CurrentHead()
}

// LastBlock returns the last observed block number.
func (repo *Repository) LastBlock() (uint64, error) {
	return repo.db.LastBlock()
}

// UpdateLastBlock updates the last observed block number.
func (repo *Repository) UpdateLastBlock(id uint64) error {
	return repo.db.UpdateLastBlock(id)
}

// GetHeader pulls given block header by the block number.
func (repo *Repository) GetHeader(id uint64) (*types.Header, error) {
	header, err := repo.rpc.GetHeader(id)
	if err != nil {
		return nil, err
	}
	return header, nil
}
