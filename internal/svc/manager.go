package svc

import (
	"ftm-gas-monetization/internal/config"
	"ftm-gas-monetization/internal/logger"
	"ftm-gas-monetization/internal/notifier"
	"ftm-gas-monetization/internal/repository"
	"sync"
)

// Manager represents the manager controlling services lifetime.
type Manager struct {
	cfg  *config.Config
	repo *repository.Repository
	wg   *sync.WaitGroup
	svc  []serviceHandler
	log  *logger.AppLogger

	// managed services
	blkScanner    *blkScanner
	blkDispatcher *blkDispatcher
}

func New(cfg *config.Config, repo *repository.Repository, log *logger.AppLogger) *Manager {
	// prep the manager
	mgr := Manager{
		cfg:  cfg,
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

	mgr.blkDispatcher = &blkDispatcher{
		service: service{
			repo: mgr.repo,
			log:  mgr.log.ModuleLogger("blk_scanner"),
			mgr:  mgr,
		},
		notifier: notifier.NewSlackNotifier(mgr.cfg.Slack.Token, mgr.cfg.Slack.ChannelId),
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
