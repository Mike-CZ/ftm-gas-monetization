package db

import (
	"context"
	"fmt"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/config"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/logger"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Db defines the database repository.
type Db struct {
	ctx *context.Context
	con *sqlx.DB
	log *logger.AppLogger
}

// New creates a new database repository.
func New(ctx *context.Context, cfg *config.Config, log *logger.AppLogger) *Db {
	dbLogger := log.ModuleLogger("db")

	// Build connection string.
	cs := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DbUser, cfg.DbPassword, cfg.DbHost, cfg.DbPort, cfg.DbName)

	// Connect to the database.
	con, err := sqlx.Connect("postgres", cs)
	if err != nil {
		dbLogger.Criticalf("failed to connect to the database: %s", err)
		return nil
	}

	db := Db{
		ctx: ctx,
		con: con,
		log: dbLogger,
	}

	// Run the database migrations.
	db.migrateTables()

	return &db
}

// migrateTables runs the database migrations.
func (db *Db) migrateTables() {
	db.log.Notice("running database migrations")
	db.migrateStateTables()
	db.migrateProjectTables()
	db.log.Notice("database migrations completed")
}
