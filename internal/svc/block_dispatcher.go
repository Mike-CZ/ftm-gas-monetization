// Package svc implements monitoring and scanning services of the API server.
package svc

import (
	"context"
	"fmt"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository/db"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/types"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/utils"
	"github.com/ethereum/go-ethereum/common"
	"time"
)

// blkDispatcher implements a service responsible for processing new blocks on the blockchain.
type blkDispatcher struct {
	service
	inBlock       chan *types.Block
	outDispatched chan uint64
	// topics represents a map of topics to their respective event handlers.
	topics map[common.Hash]EventHandler
	// watchedContracts represents a map of contracts to their respective project IDs.
	watchedContracts map[common.Address]int64
	// latestProcessedEpochId represents the latest epoch id.
	latestProcessedEpochId uint64
}

// name returns the name of the service used by orchestrator.
func (bld *blkDispatcher) name() string {
	return "block dispatcher"
}

// init prepares the block dispatcher to perform its function.
func (bld *blkDispatcher) init() {
	bld.sigStop = make(chan struct{})
	bld.outDispatched = make(chan uint64, blsBlockBufferCapacity)
	bld.initializeTopics()
	bld.initializeLastEpoch()
	bld.initializeProjects()
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
	if !bld.processTxs(blk) {
		return false
	}
	bld.log.Debugf("block #%d processed", blk.Number)
	// send the block number to the block scanner
	bld.outDispatched <- uint64(blk.Number)
	return true
}

// processTxs loops all the transactions in the block and process them.
func (bld *blkDispatcher) processTxs(blk *types.Block) bool {
	if uint64(blk.Epoch) <= bld.latestProcessedEpochId {
		bld.log.Debugf("block #%d is from an old epoch, skipping", blk.Number)
		return true
	}
	if blk.Txs == nil || len(blk.Txs) == 0 {
		bld.log.Debugf("empty block #%d processed", blk.Number)
		return true
	}
	// make backup of the contract list in case we need to rollback
	backupWatchedContracts := utils.CopyMap(bld.watchedContracts)

	// process all blockchain transactions in database transaction
	// to ensure all transactions are processed or none
	err := bld.repo.DatabaseTransaction(func(ctx context.Context, db *db.Db) error {
		for _, th := range blk.Txs {
			trx := bld.load(blk, th)
			if trx == nil {
				return fmt.Errorf("failed to load transaction %s", th.String())
			}
			if bld.checkContractAndFillTransaction(trx) {
				if err := db.StoreTransaction(ctx, trx); err != nil {
					return err
				}
			}
			// process logs
			if trx.Logs != nil && len(trx.Logs) > 0 {
				for _, log := range trx.Logs {
					// TODO: check if the log is from a contract we are interested in
					handler, ok := bld.topics[log.Topics[0]]
					if ok && log.BlockNumber == uint64(blk.Number) {
						bld.log.Infof("known topic %s found, processing", log.Topics[0].String())
						if err := handler(ctx, &log, db); err != nil {
							return err
						}
					}
				}
			}
		}
		// update last processed block number, so we can continue from here
		if err := db.UpdateLastProcessedBlock(ctx, uint64(blk.Number)); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		bld.log.Errorf("failed to process transactions; %s", err.Error())
		// restore the contract list
		bld.watchedContracts = backupWatchedContracts
		return false
	}
	return true
}

// checkContractAndFillTransaction checks if the transaction is related to a contract and fills the project id.
func (bld *blkDispatcher) checkContractAndFillTransaction(trx *types.Transaction) bool {
	if trx.From != nil && bld.watchedContracts[trx.From.Address] > 0 {
		trx.ProjectId = bld.watchedContracts[trx.From.Address]
		return true
	}
	if trx.To != nil && bld.watchedContracts[trx.To.Address] > 0 {
		trx.ProjectId = bld.watchedContracts[trx.To.Address]
		return true
	}
	return false
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
	trx.Timestamp = time.Unix(int64(blk.TimeStamp), 0)
	return trx
}

// initializeProjects initializes the list of watched contracts.
func (bld *blkDispatcher) initializeProjects() {
	bld.watchedContracts = make(map[common.Address]int64)
	// get all active projects
	pq := bld.repo.ProjectQuery()
	projects, err := pq.GetAll()
	if err != nil {
		bld.log.Fatal("failed to get active projects: %v", err)
	}
	// get slice of project ids
	ids := utils.Map(projects, func(p *types.Project) int64 { return p.Id })
	// get all enabled contracts for the active projects
	pcq := bld.repo.ProjectContractQuery()
	contracts, err := pcq.WhereIsEnabled(true).WhereProjectIdIn(ids).GetAll()
	if err != nil {
		bld.log.Fatal("failed to get project contracts: %v", err)
	}
	// initialize the map
	for _, c := range contracts {
		bld.watchedContracts[c.Address.Address] = c.ProjectId
	}
}

// initializeLastEpoch initializes the last processed epoch.
func (bld *blkDispatcher) initializeLastEpoch() {
	epoch, err := bld.repo.LastProcessedEpoch()
	if err != nil {
		bld.log.Fatal("failed to get last processed epoch: %v", err)
	}
	bld.latestProcessedEpochId = epoch
}
