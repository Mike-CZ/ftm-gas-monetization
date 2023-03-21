package repository

import (
	"context"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository/db"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/types"
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
