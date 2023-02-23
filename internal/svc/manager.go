package svc

import (
	"github.com/Mike-CZ/ftm-gas-monetization/internal/logger"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository"
	"sync"
)

// Manager represents the manager controlling services lifetime.
type Manager struct {
	wg  *sync.WaitGroup
	svc []service
	log *logger.AppLogger

	// managed services
	blkScanner  *blkScanner
	blkObserver *blkObserver
}

func New(repo *repository.Repository, log *logger.AppLogger) *Manager {
	// prep the manager
	mgr := Manager{
		wg:  new(sync.WaitGroup),
		svc: make([]service, 0),
		log: log.ModuleLogger("svc_manager"),
	}

	// make services
	mgr.blkScanner = newBlkScanner(&mgr, repo, log)
	mgr.blkObserver = newBlkObserver(&mgr, repo, log)

	// init and run
	mgr.init()
	log.Notice("all services are running")

	// TODO: remove this
	mgr.wg.Wait()

	return &mgr
}

// init initializes the services in the correct order.
func (mgr *Manager) init() {
	mgr.blkScanner.init()
}

// add managed service instance to the Manager and run it.
func (mgr *Manager) add(s service) {
	// keep track of running services
	mgr.svc = append(mgr.svc, s)

	// run the service
	mgr.wg.Add(1)
	go s.run()
	mgr.log.Noticef("service %s started", s.name())
}

// closed signals the manager a service terminated.
func (mgr *Manager) closed(s service) {
	mgr.wg.Done()
	mgr.log.Noticef("service %s stopped", s.name())
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
