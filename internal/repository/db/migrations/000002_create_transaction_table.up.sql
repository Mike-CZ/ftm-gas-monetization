CREATE TABLE IF NOT EXISTS transaction(
    hash VARCHAR(64) PRIMARY KEY,
    block_hash VARCHAR(64),
    block_number BIGINT,
    timestamp TIMESTAMP NOT NULL,
    from_address VARCHAR(40),
    to_address VARCHAR(40),
    gas_limit BIGINT NOT NULL,
    gas_used BIGINT,
    gas_price TEXT NOT NULL
);
