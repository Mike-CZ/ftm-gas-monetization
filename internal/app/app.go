package app

import (
	"github.com/Mike-CZ/ftm-gas-monetization/internal/config"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/logger"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository"
	"github.com/urfave/cli/v2"
	"sync"
)

// instance is the singleton of the appCore.
var instance App

// App defines the gas monetization app core, which holds and provides
// access to all the app's components.
type App struct {
	ctx        *cli.Context
	cfg        *config.Config
	log        *logger.AppLogger
	repository *repository.Repository
}

// Bootstrap bootstraps the app core.
func Bootstrap(ctx *cli.Context, cfg *config.Config) {
	instance = App{
		ctx: ctx,
		cfg: cfg,
		log: logger.New(ctx, cfg),
	}
}

// Repository provides access to the repository.
var onceRepository sync.Once

func Repository() *repository.Repository {
	onceRepository.Do(func() {
		instance.repository = repository.New(instance.cfg, instance.log)
	})
	return instance.repository
}
