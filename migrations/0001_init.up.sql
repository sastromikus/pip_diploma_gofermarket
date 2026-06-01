CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    login TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    number TEXT NOT NULL UNIQUE,

    status TEXT NOT NULL DEFAULT 'NEW',
    accrual NUMERIC(12, 2),

    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT orders_status_check CHECK (
        status IN ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED')
    )
);

CREATE TABLE IF NOT EXISTS withdrawals (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    order_number TEXT NOT NULL,
    sum NUMERIC(12, 2) NOT NULL,

    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT withdrawals_sum_positive CHECK (sum > 0)
);

CREATE INDEX IF NOT EXISTS idx_orders_user_id_uploaded_at
ON orders(user_id, uploaded_at DESC);

CREATE INDEX IF NOT EXISTS idx_orders_status
ON orders(status);

CREATE INDEX IF NOT EXISTS idx_withdrawals_user_id_processed_at
ON withdrawals(user_id, processed_at DESC);