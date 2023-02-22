package repository

import (
	"context"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/config"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/logger"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository/db"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository/rpc"
)

type Repository struct {
	ctx *context.Context
	rpc *rpc.Rpc
	db  *db.Db
	log *logger.AppLogger
}

// New creates a new repository from given config and logger.
func New(ctx *context.Context, cfg *config.Config, log *logger.AppLogger) *Repository {
	repoLogger := log.ModuleLogger("repository")
	repo := Repository{
		ctx: ctx,
		db:  db.New(ctx, cfg, repoLogger),
		rpc: rpc.New(ctx, cfg.OperaRpcUrl, repoLogger),
		log: repoLogger,
	}

	if repo.rpc == nil || repo.db == nil {
		repoLogger.Panicf("failed to initialize repository")
		return nil
	}

	return &repo
}
