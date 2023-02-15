package repository

import (
	"github.com/Mike-CZ/ftm-gas-monetization/internal/config"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/logger"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository/rpc"
	"github.com/op/go-logging"
	"github.com/urfave/cli/v2"
)

type Repository struct {
	rpc *rpc.Rpc
	log *logging.Logger
}

func New(ctx *cli.Context, cfg *config.Config) *Repository {
	return &Repository{
		rpc: rpc.New(cfg.OperaRpcUrl),
		log: logger.New(ctx.App.Writer, cfg.LoggingLevel, "repository"),
	}
}
