package db

import (
	"context"
	"github.com/stretchr/testify/assert"
	"math/big"
)

func (s *DbTestSuite) TestLastProcessedBlock() {
	// test setting last processed block
	err := s.db.UpdateLastProcessedBlock(context.Background(), 500)
	assert.Nil(s.T(), err)
	// test getting last processed block
	id, err := s.db.LastProcessedBlock(context.Background())
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), 500, id)
}

func (s *DbTestSuite) TestLastProcessedEpoch() {
	// test setting last processed epoch
	err := s.db.UpdateCurrentEpoch(context.Background(), 11)
	assert.Nil(s.T(), err)
	// test getting last processed epoch
	id, err := s.db.CurrentEpoch(context.Background())
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), 11, id)
}

func (s *DbTestSuite) TestTotalAmountCollected() {
	// test increasing total amount collected
	err := s.db.IncreaseTotalAmountCollected(context.Background(), new(big.Int).SetUint64(100))
	assert.Nil(s.T(), err)
	// test getting total amount collected
	amount, err := s.db.TotalAmountCollected(context.Background())
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), 100, amount.Uint64())
	// increase amount again
	err = s.db.IncreaseTotalAmountCollected(context.Background(), new(big.Int).SetUint64(50))
	assert.Nil(s.T(), err)
	// test getting total amount collected
	amount, err = s.db.TotalAmountCollected(context.Background())
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), 150, amount.Uint64())
	// test setting total amount collected
	err = s.db.SetTotalAmountCollected(context.Background(), new(big.Int).SetUint64(200))
	assert.Nil(s.T(), err)
	// test getting total amount collected
	amount, err = s.db.TotalAmountCollected(context.Background())
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), 200, amount.Uint64())
}

func (s *DbTestSuite) TestTotalAmountClaimed() {
	// test increasing total amount claimed
	err := s.db.IncreaseTotalAmountClaimed(context.Background(), new(big.Int).SetUint64(10))
	assert.Nil(s.T(), err)
	// test getting total amount claimed
	amount, err := s.db.TotalAmountClaimed(context.Background())
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), 10, amount.Uint64())
	// increase amount again
	err = s.db.IncreaseTotalAmountClaimed(context.Background(), new(big.Int).SetUint64(5))
	assert.Nil(s.T(), err)
	// test getting total amount claimed
	amount, err = s.db.TotalAmountClaimed(context.Background())
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), 15, amount.Uint64())
	// test setting total amount claimed
	err = s.db.SetTotalAmountClaimed(context.Background(), new(big.Int).SetUint64(20))
	assert.Nil(s.T(), err)
	// test getting total amount claimed
	amount, err = s.db.TotalAmountClaimed(context.Background())
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), 20, amount.Uint64())
}

func (s *DbTestSuite) TestTotalTransactionsCount() {
	// test increasing total transactions count
	err := s.db.IncreaseTotalTransactionsCount(context.Background(), 200)
	assert.Nil(s.T(), err)
	// test getting total transactions count
	count, err := s.db.TotalTransactionsCount(context.Background())
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), 200, count)
	// increase count again
	err = s.db.IncreaseTotalTransactionsCount(context.Background(), 100)
	assert.Nil(s.T(), err)
	// test getting total transactions count
	count, err = s.db.TotalTransactionsCount(context.Background())
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), 300, count)
	// test setting total transactions count
	err = s.db.SetTotalTransactionsCount(context.Background(), 400)
	assert.Nil(s.T(), err)
	// test getting total transactions count
	count, err = s.db.TotalTransactionsCount(context.Background())
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), 400, count)
}
