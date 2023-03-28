CREATE TABLE IF NOT EXISTS project(
    id serial PRIMARY KEY,
    project_id BIGINT NOT NULL,
    owner_address VARCHAR(40) NOT NULL,
    receiver_address VARCHAR(40) NOT NULL,
    -- TODO: metadata info
    last_withdrawal_epoch BIGINT,
    collected_rewards TEXT,
    claimed_rewards TEXT,
    rewards_to_claim TEXT,
    transactions_count BIGINT NOT NULL,
    active_from_epoch BIGINT NOT NULL,
    active_to_epoch BIGINT
);
