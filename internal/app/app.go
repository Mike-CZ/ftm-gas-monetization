package app

import (
	"github.com/Mike-CZ/ftm-gas-monetization/internal/config"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/logger"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/svc"
	"github.com/urfave/cli/v2"
	"sync"
)

// instance is the singleton of the App.
var instance App

// onceRepository is used to ensure the repository is initialized only once.
var onceRepository sync.Once

// App defines the gas monetization app core, which holds and provides
// access to all the app's components.
type App struct {
	ctx        *cli.Context
	cfg        *config.Config
	log        *logger.AppLogger
	repository *repository.Repository
	manager    *svc.Manager
}

// Bootstrap bootstraps the app core.
func Bootstrap(ctx *cli.Context, cfg *config.Config) {
	instance = App{
		ctx: ctx,
		cfg: cfg,
		log: logger.New(ctx.App.Writer, ctx.App.HelpName, cfg.LoggingLevel),
	}
}

// Initialize initializes the app core and services.
func Initialize() {
	//instance.manager = svc.New(instance.log)
}

// Repository provides access to the repository.
func Repository() *repository.Repository {
	onceRepository.Do(func() {
		instance.repository = repository.New(&instance.ctx.Context, instance.cfg, instance.log)
	})
	return instance.repository
}
