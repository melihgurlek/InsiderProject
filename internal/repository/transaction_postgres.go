package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/melihgurlek/backend-path/internal/domain"
)

// TransactionPostgresRepository implements domain.TransactionRepository using PostgreSQL.
type TransactionPostgresRepository struct {
	pool *pgxpool.Pool
}

// NewTransactionPostgresRepository creates a new TransactionPostgresRepository.
func NewTransactionPostgresRepository(pool *pgxpool.Pool) *TransactionPostgresRepository {
	return &TransactionPostgresRepository{pool: pool}
}

// Create inserts a new transaction into the database.
func (r *TransactionPostgresRepository) Create(tx *domain.Transaction) error {
	query := `INSERT INTO transactions (from_user_id, to_user_id, amount, type, status, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW()) RETURNING id, created_at`
	return r.pool.QueryRow(context.Background(), query,
		tx.FromUserID, tx.ToUserID, tx.Amount, tx.Type, tx.Status,
	).Scan(&tx.ID, &tx.CreatedAt)
}

// GetByID fetches a transaction by ID.
func (r *TransactionPostgresRepository) GetByID(id int) (*domain.Transaction, error) {
	tx := &domain.Transaction{}
	query := `SELECT id, from_user_id, to_user_id, amount, type, status, created_at FROM transactions WHERE id = $1`
	err := r.pool.QueryRow(context.Background(), query, id).Scan(
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
	query := `SELECT id, from_user_id, to_user_id, amount, type, status, created_at 
		FROM transactions 
		WHERE from_user_id = $1 OR to_user_id = $1 
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*domain.Transaction
	for rows.Next() {
		tx := &domain.Transaction{}
		err := rows.Scan(
			&tx.ID, &tx.FromUserID, &tx.ToUserID, &tx.Amount, &tx.Type, &tx.Status, &tx.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

// ListByUserAndTimeRange fetches transactions for a user within a time range.
func (r *TransactionPostgresRepository) ListByUserAndTimeRange(userID int, start, end time.Time) ([]*domain.Transaction, error) {
	query := `SELECT id, from_user_id, to_user_id, amount, type, status, created_at 
		FROM transactions 
		WHERE (from_user_id = $1 OR to_user_id = $1) AND created_at >= $2 AND created_at <= $3 
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(context.Background(), query, userID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*domain.Transaction
	for rows.Next() {
		tx := &domain.Transaction{}
		err := rows.Scan(
			&tx.ID, &tx.FromUserID, &tx.ToUserID, &tx.Amount, &tx.Type, &tx.Status, &tx.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

// UpdateStatus updates the status of a transaction.
func (r *TransactionPostgresRepository) UpdateStatus(id int, status string) error {
	query := `UPDATE transactions SET status = $1 WHERE id = $2`
	result, err := r.pool.Exec(context.Background(), query, status, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("transaction not found")
	}
	return nil
}

func (r *TransactionPostgresRepository) ListAll(ctx context.Context, limit int, offset int) ([]*domain.Transaction, error) {
	query := `SELECT id, from_user_id, to_user_id, amount, type, status, created_at 
		FROM transactions 
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*domain.Transaction
	for rows.Next() {
		tx := &domain.Transaction{}
		err := rows.Scan(
			&tx.ID, &tx.FromUserID, &tx.ToUserID, &tx.Amount, &tx.Type, &tx.Status, &tx.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}
