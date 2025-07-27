package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/melihgurlek/backend-path/internal/domain"
)

type BalancePostgresRepository struct {
	conn *pgx.Conn
}

func NewBalancePostgresRepository(conn *pgx.Conn) *BalancePostgresRepository {
	return &BalancePostgresRepository{conn: conn}
}

func (r *BalancePostgresRepository) Create(balance *domain.Balance) error {
	_, err := r.conn.Exec(context.Background(), "INSERT INTO balances (user_id, amount, last_updated_at) VALUES ($1, $2, $3)", balance.UserID, balance.Amount, balance.LastUpdatedAt)
	return err
}

func (r *BalancePostgresRepository) GetByUserID(userID int) (*domain.Balance, error) {
	balance := &domain.Balance{}
	query := `SELECT user_id, amount, last_updated_at FROM balances WHERE user_id = $1`
	err := r.conn.QueryRow(context.Background(), query, userID).Scan(&balance.UserID, &balance.Amount, &balance.LastUpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return balance, nil
}

func (r *BalancePostgresRepository) Update(balance *domain.Balance) error {
	query := `INSERT INTO balances (user_id, amount, last_updated_at) 
	VALUES ($1, $2, NOW()) 
	ON CONFLICT (user_id) DO UPDATE SET amount = $2, last_updated_at = NOW()`

	_, err := r.conn.Exec(context.Background(), query, balance.UserID, balance.Amount)
	return err
}

func (r *BalancePostgresRepository) GetHistoricalBalance(userID int) ([]*domain.Balance, error) {
	// Calculate daily balances for the past 30 days (including today)
	nDays := 30
	end := time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour) // tomorrow 00:00
	start := end.AddDate(0, 0, -nDays)

	// Use the transaction repository to get all transactions in the range
	trxRepo := NewTransactionPostgresRepository(r.conn)
	transactions, err := trxRepo.ListByUserAndTimeRange(userID, time.Time{}, end) // get all up to 'end'
	if err != nil {
		return nil, err
	}

	// Prepare a map of day -> net change (fix off-by-one)
	balanceByDay := make([]float64, nDays)
	for _, tx := range transactions {
		for i := 0; i < nDays; i++ {
			dayStart := start.AddDate(0, 0, i)
			dayEnd := start.AddDate(0, 0, i+1)
			if !tx.CreatedAt.Before(dayStart) && tx.CreatedAt.Before(dayEnd) {
				if tx.ToUserID != nil && *tx.ToUserID == userID && (tx.Type == "credit" || tx.Type == "transfer") {
					balanceByDay[i] += tx.Amount
				}
				if tx.FromUserID != nil && *tx.FromUserID == userID && (tx.Type == "debit" || tx.Type == "transfer") {
					balanceByDay[i] -= tx.Amount
				}
				break
			}
		}
	}

	// Build the time series
	balances := make([]*domain.Balance, nDays)
	cumulative := 0.0
	for i := 0; i < nDays; i++ {
		cumulative += balanceByDay[i]
		balances[i] = &domain.Balance{
			UserID:        userID,
			Amount:        cumulative,
			LastUpdatedAt: start.AddDate(0, 0, i),
		}
	}
	return balances, nil
}

func (r *BalancePostgresRepository) GetBalanceAtTime(userID int, t time.Time) (*domain.Balance, error) {
	query := `
        SELECT
            COALESCE(SUM(CASE
                WHEN to_user_id = $1 AND type IN ('credit', 'transfer') THEN amount
                ELSE 0 END), 0) as total_credits,
            COALESCE(SUM(CASE
                WHEN from_user_id = $1 AND type IN ('debit', 'transfer') THEN amount
                ELSE 0 END), 0) as total_debits
        FROM
            transactions
        WHERE
            (to_user_id = $1 OR from_user_id = $1) AND created_at <= $2;
    `

	var credits float64
	var debits float64

	err := r.conn.QueryRow(context.Background(), query, userID, t).Scan(&credits, &debits)
	if err != nil {
		return nil, err
	}

	finalBalance := credits - debits

	return &domain.Balance{
		UserID:        userID,
		Amount:        finalBalance,
		LastUpdatedAt: t,
	}, nil
}
