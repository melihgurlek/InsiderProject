-- +migrate Down
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS balances;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS users; 