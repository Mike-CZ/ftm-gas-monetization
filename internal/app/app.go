package app

import (
	"ftm-gas-monetization/internal/config"
	"ftm-gas-monetization/internal/logger"
	"ftm-gas-monetization/internal/repository"
	"ftm-gas-monetization/internal/svc"
	"github.com/urfave/cli/v2"
	"sync"
)

// instance is the singleton of the App.
var instance App

// onceRepository is used to ensure that the repository is initialized only once.
var onceRepository sync.Once

// App defines the gas monetization app core, which holds and provides
// access to all the app's components.
type App struct {
	cfg        *config.Config
	log        *logger.AppLogger
	repository *repository.Repository
	manager    *svc.Manager
}

// Bootstrap bootstraps the app core.
func Bootstrap(ctx *cli.Context, cfg *config.Config) {
	instance = App{
		cfg: cfg,
		log: logger.New(ctx.App.Writer, ctx.App.HelpName, cfg.LoggingLevel),
	}
}

// Start initializes and starts services.
func Start() {
	// start the manager
	if instance.manager == nil {
		instance.manager = svc.New(instance.cfg, Repository(), instance.log)
		instance.manager.Run()
	}
}

// Repository provides access to the repository.
func Repository() *repository.Repository {
	onceRepository.Do(func() {
		instance.repository = repository.New(instance.cfg, instance.log)
	})
	return instance.repository
}

// Close terminates running services.
func Close() {
	if instance.manager != nil {
		instance.manager.Close()
		instance.manager = nil
	}
}
