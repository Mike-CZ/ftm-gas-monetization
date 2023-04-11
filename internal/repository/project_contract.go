package repository

import (
	"context"
	"ftm-gas-monetization/internal/repository/db"
)

// ProjectContractQuery returns a new project contract query builder.
func (repo *Repository) ProjectContractQuery() db.ProjectContractQueryBuilder {
	return repo.db.ProjectContractQuery(context.Background())
}
