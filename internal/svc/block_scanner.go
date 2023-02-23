package svc

import (
	"github.com/Mike-CZ/ftm-gas-monetization/internal/logger"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository"
	"github.com/ethereum/go-ethereum/core/types"
	"time"
)

const (
	// blkIsScanning represents the state of active block scanning
	blkIsScanning = iota

	// blkIsIdling represents the state of passive head checks
	blkIsIdling

	// blockQueueCapacity represents the capacity of block headers processor queue
	blockQueueCapacity = 5000

	// scanTickFrequency is the time delay for the regular scanner
	scanTickFrequency = 4 * time.Millisecond

	// idleTickFrequency is the time delay for the regular scanner
	idleTickFrequency = 1 * time.Second

	// topUpdateTickFrequency is the time delay for the block head update
	topUpdateTickFrequency = 5 * time.Second

	// blkScannerHysteresis represent the number of blocks we let slide
	// until we switch back to active scan state.
	blkScannerHysteresis = 10
)

// blkScanner represents a scanner of historical data from the blockchain.
type blkScanner struct {
	repo *repository.Repository
	log  *logger.AppLogger

	// mgr represents the Manager instance
	mgr *Manager

	// sigStop represents the signal for closing the router
	sigStop chan bool

	// inObservedBlocks is a channel receiving IDs of observed blocks
	// we track the observed heads to recognize if we need to switch back to scan from idle
	inObservedBlocks chan uint64

	// inRescanBlocks is a channel receiving re-scan requests from given block number
	inRescanBlocks chan uint64

	// outBlocks represents a channel fed with past block headers being scanned.
	outBlocks chan *types.Header

	// scanTicker represents the ticker for the scanner
	scanTicker *time.Ticker

	// state represents the current state of the scanner
	// it's scanning by default and turns idle later, if not needed
	state int

	// current represents the ID of the currently processed block
	current uint64

	// target represents the ID we need to reach
	target uint64

	// lastProcessedBlock represents the ID of the last processed block notified to us
	lastProcessedBlock uint64
}

// newBlkScanner creates a new instance of the block scanner service.
func newBlkScanner(mgr *Manager, repo *repository.Repository, log *logger.AppLogger) *blkScanner {
	return &blkScanner{
		repo:      repo,
		log:       log.ModuleLogger("block_scanner"),
		mgr:       mgr,
		sigStop:   make(chan bool, 1),
		outBlocks: make(chan *types.Header, blockQueueCapacity),
	}
}

// init initializes the block scanner and registers it with the manager.
func (bs *blkScanner) init() {
	//bs.inObservedBlocks = bs.mgr.logObserver.outObservedBlocks
	//bs.inRescanBlocks = bs.mgr.collectionValidator.outRescanQueue

	bs.current = bs.startBlock()
	bs.target = bs.targetBlock()

	bs.mgr.add(bs)
}

// name provides the name of the service.
func (bs *blkScanner) name() string {
	return "block scanner"
}

// close signals the block observer to terminate
func (bs *blkScanner) close() {
	bs.sigStop <- true
}

// run scans past blocks one by one until it reaches top
// after the top is reached, it idles and checks the head state to make sure
// the API server keeps up with the most recent block
func (bs *blkScanner) run() {
	// make tickers
	topTick := time.NewTicker(topUpdateTickFrequency)
	bs.scanTicker = time.NewTicker(scanTickFrequency)

	// make sure to stop the tickers and notify the manager
	defer func() {
		topTick.Stop()
		bs.scanTicker.Stop()
		bs.mgr.closed(bs)
	}()

	for {
		// make sure to check for terminate; but do not stay in
		select {
		case <-bs.sigStop:
			return

		case <-topTick.C:
			bs.target = bs.targetBlock()
			bs.updateLastBlock()

		case bid, ok := <-bs.inObservedBlocks:
			if !ok {
				return
			}

			bs.log.Noticef("observed block #%d", bid)

			// we just casually follow the chain head
			if bs.state == blkIsIdling && bid > bs.current {
				bs.current = bid
				bs.lastProcessedBlock = bid
				continue
			}

			// we rush to catch the head, so we don't accept processed blocks above scanner head
			if bid > bs.lastProcessedBlock && bid <= bs.current {
				bs.lastProcessedBlock = bid
			}

		case bid, ok := <-bs.inRescanBlocks:
			if !ok {
				return
			}
			bs.rescan(bid)

		case <-bs.scanTicker.C:
		}

		bs.next()
		bs.checkTarget()
		bs.checkIdle()
	}
}

// startBlock provides the starting block for the scanner
func (bs *blkScanner) startBlock() uint64 {
	lb, err := bs.repo.LastBlock()
	if err != nil {
		bs.log.Criticalf("can not pull last seen block; %s", err.Error())
		return 0
	}

	if lb <= blkScannerHysteresis {
		return lb
	}

	return lb - blkScannerHysteresis
}

// targetBlock provides the number of the target block for the scanner.
func (bs *blkScanner) targetBlock() uint64 {
	cur, err := bs.repo.CurrentHead()
	if err != nil {
		bs.log.Criticalf("can not pull the latest head number; %s", err.Error())
		return 0
	}

	bs.log.Noticef("target block is #%d", cur)

	return cur
}

// updateLastBlock updates last seen block in repository, if any.
func (bs *blkScanner) updateLastBlock() {
	if bs.lastProcessedBlock == 0 {
		return
	}
	_ = bs.repo.UpdateLastBlock(bs.lastProcessedBlock)

	if bs.state == blkIsIdling {
		bs.log.Noticef("idle at #%d, head at #%d", bs.current, bs.target)
		return
	}
	bs.log.Noticef("scanner at #%d of #%d; processed #%d", bs.current, bs.target, bs.lastProcessedBlock)
}

// next tries to advance the scanner to the next block, if possible
func (bs *blkScanner) next() {
	if bs.state == blkIsScanning && bs.current <= bs.target {
		hdr, err := bs.repo.GetHeader(bs.current)
		if err != nil {
			bs.log.Errorf("block header #%s not available; %s", bs.current, err.Error())
			return
		}
		// send the block to the observer; make sure not to miss stop signal
		select {
		case bs.outBlocks <- hdr:
			bs.current += 1
		case <-bs.sigStop:
			bs.sigStop <- true
		}
	}
}

// rescan the blockchain from the given block, if relevant.
func (bs *blkScanner) rescan(from uint64) {
	// are we already on the track?
	if from > bs.current {
		return
	}

	// refresh target block
	bs.target = bs.targetBlock()

	// we know from <= current here; start at least <blkScannerHysteresis> back
	diff := bs.current - from
	if diff < blkScannerHysteresis {
		bs.current = bs.current - blkScannerHysteresis
		return
	}
	bs.current = from
}

// checkTarget checks if the scanner reached designated target head.
func (bs *blkScanner) checkTarget() {
	// reached target? make sure we are on target; switch state if so
	if bs.state == blkIsScanning && bs.current > bs.target {
		bs.target = bs.targetBlock()
		diff := int64(bs.target) - int64(bs.current)

		if diff <= 0 {
			bs.state = blkIsIdling
			bs.scanTicker.Reset(idleTickFrequency)
			bs.log.Noticef("scanner idling since #%d", bs.current)
		}
	}
}

// checkIdle checks if the idle state should be switched back to active scan.
func (bs *blkScanner) checkIdle() {
	if bs.state != blkIsIdling {
		return
	}

	diff := int64(bs.target) - int64(bs.current)
	if diff >= blkScannerHysteresis {
		bs.state = blkIsScanning
		bs.scanTicker.Reset(scanTickFrequency)
		bs.log.Noticef("scanner head at #%d of #%d with %d diff", bs.current, bs.target, diff)
	}
}
