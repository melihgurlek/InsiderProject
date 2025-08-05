-- Transaction Limit Rules Table
CREATE TABLE IF NOT EXISTS transaction_limit_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id INTEGER NOT NULL,
    rule_type TEXT NOT NULL, -- e.g., 'max_per_transaction', 'daily_total', etc.
    limit_amount NUMERIC NOT NULL,
    currency TEXT,           -- nullable if not multicurrency
    "window" INTERVAL,      -- nullable for per-tx rules
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_transaction_limit_rules_user_id ON transaction_limit_rules(user_id);
CREATE INDEX IF NOT EXISTS idx_transaction_limit_rules_active ON transaction_limit_rules(active);

-- User Transactions Table
CREATE TABLE IF NOT EXISTS user_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id INTEGER NOT NULL,
    amount NUMERIC NOT NULL,
    currency TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_transactions_user_id_created_at ON user_transactions(user_id, created_at);