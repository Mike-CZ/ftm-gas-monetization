package repository

import (
	"context"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository/db"
)

// ProjectQuery returns a new project query builder.
func (repo *Repository) ProjectQuery() db.ProjectQueryBuilder {
	return repo.db.ProjectQuery(context.Background())
}
