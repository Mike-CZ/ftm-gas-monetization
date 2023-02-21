package svc

import (
	"github.com/ethereum/go-ethereum/core/types"
	"time"
)

const (
	// blockQueueCapacity represents the capacity of block headers processor queue
	blockQueueCapacity = 5000

	// scanTickFrequency is the time delay for the regular scanner
	scanTickFrequency = 4 * time.Millisecond

	// topUpdateTickFrequency is the time delay for the block head update
	topUpdateTickFrequency = 5 * time.Second
)

// blkScanner represents a scanner of historical data from the blockchain.
type blkScanner struct {
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
func newBlkScanner(mgr *Manager) *blkScanner {
	return &blkScanner{
		mgr:       mgr,
		sigStop:   make(chan bool, 1),
		outBlocks: make(chan *types.Header, blockQueueCapacity),
	}
}

// init initializes the block scanner and registers it with the manager.
func (bs *blkScanner) init() {
	//bs.inObservedBlocks = bs.mgr.logObserver.outObservedBlocks
	//bs.inRescanBlocks = bs.mgr.collectionValidator.outRescanQueue

	// TODO: bs.current, bs.target = bs.start(), bs.top()
	bs.current = 1
	bs.target = 5_000_000

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
			//bs.target = bs.top()
			//bs.notify()

		case bid, ok := <-bs.inObservedBlocks:
			if !ok {
				return
			}

			// we just casually follow the chain head
			//if bs.state == blkIsIdling && bid > bs.current {
			//	bs.current = bid
			//	bs.lastProcessedBlock = bid
			//	continue
			//}

			// we rush to catch the head, so we don't accept processed blocks above scanner head
			if bid > bs.lastProcessedBlock && bid <= bs.current {
				bs.lastProcessedBlock = bid
			}

		//case bid, ok := <-bs.inRescanBlocks:
		//	if !ok {
		//		return
		//	}
		//	bs.rescan(bid)

		case <-bs.scanTicker.C:
		}

		//bs.next()
		//bs.checkTarget()
		//bs.checkIdle()
	}
}
