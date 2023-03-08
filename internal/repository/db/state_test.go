package db_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLastProcessedBlock(t *testing.T) {
	// test setting last processed block
	err := testDB.UpdateLastProcessedBlock(context.Background(), 500)
	assert.Nil(t, err)
	// test getting last processed block
	id, err := testDB.LastProcessedBlock(context.Background())
	assert.Nil(t, err)
	assert.EqualValues(t, 500, id)
}

func TestLastProcessedEpoch(t *testing.T) {
	// test setting last processed epoch
	err := testDB.UpdateLastProcessedEpoch(context.Background(), 11)
	assert.Nil(t, err)
	// test getting last processed epoch
	id, err := testDB.LastProcessedEpoch(context.Background())
	assert.Nil(t, err)
	assert.EqualValues(t, 11, id)
}
