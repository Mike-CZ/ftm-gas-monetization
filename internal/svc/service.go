package svc

import (
	"github.com/Mike-CZ/ftm-gas-monetization/internal/logger"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository"
)

// serviceHandler represents a service run by the Manager.
type serviceHandler interface {
	// init initializes the service
	init()
	// run executes the service
	run()
	// close terminates the service
	close()
	// name provides a name of the service
	name() string
}

// service implements general base for services implementing svc interface.
type service struct {
	repo    *repository.Repository
	log     *logger.AppLogger
	mgr     *Manager
	sigStop chan struct{}
}

// init prepares the service stop signal channel.
func (s *service) init() {
	s.sigStop = make(chan struct{})
}

// close terminates the service by sending the stop signal down the channel.
func (s *service) close() {
	if s.sigStop != nil {
		close(s.sigStop)
	}
}
