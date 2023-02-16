package svc

import "sync"

// service represents a Service run by the Manager.
type service interface {
	// init initializes the service
	init()
	// run executes the service
	run()
	// close terminates the service
	close()
	// name provides a name of the service
	name() string
}

// Manager represents the manager controlling services lifetime.
type Manager struct {
	wg  *sync.WaitGroup
	svc []service

	// managed services
	blkScanner *blkScanner
}

// add managed service instance to the Manager and run it.
func (mgr *Manager) add(s service) {
	// keep track of running services
	mgr.svc = append(mgr.svc, s)

	// run the service
	mgr.wg.Add(1)
	go s.run()
	log.Noticef("service %s started", s.name())
}
