package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/melihgurlek/backend-path/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type transactionLimitPostgresRepository struct {
	db *pgxpool.Pool
}

func NewTransactionLimitPostgresRepository(db *pgxpool.Pool) domain.TransactionLimitRepository {
	return &transactionLimitPostgresRepository{db: db}
}

func (r *transactionLimitPostgresRepository) CheckAndRecordTransaction(ctx context.Context, userID int, amount float64, currency string, timestamp time.Time) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
	}()

	// 1. Fetch active rules for user (snapshot)
	rules, err := r.getActiveRulesForUserTx(ctx, tx, userID)
	if err != nil {
		return fmt.Errorf("fetch rules: %w", err)
	}

	for _, rule := range rules {
		switch rule.RuleType {
		case "max_per_transaction":
			if amount > rule.LimitAmount {
				return errors.New("max per transaction limit exceeded")
			}
		case "daily_total":
			// Sum of today's transactions + this one <= limit
			var sum float64
			err = tx.QueryRow(ctx, `SELECT COALESCE(SUM(amount),0) FROM user_transactions WHERE user_id = $1 AND currency = $2 AND created_at >= date_trunc('day', $3)`, userID, currency, timestamp).Scan(&sum)
			if err != nil {
				return fmt.Errorf("query daily total: %w", err)
			}
			if sum+amount > rule.LimitAmount {
				return errors.New("daily total limit exceeded")
			}
		case "tx_count":
			// Count of transactions in window + this one <= limit
			windowStart := timestamp.Add(-rule.Window)
			var count int
			err = tx.QueryRow(ctx, `SELECT COUNT(*) FROM user_transactions WHERE user_id = $1 AND currency = $2 AND created_at >= $3`, userID, currency, windowStart).Scan(&count)
			if err != nil {
				return fmt.Errorf("query tx count: %w", err)
			}
			if float64(count+1) > rule.LimitAmount {
				return errors.New("transaction count limit exceeded")
			}
		case "min_interval":
			// New transaction must be at least window after last one
			var lastTime time.Time
			err = tx.QueryRow(ctx, `SELECT COALESCE(MAX(created_at), 'epoch') FROM user_transactions WHERE user_id = $1 AND currency = $2`, userID, currency).Scan(&lastTime)
			if err != nil {
				return fmt.Errorf("query last tx time: %w", err)
			}
			if !lastTime.IsZero() && timestamp.Sub(lastTime) < rule.Window {
				return errors.New("minimum interval between transactions not met")
			}
		}
	}

	// 3. If all pass, record transaction
	_, err = tx.Exec(ctx, `INSERT INTO user_transactions (user_id, amount, currency, created_at) VALUES ($1, $2, $3, $4)`, userID, amount, currency, timestamp)
	if err != nil {
		return fmt.Errorf("insert transaction: %w", err)
	}

	return nil
}

// getActiveRulesForUserTx fetches active rules for a user within a transaction
func (r *transactionLimitPostgresRepository) getActiveRulesForUserTx(ctx context.Context, tx pgx.Tx, userID int) ([]domain.TransactionLimitRule, error) {
	rows, err := tx.Query(ctx, `SELECT id, user_id, rule_type, limit_amount, currency, "window", active, created_at, updated_at FROM transaction_limit_rules WHERE user_id = $1 AND active = TRUE`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []domain.TransactionLimitRule
	for rows.Next() {
		var rule domain.TransactionLimitRule
		var window *time.Duration
		if err := rows.Scan(&rule.ID, &rule.UserID, &rule.RuleType, &rule.LimitAmount, &rule.Currency, &window, &rule.Active, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, err
		}
		if window != nil {
			rule.Window = *window
		}
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

func (r *transactionLimitPostgresRepository) AddRule(ctx context.Context, rule domain.TransactionLimitRule) (domain.TransactionLimitRule, error) {
	_, err := r.db.Exec(ctx, `
		INSERT INTO transaction_limit_rules (
			id, user_id, rule_type, limit_amount, currency, "window", active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`,
		rule.ID, rule.UserID, rule.RuleType, rule.LimitAmount, rule.Currency, rule.Window, rule.Active, rule.CreatedAt, rule.UpdatedAt,
	)
	if err != nil {
		return domain.TransactionLimitRule{}, fmt.Errorf("add rule: %w", err)
	}
	return rule, nil
}

func (r *transactionLimitPostgresRepository) RemoveRule(ctx context.Context, userID int, ruleID string) error {
	query := `DELETE FROM transaction_limit_rules WHERE id = $1 AND user_id = $2`

	result, err := r.db.Exec(ctx, query, ruleID, userID)
	if err != nil {
		return fmt.Errorf("remove rule: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("rule not found or permission denied")
	}

	return nil
}

func (r *transactionLimitPostgresRepository) GetRulesForUser(ctx context.Context, userID int) ([]domain.TransactionLimitRule, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, rule_type, limit_amount, currency, "window", active, created_at, updated_at
		FROM transaction_limit_rules
		WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("get rules: %w", err)
	}
	defer rows.Close()

	var rules []domain.TransactionLimitRule
	for rows.Next() {
		var rule domain.TransactionLimitRule
		var window *time.Duration
		if err := rows.Scan(&rule.ID, &rule.UserID, &rule.RuleType, &rule.LimitAmount, &rule.Currency, &window, &rule.Active, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, err
		}
		if window != nil {
			rule.Window = *window
		}
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

func (r *transactionLimitPostgresRepository) RecordTransaction(ctx context.Context, userID int, amount float64, currency string, timestamp time.Time) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO user_transactions (user_id, amount, currency, created_at)
		VALUES ($1, $2, $3, $4)
	`, userID, amount, currency, timestamp)
	if err != nil {
		return fmt.Errorf("record transaction: %w", err)
	}
	return nil
}

func (r *transactionLimitPostgresRepository) GetTransactionSum(ctx context.Context, userID int, window time.Duration, currency string) (float64, error) {
	windowStart := time.Now().Add(-window)
	var sum float64
	err := r.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount),0) FROM user_transactions
		WHERE user_id = $1 AND currency = $2 AND created_at >= $3
	`, userID, currency, windowStart).Scan(&sum)
	if err != nil {
		return 0, fmt.Errorf("get transaction sum: %w", err)
	}
	return sum, nil
}

func (r *transactionLimitPostgresRepository) GetTransactionCount(ctx context.Context, userID int, window time.Duration) (int, error) {
	windowStart := time.Now().Add(-window)
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM user_transactions
		WHERE user_id = $1 AND created_at >= $2
	`, userID, windowStart).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("get transaction count: %w", err)
	}
	return count, nil
}

func (r *transactionLimitPostgresRepository) GetLastTransactionTime(ctx context.Context, userID int) (time.Time, error) {
	var lastTime time.Time
	err := r.db.QueryRow(ctx, `
		SELECT COALESCE(MAX(created_at), 'epoch') FROM user_transactions
		WHERE user_id = $1
	`, userID).Scan(&lastTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("get last transaction time: %w", err)
	}
	return lastTime, nil
}
