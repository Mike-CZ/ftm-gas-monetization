package db

import (
	"context"
	"fmt"
	"ftm-gas-monetization/internal/config"
	"ftm-gas-monetization/internal/logger"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"time"
)

const (
	testDbName = "test_db"
	testDbUser = "test_user"
	testDbPass = "test_password"
	testDbPort = "5432"
	testDbHost = "localhost"
)

type TestDatabase struct {
	*Db
	container testcontainers.Container
}

// SetupTestDatabase creates a test container for postgres database and returns a database instance
func SetupTestDatabase(logger *logger.AppLogger) *TestDatabase {
	// setup db container
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	container, port, err := createContainer(ctx, logger)
	if err != nil {
		logger.Fatal("failed to setup test database", err)
	}
	db := New(
		&config.DB{
			Name:     testDbName,
			User:     testDbUser,
			Password: testDbPass,
			Host:     testDbHost,
			Port:     port,
		},
		logger)
	if db == nil {
		_ = container.Terminate(context.Background())
		logger.Fatal("failed to create database instance")
	}
	return &TestDatabase{
		Db:        db,
		container: container,
	}
}

// Migrate runs the database migrations
func (tdb *TestDatabase) Migrate() error {
	return tdb.migrateTables()
}

// Drop drops all tables
func (tdb *TestDatabase) Drop() error {
	return tdb.dropTables()
}

// TearDown removes the test container and closes the database connection
func (tdb *TestDatabase) TearDown() {
	if err := tdb.Db.db.Close(); err != nil {
		tdb.log.Fatal("failed to close database connection: ", err)
	}
	// remove test container
	_ = tdb.container.Terminate(context.Background())
}

// createContainer creates a test container for postgres database
func createContainer(ctx context.Context, logger *logger.AppLogger) (testcontainers.Container, string, error) {
	var env = map[string]string{
		"POSTGRES_PASSWORD": testDbPass,
		"POSTGRES_USER":     testDbUser,
		"POSTGRES_DB":       testDbName,
	}
	var port = fmt.Sprintf("%s/tcp", testDbPort)
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:15.2-alpine",
			ExposedPorts: []string{port},
			Env:          env,
			WaitingFor:   wait.ForLog("database system is ready to accept connections"),
		},
		Started: true,
	}
	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return container, "", fmt.Errorf("failed to start container: %v", err)
	}
	p, err := container.MappedPort(ctx, testDbPort)
	if err != nil {
		return container, "", fmt.Errorf("failed to get container external port: %v", err)
	}
	logger.Infof("postgres container ready and running at port: ", p.Port())
	// wait for the database to be ready
	time.Sleep(time.Second)
	return container, p.Port(), nil
}
