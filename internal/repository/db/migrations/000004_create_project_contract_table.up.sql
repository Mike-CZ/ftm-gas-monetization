CREATE TABLE IF NOT EXISTS project_contract(
    id serial PRIMARY KEY,
    project_id INT NOT NULL,
    address VARCHAR(40) NOT NULL,
    is_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT fk_project FOREIGN KEY(project_id) REFERENCES project(id) ON DELETE CASCADE
);
