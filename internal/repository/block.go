package repository

import (
	"github.com/ethereum/go-ethereum/core/types"
)

// GetHeader pulls given block header by the block number.
func (repo *Repository) GetHeader(id uint64) (*types.Header, error) {
	header, err := repo.rpc.GetHeader(id)
	if err != nil {
		return nil, err
	}
	return header, nil
}
