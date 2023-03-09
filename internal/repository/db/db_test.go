package db_test

import (
	"context"
	"fmt"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/logger"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository/db"
	"github.com/op/go-logging"
	"github.com/stretchr/testify/assert"
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

func TestDatabaseTransaction(t *testing.T) {
	// test update block and epoch
	err := testDB.DatabaseTransaction(context.Background(), func(ctx context.Context, db *db.Db) error {
		err := db.UpdateLastProcessedBlock(ctx, 1)
		assert.Nil(t, err)
		err = db.UpdateLastProcessedEpoch(ctx, 2)
		assert.Nil(t, err)
		return nil
	})
	assert.Nil(t, err)

	// assert values changed
	block, err := testDB.LastProcessedBlock(context.Background())
	assert.Nil(t, err)
	assert.EqualValues(t, 1, block)
	epoch, err := testDB.LastProcessedEpoch(context.Background())
	assert.Nil(t, err)
	assert.EqualValues(t, 2, epoch)
}

func TestDatabaseTransactionRollback(t *testing.T) {
	// set values
	err := testDB.UpdateLastProcessedBlock(context.Background(), 5)
	assert.Nil(t, err)
	err = testDB.UpdateLastProcessedEpoch(context.Background(), 6)
	assert.Nil(t, err)

	// test update block and epoch and then rollback changes by returning error
	err = testDB.DatabaseTransaction(context.Background(), func(ctx context.Context, db *db.Db) error {
		err := db.UpdateLastProcessedBlock(ctx, 1)
		assert.Nil(t, err)
		err = db.UpdateLastProcessedEpoch(ctx, 2)
		assert.Nil(t, err)
		return fmt.Errorf("test error")
	})
	assert.NotNil(t, err)

	// assert values has not changed
	block, err := testDB.LastProcessedBlock(context.Background())
	assert.Nil(t, err)
	assert.EqualValues(t, 5, block)
	epoch, err := testDB.LastProcessedEpoch(context.Background())
	assert.Nil(t, err)
	assert.EqualValues(t, 6, epoch)
}
