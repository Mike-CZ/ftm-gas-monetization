package rpc

import (
	"fmt"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"math/big"
	"time"
)

// BlockTypeLatest represents the latest available block in blockchain.
const (
	BlockTypeLatest = "latest"
)

// BlockHeight returns the current block height of the Opera blockchain.
func (rpc *Rpc) BlockHeight() (*hexutil.Big, error) {
	// keep track of the operation
	rpc.log.Debugf("checking current block height")

	// call for data
	var height hexutil.Big
	err := rpc.ftm.Call(&height, "eth_blockNumber")
	if err != nil {
		rpc.log.Error("block height could not be obtained")
		return nil, err
	}

	// inform and return
	rpc.log.Debugf("current block height is %s", height.String())
	return &height, nil
}

// Block returns information about a blockchain block by encoded hex number, or by a type tag.
// For tag based loading use predefined BlockType contacts.
func (rpc *Rpc) Block(numTag *string) (*types.Block, error) {
	// keep track of the operation
	rpc.log.Debugf("loading details of block num/tag %s", *numTag)

	// call for data
	var block types.Block
	err := rpc.ftm.Call(&block, "ftm_getBlockByNumber", numTag, false)
	if err != nil {
		rpc.log.Error("block could not be extracted")
		return nil, err
	}

	// detect block not found situation; block number is zero and the hash is also zero
	if uint64(block.Number) == 0 && block.Hash.Big().Cmp(big.NewInt(0)) == 0 {
		rpc.log.Debugf("block [%s] not found", *numTag)
		return nil, fmt.Errorf("block not found")
	}

	// keep track of the operation
	rpc.log.Debugf("block #%d found at mark %s",
		uint64(block.Number), time.Unix(int64(block.TimeStamp), 0).String())
	return &block, nil
}

// BlockByHash returns information about a blockchain block by hash.
func (rpc *Rpc) BlockByHash(hash *string) (*types.Block, error) {
	// keep track of the operation
	rpc.log.Debugf("loading details of block %s", *hash)

	// call for data
	var block types.Block
	err := rpc.ftm.Call(&block, "ftm_getBlockByHash", hash, false)
	if err != nil {
		rpc.log.Error("block could not be extracted")
		return nil, err
	}

	// detect block not found situation
	if uint64(block.Number) == 0 {
		rpc.log.Debugf("block [%s] not found", *hash)
		return nil, fmt.Errorf("block not found")
	}

	// inform and return
	rpc.log.Debugf("block #%d found at mark %s by hash %s",
		uint64(block.Number), time.Unix(int64(block.TimeStamp), 0).String(), *hash)
	return &block, nil
}
