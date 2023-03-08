package db_test

import (
	"github.com/Mike-CZ/ftm-gas-monetization/internal/logger"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository/db"
	"github.com/op/go-logging"
	"log"
	"os"
	"testing"
)

var testDB *db.TestDatabase

func TestMain(m *testing.M) {
	testDB = db.SetupTestDatabase(logger.New(log.Writer(), "test", logging.ERROR))
	defer testDB.TearDown()
	os.Exit(m.Run())
}
