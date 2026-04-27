CREATE TABLE IF NOT EXISTS accounts (
    id      BIGSERIAL PRIMARY KEY,
    name    VARCHAR(100) NOT NULL,
    balance NUMERIC(20, 2) NOT NULL DEFAULT 10000.00,
    CONSTRAINT balance_non_negative CHECK (balance >= 0)
);