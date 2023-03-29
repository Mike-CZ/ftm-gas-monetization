CREATE TABLE IF NOT EXISTS withdrawal_request(
    id serial PRIMARY KEY,
    project_id INT NOT NULL,
    request_epoch BIGINT NOT NULL,
    withdraw_epoch BIGINT,
    amount TEXT
);
