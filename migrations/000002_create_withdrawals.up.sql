CREATE TABLE IF NOT EXISTS withdrawals (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    order_number TEXT NOT NULL,
    sum NUMERIC(12, 2) NOT NULL,

    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT withdrawals_sum_positive CHECK (sum > 0)
);

CREATE INDEX IF NOT EXISTS idx_withdrawals_user_id_processed_at
ON withdrawals(user_id, processed_at DESC);