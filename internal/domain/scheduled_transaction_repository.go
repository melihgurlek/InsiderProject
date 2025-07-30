package domain

import "time"

// ScheduledTransactionRepository defines the interface for scheduled transaction data access
type ScheduledTransactionRepository interface {
	// Create creates a new scheduled transaction
	Create(st *ScheduledTransaction) error

	// GetByID retrieves a scheduled transaction by ID
	GetByID(id int) (*ScheduledTransaction, error)

	// GetScheduledTransactionStats returns statistics about scheduled transactions
	GetScheduledTransactionStats(userID int) (*ScheduledTransactionStats, error)

	// ListByUser retrieves all scheduled transactions for a user
	ListByUser(userID int) ([]*ScheduledTransaction, error)

	// ListPending retrieves all pending scheduled transactions that should be executed
	ListPending() ([]*ScheduledTransaction, error)

	// Update updates a scheduled transaction
	Update(st *ScheduledTransaction) error

	// Delete deletes a scheduled transaction
	Delete(id int) error

	// ListByStatus retrieves scheduled transactions by status
	ListByStatus(status string) ([]*ScheduledTransaction, error)

	// ListByTimeRange retrieves scheduled transactions within a time range
	ListByTimeRange(from, to time.Time) ([]*ScheduledTransaction, error)
}
