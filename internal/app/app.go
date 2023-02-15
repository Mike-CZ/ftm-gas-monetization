package app

import (
	"github.com/Mike-CZ/ftm-gas-monetization/internal/config"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository"
	"github.com/urfave/cli/v2"
)

// instance is the singleton of the appCore.
var instance App

// App defines the gas monetization app core, which holds and provides
// access to all the app's components.
type App struct {
	ctx        *cli.Context
	cfg        *config.Config
	repository *repository.Repository
}

// Bootstrap initializes the app.
func Bootstrap(ctx *cli.Context, cfg *config.Config) {
	instance = App{
		ctx: ctx,
		cfg: cfg,
	}
}

func Repository() *repository.Repository {
	if instance.repository == nil {
		instance.repository = repository.New(instance.ctx, instance.cfg)
	}
	return instance.repository
}
