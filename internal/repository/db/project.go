package db

import (
	"context"
	"database/sql"
	"errors"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/jmoiron/sqlx"
)

//goland:noinspection SqlDialectInspection,SqlNoDataSourceInspection
var projectSchema = `
CREATE TABLE IF NOT EXISTS project (
    id serial PRIMARY KEY,
    receiver_address VARCHAR(40) NOT NULL,
    name text
);
CREATE TABLE IF NOT EXISTS project_contract (
    project_id INT NOT NULL,
    address VARCHAR(40) NOT NULL,
    CONSTRAINT fk_project
      FOREIGN KEY(project_id) 
      REFERENCES project(id)
      ON DELETE CASCADE
);
`

// migrateProjectTables migrates the project tables.
func (db *Db) migrateProjectTables() {
	_, err := db.db.Exec(projectSchema)
	if err != nil {
		db.log.Panicf("failed to migrate project tables: %v", err)
	}
}

// StoreContract into white list. Stored contracts are eligible for monetization.
//
//goland:noinspection SqlDialectInspection,SqlNoDataSourceInspection
func (db *Db) StoreContract(ctx context.Context, ctr *types.ProjectContract) error {
	query := `INSERT INTO project_contract (project_id, address)
		VALUES (:project_id, :address)`

	_, err := sqlx.NamedExecContext(ctx, db.con, query, ctr)
	if err != nil {
		db.log.Errorf("failed to store contract %s: %v", ctr.Address.String(), err)
		return err
	}

	db.log.Debugf("contract %s added to database", ctr.Address.String())
	return nil
}

// RemoveContract from white list - contract is no longer eligible for monetization.
//
//goland:noinspection SqlDialectInspection,SqlNoDataSourceInspection
func (db *Db) RemoveContract(ctx context.Context, ctr *types.ProjectContract) error {
	query := `DELETE FROM project_contract
		WHERE address = :address`

	_, err := sqlx.NamedExecContext(ctx, db.con, query, ctr)
	if err != nil {
		db.log.Errorf("failed to remove contract %s: %v", ctr.Address.String(), err)
		return err
	}

	db.log.Debugf("contract %s removed from database", ctr.Address.String())
	return nil
}

// FetchApprovedContracts and append them to map for contractApprover.
//
//goland:noinspection SqlDialectInspection,SqlNoDataSourceInspection
func (db *Db) FetchApprovedContracts(ctx context.Context) (map[common.Address]bool, error) {
	db.log.Debugf("fetching approved contracts")

	var contracts []types.ProjectContract

	m := make(map[common.Address]bool)

	err := sqlx.GetContext(ctx, db.con, contracts, "SELECT * FROM project_contract")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			db.log.Warningf("no contracts in white-list, returning empty map")
			return m, nil
		}

		db.log.Errorf("failed to get last block: %s", err)
		return nil, err
	}

	for _, c := range contracts {
		m[c.Address] = true
	}

	return m, nil
}
