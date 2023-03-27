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

// WhereProjectId adds a where clause to the query builder.
func (qb *ProjectQueryBuilder) WhereProjectId(projectId uint64) *ProjectQueryBuilder {
	qb.where = append(qb.where, "project_id = :project_id")
	qb.parameters["project_id"] = projectId
	return qb
}

// WhereOwner adds a where clause to the query builder.
func (qb *ProjectQueryBuilder) WhereOwner(owner *types.Address) *ProjectQueryBuilder {
	qb.where = append(qb.where, "owner_address = :owner_address")
	qb.parameters["owner_address"] = owner
	return qb
}

// WhereActiveInEpoch adds a where clause to the query builder.
func (qb *ProjectQueryBuilder) WhereActiveInEpoch(epoch uint64) *ProjectQueryBuilder {
	qb.where = append(qb.where, "active_from_epoch <= :epoch AND (active_to_epoch IS NULL OR active_to_epoch > :epoch)")
	qb.parameters["epoch"] = epoch
	return qb
}

// StoreProject stores the project in the database.
func (db *Db) StoreProject(ctx context.Context, project *types.Project) error {
	query := `INSERT INTO project (owner_address, project_id, receiver_address, last_withdrawal_epoch, 
                     collected_rewards, claimed_rewards, transactions_count, active_from_epoch, active_to_epoch) 
		VALUES (:owner_address, :project_id, :receiver_address, :last_withdrawal_epoch, :collected_rewards, 
		        :claimed_rewards, :transactions_count, :active_from_epoch, :active_to_epoch)
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

// UpdateProject updates the project in the database.
func (db *Db) UpdateProject(ctx context.Context, project *types.Project) error {
	if project.Id == 0 {
		return fmt.Errorf("failed to update project %d: project id is 0", project.ProjectId)
	}
	query := `UPDATE project SET owner_address = :owner_address, receiver_address = :receiver_address,
                   last_withdrawal_epoch = :last_withdrawal_epoch, collected_rewards = :collected_rewards,
                   claimed_rewards = :claimed_rewards, transactions_count = :transactions_count,
                   active_from_epoch = :active_from_epoch, active_to_epoch = :active_to_epoch WHERE id = :id`

	_, err := sqlx.NamedExecContext(ctx, db.con, query, project)
	if err != nil {
		db.log.Errorf("failed to update project %d: %v", project.ProjectId, err)
		return err
	}

	return nil
}
