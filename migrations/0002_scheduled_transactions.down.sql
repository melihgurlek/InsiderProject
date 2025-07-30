-- Drop trigger and function
DROP TRIGGER IF EXISTS trigger_update_scheduled_transactions_updated_at ON scheduled_transactions;
DROP FUNCTION IF EXISTS update_scheduled_transactions_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_scheduled_transactions_user_id;
DROP INDEX IF EXISTS idx_scheduled_transactions_status;
DROP INDEX IF EXISTS idx_scheduled_transactions_schedule_at;
DROP INDEX IF EXISTS idx_scheduled_transactions_next_run_at;
DROP INDEX IF EXISTS idx_scheduled_transactions_pending_execution;
DROP INDEX IF EXISTS idx_scheduled_transactions_recurring_pending;

-- Drop table
DROP TABLE IF EXISTS scheduled_transactions; 