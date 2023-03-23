package db

import (
	"context"
	"fmt"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/config"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/logger"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"path/filepath"
	"runtime"
)

// Db defines the database repository.
type Db struct {
	log *logger.AppLogger
	// db is the database instance. MUST NOT be used to run queries, otherwise the transactions are not going to work.
	db *sqlx.DB
	// con is the database connection that MUST be used to run queries. (can be db or tx)
	con sqlx.ExtContext
	// migrator is the database migrator.
	migrator *migrate.Migrate
}

// New creates a new database repository.
func New(config *config.DB, log *logger.AppLogger) *Db {
	dbLogger := log.ModuleLogger("db")

	// Build connection string.
	cs := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.User, config.Password, config.Host, config.Port, config.Name)

	migrator, err := initializeMigrator(cs)
	if err != nil {
		dbLogger.Criticalf("failed to create database migrator: %s", err)
		return nil
	}

	// Connect to the database.
	con, err := sqlx.Connect("postgres", cs)
	if err != nil {
		dbLogger.Criticalf("failed to connect to the database: %s", err)
		return nil
	}

	db := Db{
		db:  con,
		log: dbLogger,
		// our database instance is also the connection
		con:      con,
		migrator: migrator,
	}

	// Run the database migrations.
	if err := db.migrateTables(); err != nil {
		dbLogger.Criticalf("failed to run the database migrations: %s", err)
		_ = con.Close()
		return nil
	}

	return &db
}

// DatabaseTransaction runs the given function in a database transaction.
func (db *Db) DatabaseTransaction(ctx context.Context, fn func(context.Context, *Db) error) error {
	// Start a database transaction.
	tx, err := db.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start a database transaction: %s", err)
	}

	// Create a new database repository with the transaction as the connection.
	dbTx := Db{
		db:  db.db,
		log: db.log,
		// transaction is our connection scope
		con: tx,
	}

	// Run the given function.
	err = fn(ctx, &dbTx)
	if err != nil {
		// Rollback the transaction.
		_ = tx.Rollback()
		return fmt.Errorf("failed to run the database transaction: %s", err)
	}

	// Commit the transaction.
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit the database transaction: %s", err)
	}

	return nil
}

// migrateTables runs the database migrations.
func (db *Db) migrateTables() error {
	if err := db.migrator.Up(); err != nil {
		if err == migrate.ErrNoChange {
			return nil
		}
		return fmt.Errorf("failed to run the database migrations: %s", err)
	}
	return nil
}

// dropTables drops all the database tables.
func (db *Db) dropTables() error {
	if err := db.migrator.Down(); err != nil {
		return fmt.Errorf("failed to drop database tables: %s", err)
	}
	return nil
}

// initializeMigrator returns the database migrator.
func initializeMigrator(cs string) (*migrate.Migrate, error) {
	_, path, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("failed to get the current database migrations file path")
	}
	pathToMigrationFiles := filepath.Dir(path) + "/migrations"
	m, err := migrate.New(fmt.Sprintf("file:%s", pathToMigrationFiles), cs)
	if err != nil {
		return nil, fmt.Errorf("failed to create the database migrations: %s", err)
	}
	return m, nil
}
