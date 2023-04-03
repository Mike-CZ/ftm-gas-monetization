// Package svc implements monitoring and scanning services of the API server.
package svc

import (
	"context"
	"fmt"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository/db"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"math/big"
	"time"
)

// rewardPercentage represents the percentage of the reward to be paid to the project.
const rewardsPercentage = 15

// blkDispatcher implements a service responsible for processing new blocks on the blockchain.
type blkDispatcher struct {
	service
	inBlock       chan *types.Block
	outDispatched chan uint64
	// topics represents a map of topics to their respective event handlers.
	topics map[common.Hash]EventHandler
	// watchedContracts represents a map of contracts to their respective project instances.
	watchedContracts map[common.Address]*types.Project
	// watchedProjectIds represents a map of projects where key is `project_id` provided by contract.
	watchedProjectIds map[uint64]*types.Project
	// currentEpochId represents the current epoch id.
	currentEpochId uint64
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
	bld.initializeTrackedData()
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
	if uint64(blk.Epoch) < bld.currentEpochId {
		bld.log.Debugf("block #%d is from an old epoch, skipping", blk.Number)
		return true
	}
	if blk.Txs == nil || len(blk.Txs) == 0 {
		bld.log.Debugf("empty block #%d processed", blk.Number)
		return true
	}
	// process all data in database transaction to ensure all transactions are processed or none
	err := bld.repo.DatabaseTransaction(func(ctx context.Context, db *db.Db) error {
		for _, th := range blk.Txs {
			trx := bld.load(blk, th)
			if trx == nil {
				return fmt.Errorf("failed to load transaction %s", th.String())
			}
			// we moved to a new epoch, store the previous one
			if uint64(blk.Epoch) > bld.currentEpochId {
				if err := bld.storePreviousEpoch(ctx, db, uint64(blk.Epoch)); err != nil {
					return fmt.Errorf("failed to store previous epoch: %s", err.Error())
				}
			}
			// store transaction into database
			trx.Epoch = blk.Epoch
			if err := bld.storeTransaction(ctx, db, trx); err != nil {
				return fmt.Errorf("failed to store transaction: %s", err.Error())
			}
			// process logs
			if trx.Logs != nil && len(trx.Logs) > 0 {
				for _, log := range trx.Logs {
					if log.Address != bld.repo.GasMonetizationAddress() {
						continue
					}
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
		// reinitialize the data, because we might have a corrupted state
		bld.initializeTrackedData()
		return false
	}
	return true
}

// storePreviousEpoch stores the previous epoch data in the database.
func (bld *blkDispatcher) storePreviousEpoch(ctx context.Context, db *db.Db, newEpochId uint64) error {
	// update the current epoch id
	if err := db.UpdateCurrentEpoch(ctx, newEpochId); err != nil {
		return err
	}
	// map for temporarily storing projects to be updated
	projects := make(map[int64]*types.Project)
	var transactionsCount uint64 = 0
	totalCollected := big.NewInt(0)
	// get all transactions for the previous epoch and update generated rewards and number of transactions
	tq := db.TransactionQuery(ctx)
	txs, err := tq.WhereEpoch(bld.currentEpochId).GetAll()
	if err != nil {
		return err
	}
	// loop all transactions from the previous epoch and update data
	for _, trx := range txs {
		transactionsCount += 1
		totalCollected = totalCollected.Add(totalCollected, trx.RewardToClaim.ToInt())
		project, exists := projects[trx.ProjectId]
		if !exists {
			pq := db.ProjectQuery(ctx)
			project, err = pq.WhereId(trx.ProjectId).GetFirstOrFail()
			if err != nil {
				return err
			}
			// if project is watched, take it from the map so the fields are updated for log handler
			watched, isWatched := bld.watchedProjectIds[project.ProjectId]
			if isWatched {
				project = watched
			}
			projects[trx.ProjectId] = project
		}
		// increase collected amount
		if project.CollectedRewards == nil {
			project.CollectedRewards = trx.RewardToClaim
		} else {
			res := new(big.Int).Add(project.CollectedRewards.ToInt(), trx.RewardToClaim.ToInt())
			project.CollectedRewards = &types.Big{Big: hexutil.Big(*res)}
		}
		// increase rewards to claim
		if project.RewardsToClaim == nil {
			project.RewardsToClaim = trx.RewardToClaim
		} else {
			res := new(big.Int).Add(project.RewardsToClaim.ToInt(), trx.RewardToClaim.ToInt())
			project.RewardsToClaim = &types.Big{Big: hexutil.Big(*res)}
		}
		// increase number of transactions
		project.TransactionsCount += 1
	}
	// increase the total amount collected
	if totalCollected.Cmp(big.NewInt(0)) > 0 {
		if err = db.IncreaseTotalAmountCollected(ctx, totalCollected); err != nil {
			return err
		}
	}
	// update the number of transactions
	if transactionsCount > 0 {
		if err = db.IncreaseTotalTransactionsCount(ctx, transactionsCount); err != nil {
			return err
		}
	}
	// loop through all projects and update the data
	for _, project := range projects {
		if err = db.UpdateProject(ctx, project); err != nil {
			return err
		}
	}
	// set the new epoch id
	bld.currentEpochId = newEpochId
	return nil
}

// storeTransaction stores a transaction in the repository.
func (bld *blkDispatcher) storeTransaction(ctx context.Context, db *db.Db, trx *types.Transaction) error {
	traceResult, err := bld.repo.TraceTransaction(trx.Hash.Hash)
	if err != nil {
		return err
	}
	if traceResult == nil || len(traceResult) == 0 {
		return nil
	}
	// on first position is root transaction, if transaction failed, we don't need to process it further
	if traceResult[0].Error != nil {
		return nil
	}

	for _, trace := range traceResult {

	}

	if trx.From != nil && bld.watchedContracts[trx.From.Address] != nil {
		trx.ProjectId = bld.watchedContracts[trx.From.Address].Id
	}
	if trx.To != nil && bld.watchedContracts[trx.To.Address] != nil {
		trx.ProjectId = bld.watchedContracts[trx.To.Address].Id
	}

	total := new(big.Int).Mul(trx.GasPrice.ToInt(), new(big.Int).SetUint64(uint64(*trx.GasUsed)))
	reward := new(big.Int).Mul(total, big.NewInt(rewardsPercentage))
	finalReward := new(big.Int).Div(reward, big.NewInt(100))
	trx.RewardToClaim = &types.Big{Big: hexutil.Big(*finalReward)}
	if err := db.StoreTransaction(ctx, trx); err != nil {
		return err
	}

	bld.repo.TraceTransaction(trx.Hash.Hash)

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

// initializeCurrentEpoch initializes the current epoch.
func (bld *blkDispatcher) initializeCurrentEpoch() {
	epoch, err := bld.repo.CurrentEpoch()
	if err != nil {
		bld.log.Fatal("failed to get current epoch: %v", err)
	}
	bld.currentEpochId = epoch
}

// initializeProjects initializes the list of watched projects.
func (bld *blkDispatcher) initializeProjects() {
	bld.watchedContracts = make(map[common.Address]*types.Project)
	bld.watchedProjectIds = make(map[uint64]*types.Project)
	// get all active projects
	pq := bld.repo.ProjectQuery()
	projects, err := pq.WhereActiveInEpoch(bld.currentEpochId).GetAll()
	if err != nil {
		bld.log.Fatal("failed to get active projects: %v", err)
	}
	for _, project := range projects {
		pcq := bld.repo.ProjectContractQuery()
		contracts, err := pcq.WhereIsApproved(true).WhereProjectId(project.Id).GetAll()
		if err != nil {
			bld.log.Fatal("failed to get project contracts: %v", err)
		}
		for _, c := range contracts {
			// store reference to the project for fast lookups
			bld.watchedContracts[c.Address.Address] = &project
			bld.watchedProjectIds[project.ProjectId] = &project
		}
	}
}

// initializeTrackedData initializes the data tracked by the block dispatcher.
func (bld *blkDispatcher) initializeTrackedData() {
	bld.initializeCurrentEpoch()
	bld.initializeProjects()
}
