package db

import (
	"context"
	"fmt"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/types"
	"github.com/jmoiron/sqlx"
)

type ProjectQueryBuilder struct {
	queryBuilder[types.Project]
}

// ProjectQuery returns a new project query builder.
func (db *Db) ProjectQuery(ctx context.Context) ProjectQueryBuilder {
	return ProjectQueryBuilder{
		queryBuilder: newQueryBuilder[types.Project](ctx, db.con, "project"),
	}
}

// WhereOwner adds a where clause to the query builder.
func (qb *ProjectQueryBuilder) WhereOwner(owner *types.Address) *ProjectQueryBuilder {
	qb.where = append(qb.where, "owner_address = :owner_address")
	qb.parameters["owner_address"] = owner
	return qb
}

// StoreProject stores the project in the database.
func (db *Db) StoreProject(ctx context.Context, project *types.Project) error {
	query := `INSERT INTO project (owner_address, project_id, receiver_address, last_withdrawal_epoch, active_from_epoch, active_to_epoch) 
		VALUES (:owner_address, :project_id, :receiver_address, :last_withdrawal_epoch, :active_from_epoch, :active_to_epoch)
		RETURNING id`

	rows, err := sqlx.NamedQueryContext(ctx, db.con, query, project)
	if err != nil {
		db.log.Errorf("failed to store project %d: %v", project.ProjectId, err)
		return err
	}
	defer func(rows *sqlx.Rows) {
		err := rows.Close()
		if err != nil {
			db.log.Errorf("failed to close rows: %v", err)
		}
	}(rows)

	if !rows.Next() {
		return fmt.Errorf("failed to store project %d: no rows returned", project.ProjectId)
	}

	// get project db id and set it to the project
	var id int64
	if err := rows.Scan(&id); err != nil {
		db.log.Errorf("failed to scan project id: %v", err)
		return err
	}
	project.Id = id

	return nil
}
