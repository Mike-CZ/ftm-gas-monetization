package svc

import (
	"context"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository/db"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/types"
	"github.com/ethereum/go-ethereum/common"
)

// contract implements a service responsible for processing new blocks on the blockchain.
type contractApprover struct {
	service
	inContract        chan *types.ProjectContract
	removeContractCh  chan *types.ProjectContract
	whiteListedCh     chan whiteListed
	approvedContracts map[common.Address]bool
}

type whiteListed struct {
	addr   common.Address
	answer chan bool
}

// name returns the name of the service used by orchestrator.
func (appr *contractApprover) name() string {
	return "contract approver"
}

// init prepares the contract approver to perform its function.
func (appr *contractApprover) init() {
	appr.removeContractCh = make(chan *types.ProjectContract)
	appr.sigStop = make(chan struct{})
	appr.whiteListedCh = make(chan whiteListed)
	appr.fetchApprovedContractsFromDb()
}

// run starts the block dispatcher
func (appr *contractApprover) run() {
	// signal orchestrator we started and go
	appr.mgr.started(appr)
	go appr.execute()
}

// execute collects contracts from the input channel
// and processes them.
func (appr *contractApprover) execute() {
	// make sure to clean up
	defer func() {
		close(appr.removeContractCh)
		close(appr.whiteListedCh)
		appr.mgr.finished(appr)
	}()

	for {
		select {
		case <-appr.sigStop:
			return
		case ctr, ok := <-appr.inContract:
			if !ok {
				appr.log.Criticalf("store contract channel closed, terminating %v", appr.name())
				return
			}

			appr.log.Debugf("new contract %v arrived", ctr.Address)
			if !appr.storeContract(ctr) {
				continue
			}
		case ctr, ok := <-appr.removeContractCh:
			if !ok {
				appr.log.Criticalf("remove contract channel closed, terminating %v", appr.name())
				return
			}

			appr.log.Debugf("removing contract %v from white list", ctr.Address)
			if !appr.removeContact(ctr) {
				continue
			}
		case wl, ok := <-appr.whiteListedCh:
			if !ok {
				appr.log.Criticalf("white listed channel closed, terminating %v", appr.name())
				return
			}

			// is contract within the white-list
			_, approved := appr.approvedContracts[wl.addr]
			wl.answer <- approved

		}
	}
}

// isContractApproved returns whether contract is in the white list.
func (appr *contractApprover) isContractApproved(ctrAddr common.Address) bool {
	vessel := whiteListed{
		addr:   ctrAddr,
		answer: make(chan bool, 1),
	}

	appr.whiteListedCh <- vessel

	return <-vessel.answer
}

// storeContract into db and map since it is eligible for gas monetization
func (appr *contractApprover) storeContract(ctr *types.ProjectContract) bool {
	err := appr.repo.DatabaseTransaction(func(ctx context.Context, db *db.Db) error {
		if err := db.StoreContract(ctx, ctr); err != nil {
			return err
		}

		// add ctr to the map for faster fetching
		appr.approvedContracts[ctr.Address] = true
		return nil
	})

	if err != nil {
		appr.log.Errorf("can not storeContract. Err: %v", err)
		return false
	}

	return true
}

// removeContact from db and map since it is not eligible for gas monetization anymore.
func (appr *contractApprover) removeContact(ctr *types.ProjectContract) bool {
	if _, ok := appr.approvedContracts[ctr.Address]; !ok {
		appr.log.Debugf("trying to remove non-approved contract %v", ctr.Address)
		return true
	}

	// remove ctr from the map
	delete(appr.approvedContracts, ctr.Address)

	err := appr.repo.DatabaseTransaction(func(ctx context.Context, db *db.Db) error {
		if err := db.RemoveContract(ctx, ctr); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		appr.log.Errorf("can not removeContract. Err: %v", err)
		return false
	}

	return true
}

// fetchApprovedContractsFromDb and store them into the map when app is restarted.
func (appr *contractApprover) fetchApprovedContractsFromDb() {
	err := appr.repo.DatabaseTransaction(func(ctx context.Context, db *db.Db) error {
		var err error
		appr.approvedContracts, err = db.FetchApprovedContracts(ctx)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		appr.log.Fatalf("can not fetchApprovedContractsFromDb. Err: %v", err)
	}

}
