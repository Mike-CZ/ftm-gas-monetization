CREATE TABLE IF NOT EXISTS project (
    id serial PRIMARY KEY,
    receiver_address VARCHAR(40) NOT NULL,
    name text
);