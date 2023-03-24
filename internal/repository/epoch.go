package repository

import (
	"context"
)

// LastProcessedEpoch returns the last processed epoch number.
func (repo *Repository) LastProcessedEpoch() (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbQueryTimeoutDuration)
	defer cancel()
	return repo.db.LastProcessedEpoch(ctx)
}

// UpdateLastProcessedEpoch updates the last observed epoch number.
func (repo *Repository) UpdateLastProcessedEpoch(id uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbQueryTimeoutDuration)
	defer cancel()
	return repo.db.UpdateLastProcessedEpoch(ctx, id)
}
