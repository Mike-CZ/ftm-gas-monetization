CREATE TABLE IF NOT EXISTS transaction(
    id serial PRIMARY KEY,
    project_id INT NOT NULL,
    hash VARCHAR(64) NOT NULL,
    block_hash VARCHAR(64),
    block_number BIGINT,
    timestamp TIMESTAMP NOT NULL,
    from_address VARCHAR(40),
    to_address VARCHAR(40),
    gas_used BIGINT NOT NULL,
    gas_price TEXT NOT NULL
);
