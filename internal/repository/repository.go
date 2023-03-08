package repository

import (
	"context"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/config"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/logger"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository/db"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository/rpc"
	"time"
)

// dbQueryTimeoutDuration is the maximum time we wait for a database query to finish.
const dbQueryTimeoutDuration = 30 * time.Second

type Repository struct {
	rpc *rpc.Rpc
	db  *db.Db
	log *logger.AppLogger
}

// New creates a new repository from given config and logger.
func New(cfg *config.Config, log *logger.AppLogger) *Repository {
	repoLogger := log.ModuleLogger("repository")
	repo := Repository{
		db:  db.New(&cfg.DB, repoLogger),
		rpc: rpc.New(cfg.OperaRpcUrl, repoLogger),
		log: repoLogger,
	}

	if repo.rpc == nil || repo.db == nil {
		repoLogger.Panicf("failed to initialize repository")
		return nil
	}

	return &repo
}

// DatabaseTransaction runs the given function in a database transaction. The callback function is passed the repository
// instance with the transaction as the connection. The transaction is automatically committed if the callback function
// returns nil, otherwise it is rolled back. The callback function is passed a context that is cancelled after
// dbQueryTimeoutDuration. If the callback function does not return within this time, the transaction is rolled back.
func (repo *Repository) DatabaseTransaction(fn func(context.Context, *db.Db) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbQueryTimeoutDuration)
	defer cancel()
	return repo.db.DatabaseTransaction(ctx, fn)
}
