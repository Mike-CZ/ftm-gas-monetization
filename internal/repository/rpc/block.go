package rpc

import (
	"context"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

// CurrentHead provides the ID of the latest known block.
func (rpc *Rpc) CurrentHead() (uint64, error) {
	return rpc.ftm.BlockNumber(context.Background())
}

// GetHeader pulls given block header by the block number.
func (rpc *Rpc) GetHeader(id uint64) (*types.Header, error) {
	return rpc.ftm.HeaderByNumber(context.Background(), new(big.Int).SetUint64(id))
}
