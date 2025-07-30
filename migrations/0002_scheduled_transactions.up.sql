-- Create scheduled_transactions table
CREATE TABLE IF NOT EXISTS scheduled_transactions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    to_user_id INTEGER,
    amount DECIMAL(15,2) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('credit', 'debit', 'transfer')),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'failed', 'cancelled')),
    schedule_at TIMESTAMP WITH TIME ZONE NOT NULL,
    recurring BOOLEAN NOT NULL DEFAULT FALSE,
    recurrence VARCHAR(20) CHECK (recurrence IN ('daily', 'weekly', 'monthly', 'yearly')),
    next_run_at TIMESTAMP WITH TIME ZONE,
    max_runs INTEGER,
    runs_count INTEGER NOT NULL DEFAULT 0,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Foreign key constraints
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (to_user_id) REFERENCES users(id) ON DELETE CASCADE,
    
    -- Constraints
    CONSTRAINT valid_transfer CHECK (
        (type = 'transfer' AND to_user_id IS NOT NULL) OR 
        (type IN ('credit', 'debit') AND to_user_id IS NULL)
    ),
    CONSTRAINT valid_recurring CHECK (
        (recurring = TRUE AND recurrence IS NOT NULL) OR 
        (recurring = FALSE AND recurrence IS NULL)
    ),
    CONSTRAINT valid_max_runs CHECK (
        (recurring = TRUE AND max_runs IS NULL) OR 
        (recurring = TRUE AND max_runs > 0) OR 
        (recurring = FALSE AND max_runs IS NULL)
    )
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_scheduled_transactions_user_id ON scheduled_transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_scheduled_transactions_status ON scheduled_transactions(status);
CREATE INDEX IF NOT EXISTS idx_scheduled_transactions_schedule_at ON scheduled_transactions(schedule_at);
CREATE INDEX IF NOT EXISTS idx_scheduled_transactions_next_run_at ON scheduled_transactions(next_run_at);
CREATE INDEX IF NOT EXISTS idx_scheduled_transactions_pending_execution ON scheduled_transactions(status, schedule_at) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_scheduled_transactions_recurring_pending ON scheduled_transactions(status, next_run_at) WHERE status = 'pending' AND recurring = TRUE;

-- Create trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_scheduled_transactions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_scheduled_transactions_updated_at
    BEFORE UPDATE ON scheduled_transactions
    FOR EACH ROW
    EXECUTE FUNCTION update_scheduled_transactions_updated_at(); 