package svc

import (
	"github.com/Mike-CZ/ftm-gas-monetization/internal/logger"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository"
	"sync"
)

// Manager represents the manager controlling services lifetime.
type Manager struct {
	repo *repository.Repository
	wg   *sync.WaitGroup
	svc  []serviceHandler
	log  *logger.AppLogger

	// managed services
	blkScanner    *blkScanner
	blkDispatcher *blkDispatcher
}

func New(repo *repository.Repository, log *logger.AppLogger) *Manager {
	// prep the manager
	mgr := Manager{
		repo: repo,
		wg:   new(sync.WaitGroup),
		svc:  make([]serviceHandler, 0),
		log:  log.ModuleLogger("svc_manager"),
	}
	mgr.init()
	return &mgr
}

// Run starts all the services prepared to be run.
func (mgr *Manager) Run() {
	// init all the services to the starting state
	for _, s := range mgr.svc {
		s.init()
	}

	// connect services' input channels to their source
	mgr.blkDispatcher.inBlock = mgr.blkScanner.outBlock
	mgr.blkScanner.inDispatched = mgr.blkDispatcher.outDispatched

	// start services
	for _, s := range mgr.svc {
		s.run()
	}

	mgr.wg.Wait()
}

// Close terminates the service manager
// and all the managed services along with it.
func (mgr *Manager) Close() {
	mgr.log.Notice("services are being terminated")

	for _, s := range mgr.svc {
		mgr.log.Noticef("closing %s", s.name())
		s.close()
	}

	mgr.wg.Wait()
	mgr.log.Notice("services closed")
}

// init initializes the services in the correct order.
func (mgr *Manager) init() {
	// make services
	mgr.blkScanner = &blkScanner{
		service: service{
			repo: mgr.repo,
			log:  mgr.log.ModuleLogger("blk_scanner"),
			mgr:  mgr,
		},
	}
	mgr.svc = append(mgr.svc, mgr.blkScanner)

	lastEpoch, err := mgr.repo.LastProcessedEpoch()
	if err != nil {
		mgr.log.Fatalf("can not get last processed epoch. Err: %v", err)
	}

	mgr.blkDispatcher = &blkDispatcher{
		lastProcessedEpoch: lastEpoch,
		service: service{
			repo: mgr.repo,
			log:  mgr.log.ModuleLogger("blk_scanner"),
			mgr:  mgr,
		},
	}
	mgr.svc = append(mgr.svc, mgr.blkDispatcher)
}

// started signals to the manager that the calling service
// has been started and is functioning.
func (mgr *Manager) started(svc serviceHandler) {
	mgr.wg.Add(1)
	mgr.log.Noticef("%s is running", svc.name())
}

// finished signals to the manager that the calling service
// has been terminated and is no longer running.
func (mgr *Manager) finished(svc serviceHandler) {
	mgr.wg.Done()
	mgr.log.Noticef("%s terminated", svc.name())
}
