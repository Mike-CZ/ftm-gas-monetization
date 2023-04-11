package repository

import (
	"context"
	"ftm-gas-monetization/internal/repository/db"
	"ftm-gas-monetization/internal/types"
	"math/big"
)

// StoreProject stores the project in the database.
func (repo *Repository) StoreProject(project *types.Project) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbQueryTimeoutDuration)
	defer cancel()
	return repo.db.StoreProject(ctx, project)
}

// UpdateProject updates the project in the database.
func (repo *Repository) UpdateProject(project *types.Project) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbQueryTimeoutDuration)
	defer cancel()
	return repo.db.UpdateProject(ctx, project)
}

// ProjectQuery returns a new project query builder.
func (repo *Repository) ProjectQuery() db.ProjectQueryBuilder {
	return repo.db.ProjectQuery(context.Background())
}

// TotalAmountCollected returns the total amount collected for all projects.
func (repo *Repository) TotalAmountCollected() (*big.Int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbQueryTimeoutDuration)
	defer cancel()
	return repo.db.TotalAmountCollected(ctx)
}

// SetTotalAmountCollected sets the total amount collected for all projects.
func (repo *Repository) SetTotalAmountCollected(amount *big.Int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbQueryTimeoutDuration)
	defer cancel()
	return repo.db.SetTotalAmountCollected(ctx, amount)
}

// IncreaseTotalAmountCollected increases the total amount collected for all projects.
func (repo *Repository) IncreaseTotalAmountCollected(amount *big.Int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbQueryTimeoutDuration)
	defer cancel()
	return repo.db.IncreaseTotalAmountCollected(ctx, amount)
}

// TotalAmountClaimed returns the total amount claimed for all projects.
func (repo *Repository) TotalAmountClaimed() (*big.Int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbQueryTimeoutDuration)
	defer cancel()
	return repo.db.TotalAmountClaimed(ctx)
}

// SetTotalAmountClaimed sets the total amount claimed for all projects.
func (repo *Repository) SetTotalAmountClaimed(amount *big.Int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbQueryTimeoutDuration)
	defer cancel()
	return repo.db.SetTotalAmountClaimed(ctx, amount)
}

// IncreaseTotalAmountClaimed increases the total amount claimed for all projects.
func (repo *Repository) IncreaseTotalAmountClaimed(amount *big.Int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbQueryTimeoutDuration)
	defer cancel()
	return repo.db.IncreaseTotalAmountClaimed(ctx, amount)
}

// TotalTransactionsCount returns the total transactions count for all projects.
func (repo *Repository) TotalTransactionsCount() (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbQueryTimeoutDuration)
	defer cancel()
	return repo.db.TotalTransactionsCount(ctx)
}

// SetTotalTransactionsCount sets the total transactions count for all projects.
func (repo *Repository) SetTotalTransactionsCount(count uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbQueryTimeoutDuration)
	defer cancel()
	return repo.db.SetTotalTransactionsCount(ctx, count)
}

// IncreaseTotalTransactionsCount increases the total transactions count for all projects.
func (repo *Repository) IncreaseTotalTransactionsCount(amount uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbQueryTimeoutDuration)
	defer cancel()
	return repo.db.IncreaseTotalTransactionsCount(ctx, amount)
}
