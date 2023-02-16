package repository

import (
	"github.com/Mike-CZ/ftm-gas-monetization/internal/config"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/logger"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository/rpc"
)

type Repository struct {
	rpc *rpc.Rpc
	log *logger.AppLogger
}

func New(cfg *config.Config, log *logger.AppLogger) *Repository {
	return &Repository{
		rpc: rpc.New(cfg.OperaRpcUrl),
		log: log.ModuleLogger("repository"),
	}
}
