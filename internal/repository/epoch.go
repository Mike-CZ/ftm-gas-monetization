package repository

import (
	"context"
)

// CurrentEpoch returns the current epoch number.
func (repo *Repository) CurrentEpoch() (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbQueryTimeoutDuration)
	defer cancel()
	return repo.db.CurrentEpoch(ctx)
}

// UpdateCurrentEpoch updates the last observed epoch number.
func (repo *Repository) UpdateCurrentEpoch(id uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbQueryTimeoutDuration)
	defer cancel()
	return repo.db.UpdateCurrentEpoch(ctx, id)
}
