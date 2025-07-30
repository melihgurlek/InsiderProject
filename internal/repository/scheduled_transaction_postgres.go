package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/melihgurlek/backend-path/internal/domain"
)

// ScheduledTransactionPostgresRepository implements domain.ScheduledTransactionRepository using PostgreSQL.
type ScheduledTransactionPostgresRepository struct {
	pool *pgxpool.Pool
}

// NewScheduledTransactionPostgresRepository creates a new ScheduledTransactionPostgresRepository.
func NewScheduledTransactionPostgresRepository(pool *pgxpool.Pool) *ScheduledTransactionPostgresRepository {
	return &ScheduledTransactionPostgresRepository{pool: pool}
}

// Create inserts a new scheduled transaction into the database.
func (r *ScheduledTransactionPostgresRepository) Create(st *domain.ScheduledTransaction) error {
	query := `
		INSERT INTO scheduled_transactions (
			user_id, to_user_id, amount, type, status, schedule_at, 
			recurring, recurrence, next_run_at, max_runs, runs_count, description, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW(), NOW()) 
		RETURNING id, created_at, updated_at
	`
	return r.pool.QueryRow(context.Background(), query,
		st.UserID, st.ToUserID, st.Amount, st.Type, st.Status, st.ScheduleAt,
		st.Recurring, st.Recurrence, st.NextRunAt, st.MaxRuns, st.RunsCount, st.Description,
	).Scan(&st.ID, &st.CreatedAt, &st.UpdatedAt)
}

// GetByID fetches a scheduled transaction by ID.
func (r *ScheduledTransactionPostgresRepository) GetByID(id int) (*domain.ScheduledTransaction, error) {
	st := &domain.ScheduledTransaction{}
	query := `
		SELECT id, user_id, to_user_id, amount, type, status, schedule_at, 
		       recurring, recurrence, next_run_at, max_runs, runs_count, description, created_at, updated_at
		FROM scheduled_transactions WHERE id = $1
	`
	err := r.pool.QueryRow(context.Background(), query, id).Scan(
		&st.ID, &st.UserID, &st.ToUserID, &st.Amount, &st.Type, &st.Status, &st.ScheduleAt,
		&st.Recurring, &st.Recurrence, &st.NextRunAt, &st.MaxRuns, &st.RunsCount, &st.Description,
		&st.CreatedAt, &st.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // not found
		}
		return nil, err
	}
	return st, nil
}

// ListByUser fetches all scheduled transactions for a user.
func (r *ScheduledTransactionPostgresRepository) ListByUser(userID int) ([]*domain.ScheduledTransaction, error) {
	query := `
		SELECT id, user_id, to_user_id, amount, type, status, schedule_at, 
		       recurring, recurrence, next_run_at, max_runs, runs_count, description, created_at, updated_at
		FROM scheduled_transactions 
		WHERE user_id = $1 
		ORDER BY schedule_at ASC
	`
	rows, err := r.pool.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*domain.ScheduledTransaction
	for rows.Next() {
		st := &domain.ScheduledTransaction{}
		err := rows.Scan(
			&st.ID, &st.UserID, &st.ToUserID, &st.Amount, &st.Type, &st.Status, &st.ScheduleAt,
			&st.Recurring, &st.Recurrence, &st.NextRunAt, &st.MaxRuns, &st.RunsCount, &st.Description,
			&st.CreatedAt, &st.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, st)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

// ListPending fetches all pending scheduled transactions that should be executed
func (r *ScheduledTransactionPostgresRepository) ListPending() ([]*domain.ScheduledTransaction, error) {
	query := `
		SELECT id, user_id, to_user_id, amount, type, status, schedule_at, 
		       recurring, recurrence, next_run_at, max_runs, runs_count, description, created_at, updated_at
		FROM scheduled_transactions 
		WHERE status = 'pending' AND (
			(recurring = FALSE AND schedule_at <= NOW()) OR
			(recurring = TRUE AND next_run_at <= NOW())
		)
		ORDER BY schedule_at ASC
	`

	rows, err := r.pool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*domain.ScheduledTransaction
	for rows.Next() {
		st := &domain.ScheduledTransaction{}
		err := rows.Scan(
			&st.ID, &st.UserID, &st.ToUserID, &st.Amount, &st.Type, &st.Status, &st.ScheduleAt,
			&st.Recurring, &st.Recurrence, &st.NextRunAt, &st.MaxRuns, &st.RunsCount, &st.Description,
			&st.CreatedAt, &st.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, st)
	}

	return transactions, nil
}

// Update updates a scheduled transaction
func (r *ScheduledTransactionPostgresRepository) Update(st *domain.ScheduledTransaction) error {
	query := `
		UPDATE scheduled_transactions SET
			user_id = $1, to_user_id = $2, amount = $3, type = $4, status = $5, schedule_at = $6,
			recurring = $7, recurrence = $8, next_run_at = $9, max_runs = $10, runs_count = $11, 
			description = $12, updated_at = NOW()
		WHERE id = $13
	`

	result, err := r.pool.Exec(context.Background(), query,
		st.UserID, st.ToUserID, st.Amount, st.Type, st.Status, st.ScheduleAt,
		st.Recurring, st.Recurrence, st.NextRunAt, st.MaxRuns, st.RunsCount, st.Description, st.ID,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("scheduled transaction not found")
	}

	return nil
}

// Delete deletes a scheduled transaction
func (r *ScheduledTransactionPostgresRepository) Delete(id int) error {
	query := `DELETE FROM scheduled_transactions WHERE id = $1`
	result, err := r.pool.Exec(context.Background(), query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("scheduled transaction not found")
	}
	return nil
}

// GetStats returns statistics about scheduled transactions
func (r *ScheduledTransactionPostgresRepository) GetScheduledTransactionStats(userID int) (*domain.ScheduledTransactionStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_scheduled,
			COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_count,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_count,
			COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_count,
			COUNT(CASE WHEN status = 'cancelled' THEN 1 END) as cancelled_count,
			COUNT(CASE WHEN recurring = TRUE THEN 1 END) as recurring_count,
			COUNT(CASE WHEN recurring = FALSE THEN 1 END) as one_time_count
		FROM scheduled_transactions 
		WHERE user_id = $1
	`

	stats := &domain.ScheduledTransactionStats{}
	err := r.pool.QueryRow(context.Background(), query, userID).Scan(
		&stats.TotalScheduled, &stats.PendingCount, &stats.CompletedCount,
		&stats.FailedCount, &stats.CancelledCount, &stats.RecurringCount, &stats.OneTimeCount,
	)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// ListByStatus fetches scheduled transactions by status
func (r *ScheduledTransactionPostgresRepository) ListByStatus(status string) ([]*domain.ScheduledTransaction, error) {
	query := `
		SELECT id, user_id, to_user_id, amount, type, status, schedule_at, 
		       recurring, recurrence, next_run_at, max_runs, runs_count, description, created_at, updated_at
		FROM scheduled_transactions 
		WHERE status = $1 
		ORDER BY schedule_at ASC
	`

	rows, err := r.pool.Query(context.Background(), query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*domain.ScheduledTransaction
	for rows.Next() {
		st := &domain.ScheduledTransaction{}
		err := rows.Scan(
			&st.ID, &st.UserID, &st.ToUserID, &st.Amount, &st.Type, &st.Status, &st.ScheduleAt,
			&st.Recurring, &st.Recurrence, &st.NextRunAt, &st.MaxRuns, &st.RunsCount, &st.Description,
			&st.CreatedAt, &st.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, st)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

// ListByTimeRange fetches scheduled transactions within a time range
func (r *ScheduledTransactionPostgresRepository) ListByTimeRange(from, to time.Time) ([]*domain.ScheduledTransaction, error) {
	query := `
		SELECT id, user_id, to_user_id, amount, type, status, schedule_at, 
		       recurring, recurrence, next_run_at, max_runs, runs_count, description, created_at, updated_at
		FROM scheduled_transactions 
		WHERE schedule_at >= $1 AND schedule_at <= $2
		ORDER BY schedule_at ASC
	`

	rows, err := r.pool.Query(context.Background(), query, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*domain.ScheduledTransaction
	for rows.Next() {
		st := &domain.ScheduledTransaction{}
		err := rows.Scan(
			&st.ID, &st.UserID, &st.ToUserID, &st.Amount, &st.Type, &st.Status, &st.ScheduleAt,
			&st.Recurring, &st.Recurrence, &st.NextRunAt, &st.MaxRuns, &st.RunsCount, &st.Description,
			&st.CreatedAt, &st.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, st)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}
