package db

import (
	"context"
	"fmt"
	"ftm-gas-monetization/internal/logger"
	"github.com/op/go-logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"log"
	"testing"
)

type DbTestSuite struct {
	suite.Suite
	db *TestDatabase
}

func TestDbTestSuite(t *testing.T) {
	suite.Run(t, new(DbTestSuite))
}

func (s *DbTestSuite) SetupSuite() {
	s.db = SetupTestDatabase(logger.New(log.Writer(), "test", logging.ERROR))
}

func (s *DbTestSuite) SetupTest() {
	err := s.db.Migrate()
	assert.Nil(s.T(), err)
}

func (s *DbTestSuite) TearDownTest() {
	err := s.db.Drop()
	assert.Nil(s.T(), err)
}

func (s *DbTestSuite) TearDownSuite() {
	s.db.TearDown()
}

func (s *DbTestSuite) TestDatabaseTransaction() {
	// test update block and epoch
	err := s.db.DatabaseTransaction(context.Background(), func(ctx context.Context, db *Db) error {
		err := s.db.UpdateLastProcessedBlock(ctx, 1)
		assert.Nil(s.T(), err)
		err = s.db.UpdateCurrentEpoch(ctx, 2)
		assert.Nil(s.T(), err)
		return nil
	})
	assert.Nil(s.T(), err)

	// assert values changed
	block, err := s.db.LastProcessedBlock(context.Background())
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), 1, block)
	epoch, err := s.db.CurrentEpoch(context.Background())
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), 2, epoch)
}

func (s *DbTestSuite) TestDatabaseTransactionRollback() {
	// set values
	err := s.db.UpdateLastProcessedBlock(context.Background(), 5)
	assert.Nil(s.T(), err)
	err = s.db.UpdateCurrentEpoch(context.Background(), 6)
	assert.Nil(s.T(), err)

	// test update block and epoch and then rollback changes by returning error
	err = s.db.DatabaseTransaction(context.Background(), func(ctx context.Context, db *Db) error {
		err := db.UpdateLastProcessedBlock(ctx, 1)
		assert.Nil(s.T(), err)
		err = db.UpdateCurrentEpoch(ctx, 2)
		assert.Nil(s.T(), err)
		return fmt.Errorf("test error")
	})
	assert.NotNil(s.T(), err)

	// assert values has not changed
	block, err := s.db.LastProcessedBlock(context.Background())
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), 5, block)
	epoch, err := s.db.CurrentEpoch(context.Background())
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), 6, epoch)
}
