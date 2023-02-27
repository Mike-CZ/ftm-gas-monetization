package db

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
