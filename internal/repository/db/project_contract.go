package db

import (
	"context"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/types"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type ProjectContractQueryBuilder struct {
	queryBuilder[types.ProjectContract]
}

// ProjectContractQuery returns a new project contract query builder.
func (db *Db) ProjectContractQuery(ctx context.Context) ProjectContractQueryBuilder {
	return ProjectContractQueryBuilder{
		queryBuilder: newQueryBuilder[types.ProjectContract](ctx, db.con, "project_contract"),
	}
}

// WhereProjectId adds a where clause to the query builder.
func (qb *ProjectContractQueryBuilder) WhereProjectId(projectId int64) *ProjectContractQueryBuilder {
	qb.where = append(qb.where, "project_id = :project_id")
	qb.parameters["project_id"] = projectId
	return qb
}

// WhereProjectIdIn adds a where clause to the query builder.
func (qb *ProjectContractQueryBuilder) WhereProjectIdIn(ids []int64) *ProjectContractQueryBuilder {
	qb.where = append(qb.where, "project_id = ANY(:project_ids)")
	qb.parameters["project_ids"] = pq.Array(ids)
	return qb
}

// WhereIsApproved adds a where clause to the query builder.
func (qb *ProjectContractQueryBuilder) WhereIsApproved(isApproved bool) *ProjectContractQueryBuilder {
	qb.where = append(qb.where, "is_approved = :is_approved")
	qb.parameters["is_approved"] = isApproved
	return qb
}

// WhereAddress adds a where clause to the query builder.
func (qb *ProjectContractQueryBuilder) WhereAddress(address *types.Address) *ProjectContractQueryBuilder {
	qb.where = append(qb.where, "address = :address")
	qb.parameters["address"] = address
	return qb
}

// StoreProjectContract stores the project contract in the database.
func (db *Db) StoreProjectContract(ctx context.Context, project *types.ProjectContract) error {
	query := `INSERT INTO project_contract (project_id, address, is_approved) VALUES (:project_id, :address, :is_approved)`
	_, err := sqlx.NamedExecContext(ctx, db.con, query, project)
	if err != nil {
		db.log.Errorf("failed to store project contract %d: %v", project.Address.Hex(), err)
		return err
	}
	return nil
}
