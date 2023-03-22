CREATE TABLE IF NOT EXISTS withdrawal_request(
    id serial PRIMARY KEY,
    project_id INT NOT NULL,
    epoch BIGINT NOT NULL
);
