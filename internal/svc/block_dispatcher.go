// Package svc implements monitoring and scanning services of the API server.
package svc

import (
	"context"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository/db"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"time"
)

// blkDispatcher implements a service responsible for processing new blocks on the blockchain.
type blkDispatcher struct {
	service
	inBlock       chan *types.Block
	outDispatched chan uint64
}

// name returns the name of the service used by orchestrator.
func (bld *blkDispatcher) name() string {
	return "block dispatcher"
}

// init prepares the block dispatcher to perform its function.
func (bld *blkDispatcher) init() {
	bld.sigStop = make(chan struct{})
	bld.outDispatched = make(chan uint64, blsBlockBufferCapacity)
}

// run starts the block dispatcher
func (bld *blkDispatcher) run() {
	// signal orchestrator we started and go
	bld.mgr.started(bld)
	go bld.execute()
}

// execute collects blocks from an input channel
// and processes them.
func (bld *blkDispatcher) execute() {
	// make sure to clean up
	defer func() {
		// close our channels
		close(bld.outDispatched)

		// signal we are done
		bld.mgr.finished(bld)
	}()

	// loop here
	for {
		select {
		case <-bld.sigStop:
			return

		case blk, ok := <-bld.inBlock:
			// do we have a working channel?
			if !ok {
				bld.log.Notice("block channel closed, terminating %s", bld.name())
				return
			}

			// process the new block
			bld.log.Debugf("block #%d arrived", uint64(blk.Number))
			if !bld.process(blk) {
				continue
			}
		}
	}
}

// process the given block by loading its content and sending block transactions
// into the trx dispatcher. Observe terminate signal.
func (bld *blkDispatcher) process(blk *types.Block) bool {
	// dispatched block number is used by the block scanner
	// to keep track of the work done vs. work pending
	select {
	case bld.outDispatched <- uint64(blk.Number):
	case <-bld.sigStop:
		return false
	}

	if blk.Txs == nil || len(blk.Txs) == 0 {
		bld.log.Debugf("empty block #%d processed", blk.Number)
		return true
	}

	bld.log.Debugf("%d transaction found in block #%d", len(blk.Txs), blk.Number)

	if !bld.processTxs(blk) {
		return false
	}

	bld.log.Debugf("block #%d processed", blk.Number)
	return true
}

// processTxs loops all the transactions in the block and process them.
func (bld *blkDispatcher) processTxs(blk *types.Block) bool {
	// process all transactions in database transaction to ensure
	// all transactions are processed or none
	err := bld.repo.DatabaseTransaction(func(ctx context.Context, db *db.Db) error {
		for _, th := range blk.Txs {
			trx := bld.load(blk, th)
			//TODO: What to do when transaction fails to load?
			if trx != nil {
				if err := db.StoreTransaction(ctx, trx); err != nil {
					return err
				}
			}
		}

		var currentEpoch hexutil.Uint64
		var err error

		if currentEpoch, err = db.LastProcessedEpoch(ctx); err != nil {
			return err
		}

		// update epoch number
		if blk.Epoch > currentEpoch {
			if err = db.UpdateLastProcessedEpoch(ctx, blk.Epoch); err != nil {
				return err
			}
		}

		// update last processed block number, so we can continue from here
		if err = db.UpdateLastProcessedBlock(ctx, uint64(blk.Number)); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		bld.log.Errorf("failed to process transactions; %s", err.Error())
		return false
	}

	return true
}

// load a transaction detail from repository, if possible.
func (bld *blkDispatcher) load(blk *types.Block, th *common.Hash) *types.Transaction {
	// get transaction
	trx, err := bld.repo.Transaction(th)
	if err != nil {
		bld.log.Errorf("transaction %s detail not available; %s", th.String(), err.Error())
		return nil
	}

	// update time stamp using the block data
	trx.TimeStamp = time.Unix(int64(blk.TimeStamp), 0)
	return trx
}
