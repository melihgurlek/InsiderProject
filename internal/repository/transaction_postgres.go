package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/melihgurlek/backend-path/internal/domain"
)

// TransactionPostgresRepository implements domain.TransactionRepository using PostgreSQL.
type TransactionPostgresRepository struct {
	conn *pgx.Conn
}

// NewTransactionPostgresRepository creates a new TransactionPostgresRepository.
func NewTransactionPostgresRepository(conn *pgx.Conn) *TransactionPostgresRepository {
	return &TransactionPostgresRepository{conn: conn}
}

// Create inserts a new transaction into the database.
func (r *TransactionPostgresRepository) Create(tx *domain.Transaction) error {
	query := `INSERT INTO transactions (from_user_id, to_user_id, amount, type, status, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW()) RETURNING id, created_at`
	return r.conn.QueryRow(context.Background(), query,
		tx.FromUserID, tx.ToUserID, tx.Amount, tx.Type, tx.Status,
	).Scan(&tx.ID, &tx.CreatedAt)
}

// GetByID fetches a transaction by ID.
func (r *TransactionPostgresRepository) GetByID(id int) (*domain.Transaction, error) {
	tx := &domain.Transaction{}
	query := `SELECT id, from_user_id, to_user_id, amount, type, status, created_at FROM transactions WHERE id = $1`
	err := r.conn.QueryRow(context.Background(), query, id).Scan(
		&tx.ID, &tx.FromUserID, &tx.ToUserID, &tx.Amount, &tx.Type, &tx.Status, &tx.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // not found
		}
		return nil, err
	}
	return tx, nil
}

// ListByUser fetches all transactions for a user (as sender or receiver).
func (r *TransactionPostgresRepository) ListByUser(userID int) ([]*domain.Transaction, error) {
	query := `SELECT id, from_user_id, to_user_id, amount, type, status, created_at FROM transactions WHERE from_user_id = $1 OR to_user_id = $1 ORDER BY created_at DESC`
	rows, err := r.conn.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []*domain.Transaction
	for rows.Next() {
		tx := &domain.Transaction{}
		err := rows.Scan(&tx.ID, &tx.FromUserID, &tx.ToUserID, &tx.Amount, &tx.Type, &tx.Status, &tx.CreatedAt)
		if err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	return txs, nil
}

// ListAll fetches all transactions.
func (r *TransactionPostgresRepository) ListAll() ([]*domain.Transaction, error) {
	query := `SELECT id, from_user_id, to_user_id, amount, type, status, created_at FROM transactions ORDER BY created_at DESC`
	rows, err := r.conn.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []*domain.Transaction
	for rows.Next() {
		tx := &domain.Transaction{}
		err := rows.Scan(&tx.ID, &tx.FromUserID, &tx.ToUserID, &tx.Amount, &tx.Type, &tx.Status, &tx.CreatedAt)
		if err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	return txs, nil
}

// ListByUserAndTimeRange fetches all transactions for a user (as sender or receiver) within a time range.
func (r *TransactionPostgresRepository) ListByUserAndTimeRange(userID int, from, to time.Time) ([]*domain.Transaction, error) {
	query := `SELECT id, from_user_id, to_user_id, amount, type, status, created_at FROM transactions WHERE (from_user_id = $1 OR to_user_id = $1) AND created_at >= $2 AND created_at <= $3 ORDER BY created_at ASC`
	rows, err := r.conn.Query(context.Background(), query, userID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []*domain.Transaction
	for rows.Next() {
		tx := &domain.Transaction{}
		err := rows.Scan(&tx.ID, &tx.FromUserID, &tx.ToUserID, &tx.Amount, &tx.Type, &tx.Status, &tx.CreatedAt)
		if err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	return txs, nil
}
