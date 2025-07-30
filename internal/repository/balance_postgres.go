package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/melihgurlek/backend-path/internal/domain"
)

type BalancePostgresRepository struct {
	pool *pgxpool.Pool
}

func NewBalancePostgresRepository(pool *pgxpool.Pool) *BalancePostgresRepository {
	return &BalancePostgresRepository{pool: pool}
}

func (r *BalancePostgresRepository) Create(balance *domain.Balance) error {
	_, err := r.pool.Exec(context.Background(), "INSERT INTO balances (user_id, amount, last_updated_at) VALUES ($1, $2, $3)", balance.UserID, balance.Amount, balance.LastUpdatedAt)
	return err
}

func (r *BalancePostgresRepository) GetByUserID(userID int) (*domain.Balance, error) {
	balance := &domain.Balance{}
	query := `SELECT user_id, amount, last_updated_at FROM balances WHERE user_id = $1`
	err := r.pool.QueryRow(context.Background(), query, userID).Scan(&balance.UserID, &balance.Amount, &balance.LastUpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return balance, nil
}

// Update updates a user's balance with proper locking to prevent race conditions
func (r *BalancePostgresRepository) Update(balance *domain.Balance) error {
	// Use a transaction to ensure atomicity and prevent race conditions
	tx, err := r.pool.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	// Lock the row for update to prevent concurrent modifications
	query := `SELECT user_id, amount, last_updated_at FROM balances WHERE user_id = $1 FOR UPDATE`
	var currentBalance domain.Balance
	err = tx.QueryRow(context.Background(), query, balance.UserID).Scan(
		&currentBalance.UserID, &currentBalance.Amount, &currentBalance.LastUpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// User doesn't have a balance record yet, create one
			insertQuery := `INSERT INTO balances (user_id, amount, last_updated_at) VALUES ($1, $2, NOW())`
			_, err = tx.Exec(context.Background(), insertQuery, balance.UserID, balance.Amount)
		}
	} else {
		// Update existing balance
		updateQuery := `UPDATE balances SET amount = $1, last_updated_at = NOW() WHERE user_id = $2`
		_, err = tx.Exec(context.Background(), updateQuery, balance.Amount, balance.UserID)
	}

	if err != nil {
		return err
	}

	return tx.Commit(context.Background())
}

// GetHistoricalBalances calculates balance history from transaction data
func (r *BalancePostgresRepository) GetHistoricalBalance(userID int, limit int) ([]*domain.Balance, error) {
	query := `
		WITH daily_balances AS (
			SELECT 
				DATE(created_at) as balance_date,
				SUM(CASE 
					WHEN to_user_id = $1 AND type IN ('credit', 'transfer') THEN amount
					WHEN from_user_id = $1 AND type IN ('debit', 'transfer') THEN -amount
					ELSE 0 
				END) as daily_change
			FROM transactions 
			WHERE (to_user_id = $1 OR from_user_id = $1) 
				AND status = 'completed'
				AND created_at >= CURRENT_DATE - INTERVAL '30 days'
			GROUP BY DATE(created_at)
			ORDER BY balance_date DESC
		),
		cumulative_balances AS (
			SELECT 
				balance_date,
				daily_change,
				SUM(daily_change) OVER (
					ORDER BY balance_date DESC 
					ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
				) as cumulative_balance
			FROM daily_balances
		)
		SELECT 
			$1::integer as user_id,
			cumulative_balance as amount,
			balance_date as last_updated_at
		FROM cumulative_balances
		ORDER BY balance_date DESC
		LIMIT $2
	`

	rows, err := r.pool.Query(context.Background(), query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var balances []*domain.Balance
	for rows.Next() {
		balance := &domain.Balance{}
		err := rows.Scan(&balance.UserID, &balance.Amount, &balance.LastUpdatedAt)
		if err != nil {
			return nil, err
		}
		balances = append(balances, balance)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return balances, nil
}

// GetBalanceAtTime calculates the balance at a specific point in time from transaction history
func (r *BalancePostgresRepository) GetBalanceAtTime(userID int, timestamp time.Time) (*domain.Balance, error) {
	query := `
		SELECT 
			$1::integer as user_id,
			COALESCE(SUM(CASE 
				WHEN to_user_id = $1 AND type IN ('credit', 'transfer') THEN amount
				WHEN from_user_id = $1 AND type IN ('debit', 'transfer') THEN -amount
				ELSE 0 
			END), 0) as amount,
			$2::timestamp as last_updated_at
		FROM transactions 
		WHERE (to_user_id = $1 OR from_user_id = $1) 
			AND status = 'completed'
			AND created_at <= $2
	`

	balance := &domain.Balance{}
	err := r.pool.QueryRow(context.Background(), query, userID, timestamp).Scan(
		&balance.UserID, &balance.Amount, &balance.LastUpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// No transactions found, return zero balance
			return &domain.Balance{
				UserID:        userID,
				Amount:        0,
				LastUpdatedAt: timestamp,
			}, nil
		}
		return nil, err
	}

	return balance, nil
}

func (r *BalancePostgresRepository) GetCurrentBalance(userID int) (*domain.Balance, error) {
	query := `
		SELECT 
			$1::integer as user_id,
			COALESCE(SUM(CASE 
				WHEN to_user_id = $1 AND type IN ('credit', 'transfer') THEN amount
				WHEN from_user_id = $1 AND type IN ('debit', 'transfer') THEN -amount
				ELSE 0 
			END), 0) as amount,
			NOW()::timestamp as last_updated_at
		FROM transactions 
		WHERE (to_user_id = $1 OR from_user_id = $1) 
			AND status = 'completed'
	`

	balance := &domain.Balance{}
	err := r.pool.QueryRow(context.Background(), query, userID).Scan(
		&balance.UserID, &balance.Amount, &balance.LastUpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return balance, nil
}
