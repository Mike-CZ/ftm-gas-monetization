package svc

import (
	"github.com/Mike-CZ/ftm-gas-monetization/internal/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"time"
)

// blsObserverTickBaseDuration represents the frequency of the scanner status observer.
const blsObserverTickBaseDuration = 5 * time.Second

// blsScanTickBaseDuration represents the frequency of the scanner default progress.
const blsScanTickBaseDuration = 5 * time.Millisecond

// blsObserverTickIdleDuration represents the frequency of the scanner status observer on idle.
const blsObserverTickIdleDuration = 1 * time.Minute

// blsScanTickIdleDuration represents the frequency of the scanner re-check on idle.
const blsScanTickIdleDuration = 5 * time.Minute

// blsBlockBufferCapacity represents the capacity of the found blocks channel.
// When the channel is full, the push will have to wait for room here and the scanner
// will be slowed down naturally.
const blsBlockBufferCapacity = 1000

// blsReScanHysteresis is the number of blocks we wait from dispatcher until a re-scan kicks in.
const blsReScanHysteresis = 100

// blkScanner represents a scanner of historical data from the blockchain.
type blkScanner struct {
	service
	outBlock       chan *types.Block
	outStateSwitch chan bool
	inDispatched   chan uint64
	observeTick    *time.Ticker
	scanTick       *time.Ticker
	onIdle         bool
	from           uint64
	next           uint64
	to             uint64
	done           uint64
}

// init initializes the block scanner and registers it with the manager.
func (bls *blkScanner) init() {
	bls.onIdle = false
	bls.sigStop = make(chan struct{})
	bls.outStateSwitch = make(chan bool, 1)
	bls.outBlock = make(chan *types.Block, blsBlockBufferCapacity)
}

// run scans past blocks one by one until it reaches top
// after the top is reached, it idles and checks the head state to make sure
// the API server keeps up with the most recent block
func (bls *blkScanner) run() {
	// get the scanner start block
	start, err := bls.startBlock()
	if err != nil {
		bls.log.Errorf("scanner can not proceed; %s", err.Error())
		return
	}

	// signal orchestrator we started and go
	bls.log.Noticef("block scan starts at #%d", start)
	bls.from = start
	bls.next = start

	bls.mgr.started(bls)
	go bls.execute()
}

// name provides the name of the service.
func (bls *blkScanner) name() string {
	return "block scanner"
}

// close signals the block scanner to terminate
func (bls *blkScanner) close() {
	if bls.scanTick != nil {
		bls.scanTick.Stop()
		bls.observeTick.Stop()
	}
	if bls.sigStop != nil {
		close(bls.sigStop)
	}
}

// execute scans blockchain blocks in the given range and push found blocks
// to the output channel for processing.
func (bls *blkScanner) execute() {
	defer func() {
		close(bls.outBlock)
		close(bls.outStateSwitch)
		bls.mgr.finished(bls)
	}()

	// set initial state and start the tickers for observer and scanner
	bls.observe()
	bls.observeTick = time.NewTicker(blsObserverTickBaseDuration)
	bls.scanTick = time.NewTicker(blsScanTickBaseDuration)

	// do the scan
	for {
		select {
		case <-bls.sigStop:
			return
		case bin, ok := <-bls.inDispatched:
			// ignore block re-scans; do not skip blocks in dispatched # counter
			if ok && (bls.done == 0 || int64(bin)-int64(bls.done) == 1) {
				bls.done = bin
			}
		case <-bls.observeTick.C:
			bls.updateState(bls.observe())
		case <-bls.scanTick.C:
			bls.shift()
		}
	}
}

// observe updates the scanner final block and logs the progress.
// It returns expected idle state to be used to transition if needed.
func (bls *blkScanner) observe() bool {
	// try to get the block height
	bh, err := bls.repo.BlockHeight()
	if err != nil {
		bls.log.Errorf("can not get current block height; %s", err.Error())
		return false
	}

	// if on idle, wait for the dispatcher to catch up with the blocks
	// we use a hysteresis to delay state flip back to active scan
	// we compare current block height with the latest known dispatched block number
	target := bh.ToInt().Uint64()
	if bls.onIdle && target < bls.done+blsReScanHysteresis {
		bls.next = bls.done
		bls.from = bls.done
		bls.log.Infof("block scanner idling at #%d, head at #%d", bls.next, target)
		return true
	}

	// adjust target block number; log the progress of the scan
	bls.to = target
	bls.log.Infof("block scanner at #%d of <#%d, #%d>, #%d dispatched", bls.next, bls.from, bls.to, bls.done)
	return bls.to < bls.next
}

// updateState change scanner state if needed.
// It resets the internal tickers according to the target state.
func (bls *blkScanner) updateState(target bool) {
	// if the state already match, do nothing
	if target == bls.onIdle {
		return
	}

	// switch the state; advertise the transition
	bls.log.Noticef("block scanner idle state toggled to %t", target)
	bls.onIdle = target

	select {
	case bls.outStateSwitch <- target:
	case <-bls.sigStop:
		return
	}

	// going full speed
	if !target {
		bls.observeTick.Reset(blsObserverTickBaseDuration)
		bls.scanTick.Reset(blsScanTickBaseDuration)
		return
	}

	// going idle
	bls.observeTick.Reset(blsObserverTickIdleDuration)
	bls.scanTick.Reset(blsScanTickIdleDuration)
}

// startBlock provides the starting block for the scanner
func (bls *blkScanner) startBlock() (uint64, error) {
	lb, err := bls.repo.LastProcessedBlock()
	if err != nil {
		bls.log.Criticalf("can not pull last seen block; %s", err.Error())
		return 0, err
	}
	return lb, nil
}

// shift pulls the next block if available and pushes it for processing.
func (bls *blkScanner) shift() {
	// we may not need to pull at all, if on updateState
	if bls.onIdle {
		return
	}

	// are we at the end? check the status
	if bls.next > bls.to {
		bls.updateState(bls.observe())
		return
	}

	// pull the current block
	block, err := bls.repo.BlockByNumber((*hexutil.Uint64)(&bls.next))
	if err != nil {
		bls.log.Errorf("block #%d not available; %s", bls.next, err.Error())
		return
	}

	// push the block for processing and advance to the next expected block
	// observe possible stop signal during a wait for the block queue slot
	select {
	case bls.outBlock <- block:
		bls.next++
	case <-bls.sigStop:
	}
}
