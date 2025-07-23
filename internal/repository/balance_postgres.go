package repository

import (
	"context"
	"errors"

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
