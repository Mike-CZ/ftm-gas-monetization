package repository

import (
	"context"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/config"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/logger"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository/db"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository/rpc"
	"sync"
	"time"
)

// dbQueryTimeoutDuration is the maximum time we wait for a database query to finish.
const dbQueryTimeoutDuration = 30 * time.Second

type Repository struct {
	rpc *rpc.Rpc
	db  *db.Db
	log *logger.AppLogger
}

// config represents the configuration setup used by the repository
// to establish and maintain required connectivity to external services
// as needed.
var cfg *config.Config

// log represents the logger to be used by the repository.
var log *logger.AppLogger

// instance is the singleton of the Proxy, implementing Repository interface.
var instance *Repository

// oneInstance is the sync guarding Repository singleton creation.
var oneInstance sync.Once

// instanceMux controls access to the repository instance
var instanceMux sync.Mutex

// SetConfig sets the repository configuration to be used to establish
// and maintain external repository connections.
func SetConfig(c *config.Config) {
	cfg = c
}

// SetLogger sets the repository logger to be used to collect logging info.
func SetLogger(l logger.AppLogger) {
	log = l.ModuleLogger("repo")
}

// New creates a new repository from given config and logger.
func New(cfg *config.Config, log *logger.AppLogger) *Repository {
	repoLogger := log.ModuleLogger("repository")
	repo := Repository{
		db:  db.New(&cfg.DB, repoLogger),
		rpc: rpc.New(&cfg.Rpc, &cfg.GasMonetization, repoLogger),
		log: repoLogger,
	}

	if repo.rpc == nil || repo.db == nil {
		repoLogger.Panicf("failed to initialize repository")
		return nil
	}

	return &repo
}

// R provides access to the singleton instance of the Repository.
func R() *Repository {
	instanceMux.Lock()
	defer instanceMux.Unlock()

	// make sure to instantiate the Repository only once
	oneInstance.Do(func() {
		instance = newProxy()
	})
	return instance
}

// newProxy creates new instance of Proxy, implementing the Repository interface.
func newProxy() *Repository {
	// make Proxy instance
	p := Repository{
		db:  db.New(&cfg.DB, log),
		log: log,
	}

	if p.db == nil {
		log.Panicf("repository init failed")
		return nil
	}

	log.Notice("repository ready")
	return &p
}

// NewWithInstances creates a new repository from given instances.
func NewWithInstances(db *db.Db, rpc *rpc.Rpc, log *logger.AppLogger) *Repository {
	repo := Repository{
		db:  db,
		rpc: rpc,
		log: log,
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
