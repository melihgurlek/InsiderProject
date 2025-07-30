package domain

// ScheduledTransactionService defines the interface for scheduled transaction business logic
type ScheduledTransactionService interface {
	// CreateScheduledTransaction creates a new scheduled transaction
	CreateScheduledTransaction(st *ScheduledTransaction) error

	// GetScheduledTransaction retrieves a scheduled transaction by ID
	GetScheduledTransaction(id int) (*ScheduledTransaction, error)

	// ListUserScheduledTransactions retrieves all scheduled transactions for a user
	ListUserScheduledTransactions(userID int) ([]*ScheduledTransaction, error)

	// UpdateScheduledTransaction updates a scheduled transaction
	UpdateScheduledTransaction(st *ScheduledTransaction) error

	// CancelScheduledTransaction cancels a scheduled transaction
	CancelScheduledTransaction(id int) error

	// ExecuteScheduledTransactions executes all pending scheduled transactions
	ExecuteScheduledTransactions() error

	// GetScheduledTransactionStats returns statistics about scheduled transactions
	GetScheduledTransactionStats() (*ScheduledTransactionStats, error)
}

// ScheduledTransactionStats holds statistics about scheduled transactions
type ScheduledTransactionStats struct {
	TotalScheduled    int64
	PendingCount      int64
	CompletedCount    int64
	FailedCount       int64
	CancelledCount    int64
	RecurringCount    int64
	OneTimeCount      int64
	NextExecutionTime *string // ISO format string
}
