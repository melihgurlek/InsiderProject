package domain

import "context"

// TransactionTask represents a task to be processed by the worker pool
type TransactionTask struct {
	ID       string
	Type     string // "credit", "debit", "transfer"
	UserID   int
	ToUserID *int // for transfers
	Amount   float64
	Priority int // higher number = higher priority
}

// TransactionResult represents the result of processing a transaction task
type TransactionResult struct {
	TaskID    string
	Success   bool
	Error     error
	Message   string
	Timestamp int64
}

// TransactionProcessor defines the interface for concurrent transaction processing
type TransactionProcessor interface {
	// SubmitTask submits a transaction task to the processing queue
	SubmitTask(ctx context.Context, task *TransactionTask) error

	// Start starts the worker pool
	Start(ctx context.Context) error

	// Stop gracefully stops the worker pool
	Stop(ctx context.Context) error

	// GetStats returns current processing statistics
	GetStats() *ProcessingStats
}

// ProcessingStats holds statistics about transaction processing
type ProcessingStats struct {
	TotalProcessed     int64
	SuccessfulTasks    int64
	FailedTasks        int64
	QueueSize          int
	ActiveWorkers      int
	AverageProcessTime float64
}
