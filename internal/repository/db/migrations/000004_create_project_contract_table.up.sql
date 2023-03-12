CREATE TABLE IF NOT EXISTS project_contract (
    project_id INT NOT NULL,
    address VARCHAR(40) NOT NULL,
    CONSTRAINT fk_project FOREIGN KEY(project_id) REFERENCES project(id) ON DELETE CASCADE
);
