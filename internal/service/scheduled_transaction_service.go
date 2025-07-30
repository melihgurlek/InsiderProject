package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/melihgurlek/backend-path/internal/domain"
	"github.com/melihgurlek/backend-path/pkg/metrics"
)

// ScheduledTransactionServiceImpl implements domain.ScheduledTransactionService
type ScheduledTransactionServiceImpl struct {
	scheduledRepo      domain.ScheduledTransactionRepository
	transactionService domain.TransactionService
	mu                 sync.RWMutex
	executionTicker    *time.Ticker
	stopChan           chan struct{}
	isRunning          bool
}

// NewScheduledTransactionService creates a new ScheduledTransactionServiceImpl
func NewScheduledTransactionService(
	scheduledRepo domain.ScheduledTransactionRepository,
	transactionService domain.TransactionService,
) *ScheduledTransactionServiceImpl {
	return &ScheduledTransactionServiceImpl{
		scheduledRepo:      scheduledRepo,
		transactionService: transactionService,
		stopChan:           make(chan struct{}),
	}
}

// CreateScheduledTransaction creates a new scheduled transaction
func (s *ScheduledTransactionServiceImpl) CreateScheduledTransaction(st *domain.ScheduledTransaction) error {
	// Validate the scheduled transaction
	if err := st.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Set default status
	if st.Status == "" {
		st.Status = "pending"
	}

	// Set default runs count
	if st.RunsCount == 0 {
		st.RunsCount = 0
	}

	// Calculate next run for recurring transactions
	if st.Recurring {
		st.NextRunAt = &st.ScheduleAt
	}

	// Create the scheduled transaction
	if err := s.scheduledRepo.Create(st); err != nil {
		return fmt.Errorf("failed to create scheduled transaction: %w", err)
	}

	// Record metrics
	metrics.ScheduledTransactionCount.WithLabelValues(st.Type, st.Status).Inc()
	if st.Recurring {
		metrics.ScheduledTransactionCount.WithLabelValues("recurring", st.Recurrence).Inc()
	}

	log.Info().
		Int("id", st.ID).
		Int("user_id", st.UserID).
		Str("type", st.Type).
		Float64("amount", st.Amount).
		Time("schedule_at", st.ScheduleAt).
		Bool("recurring", st.Recurring).
		Msg("Scheduled transaction created")

	return nil
}

// GetScheduledTransaction retrieves a scheduled transaction by ID
func (s *ScheduledTransactionServiceImpl) GetScheduledTransaction(id int) (*domain.ScheduledTransaction, error) {
	st, err := s.scheduledRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get scheduled transaction: %w", err)
	}
	return st, nil
}

// ListUserScheduledTransactions retrieves all scheduled transactions for a user
func (s *ScheduledTransactionServiceImpl) ListUserScheduledTransactions(userID int) ([]*domain.ScheduledTransaction, error) {
	transactions, err := s.scheduledRepo.ListByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user scheduled transactions: %w", err)
	}
	return transactions, nil
}

// UpdateScheduledTransaction updates a scheduled transaction
func (s *ScheduledTransactionServiceImpl) UpdateScheduledTransaction(st *domain.ScheduledTransaction) error {
	// Validate the scheduled transaction
	if err := st.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Get existing transaction to check if it can be updated
	existing, err := s.scheduledRepo.GetByID(st.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing scheduled transaction: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("scheduled transaction not found")
	}

	// Don't allow updates to completed, failed, or cancelled transactions
	if existing.Status == "completed" || existing.Status == "failed" || existing.Status == "cancelled" {
		return fmt.Errorf("cannot update %s scheduled transaction", existing.Status)
	}

	// Update the scheduled transaction
	if err := s.scheduledRepo.Update(st); err != nil {
		return fmt.Errorf("failed to update scheduled transaction: %w", err)
	}

	log.Info().
		Int("id", st.ID).
		Str("status", st.Status).
		Msg("Scheduled transaction updated")

	return nil
}

// CancelScheduledTransaction cancels a scheduled transaction
func (s *ScheduledTransactionServiceImpl) CancelScheduledTransaction(id int) error {
	st, err := s.scheduledRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get scheduled transaction: %w", err)
	}
	if st == nil {
		return fmt.Errorf("scheduled transaction not found")
	}

	// Don't allow cancellation of completed, failed, or already cancelled transactions
	if st.Status == "completed" || st.Status == "failed" || st.Status == "cancelled" {
		return fmt.Errorf("cannot cancel %s scheduled transaction", st.Status)
	}

	st.MarkCancelled()

	if err := s.scheduledRepo.Update(st); err != nil {
		return fmt.Errorf("failed to cancel scheduled transaction: %w", err)
	}

	// Record metrics
	metrics.ScheduledTransactionCount.WithLabelValues(st.Type, "cancelled").Inc()

	log.Info().
		Int("id", st.ID).
		Msg("Scheduled transaction cancelled")

	return nil
}

// ExecuteScheduledTransactions executes all pending scheduled transactions
func (s *ScheduledTransactionServiceImpl) ExecuteScheduledTransactions() error {
	// Get pending transactions
	pending, err := s.scheduledRepo.ListPending()
	if err != nil {
		return fmt.Errorf("failed to get pending scheduled transactions: %w", err)
	}

	if len(pending) == 0 {
		return nil // No pending transactions
	}

	log.Info().Int("count", len(pending)).Msg("Executing scheduled transactions")

	// Execute each pending transaction
	for _, st := range pending {
		if err := s.ExecuteSingleScheduledTransaction(st); err != nil {
			log.Error().Err(err).Int("id", st.ID).Msg("Failed to execute scheduled transaction")
			// Continue with other transactions
		}
	}

	return nil
}

// ExecuteSingleScheduledTransaction executes a single scheduled transaction
func (s *ScheduledTransactionServiceImpl) ExecuteSingleScheduledTransaction(st *domain.ScheduledTransaction) error {
	// Create span for tracing
	ctx, span := otel.Tracer("scheduled-transaction-service").Start(context.Background(), "execute-scheduled-transaction")
	defer span.End()

	span.SetAttributes(
		attribute.Int("scheduled_transaction.id", st.ID),
		attribute.String("scheduled_transaction.type", st.Type),
		attribute.Int("scheduled_transaction.user_id", st.UserID),
		attribute.Float64("scheduled_transaction.amount", st.Amount),
		attribute.Bool("scheduled_transaction.recurring", st.Recurring),
	)

	startTime := time.Now()

	// Execute the transaction based on type
	var err error
	switch st.Type {
	case "credit":
		err = s.transactionService.Credit(st.UserID, st.Amount)
	case "debit":
		err = s.transactionService.Debit(st.UserID, st.Amount)
	case "transfer":
		if st.ToUserID == nil {
			err = fmt.Errorf("transfer requires to_user_id")
		} else {
			err = s.transactionService.Transfer(st.UserID, *st.ToUserID, st.Amount)
		}
	default:
		err = fmt.Errorf("unknown transaction type: %s", st.Type)
	}

	// Check if context was cancelled
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Update the scheduled transaction status
	if err != nil {
		st.MarkFailed()
		span.RecordError(err)
		metrics.ScheduledTransactionExecutionFailure.WithLabelValues(st.Type).Inc()
	} else {
		st.MarkCompleted()
		metrics.ScheduledTransactionExecutionSuccess.WithLabelValues(st.Type).Inc()
	}

	// Update the scheduled transaction in the database
	if updateErr := s.scheduledRepo.Update(st); updateErr != nil {
		log.Error().Err(updateErr).Int("id", st.ID).Msg("Failed to update scheduled transaction status")
	}

	// Record execution time
	executionTime := time.Since(startTime)
	metrics.ScheduledTransactionExecutionDuration.WithLabelValues(st.Type).Observe(executionTime.Seconds())

	span.SetAttributes(attribute.Float64("execution_time_seconds", executionTime.Seconds()))

	log.Info().
		Int("id", st.ID).
		Str("type", st.Type).
		Bool("success", err == nil).
		Dur("execution_time", executionTime).
		Msg("Scheduled transaction executed")

	return err
}

// GetScheduledTransactionStats returns statistics about scheduled transactions
func (s *ScheduledTransactionServiceImpl) GetScheduledTransactionStats() (*domain.ScheduledTransactionStats, error) {
	stats := &domain.ScheduledTransactionStats{}

	// Get counts by status
	statuses := []string{"pending", "completed", "failed", "cancelled"}
	for _, status := range statuses {
		transactions, err := s.scheduledRepo.ListByStatus(status)
		if err != nil {
			return nil, fmt.Errorf("failed to get %s scheduled transactions: %w", status, err)
		}

		count := int64(len(transactions))
		switch status {
		case "pending":
			stats.PendingCount = count
		case "completed":
			stats.CompletedCount = count
		case "failed":
			stats.FailedCount = count
		case "cancelled":
			stats.CancelledCount = count
		}
		stats.TotalScheduled += count
	}

	// Get recurring vs one-time counts
	allTransactions, err := s.scheduledRepo.ListByStatus("pending")
	if err != nil {
		return nil, fmt.Errorf("failed to get pending scheduled transactions: %w", err)
	}

	for _, st := range allTransactions {
		if st.Recurring {
			stats.RecurringCount++
		} else {
			stats.OneTimeCount++
		}
	}

	// Find next execution time
	if len(allTransactions) > 0 {
		var nextExecution *time.Time
		for _, st := range allTransactions {
			if st.Recurring && st.NextRunAt != nil {
				if nextExecution == nil || st.NextRunAt.Before(*nextExecution) {
					nextExecution = st.NextRunAt
				}
			} else if !st.Recurring {
				if nextExecution == nil || st.ScheduleAt.Before(*nextExecution) {
					nextExecution = &st.ScheduleAt
				}
			}
		}

		if nextExecution != nil {
			nextExecStr := nextExecution.Format(time.RFC3339)
			stats.NextExecutionTime = &nextExecStr
		}
	}

	return stats, nil
}

// Start begins the background execution of scheduled transactions
func (s *ScheduledTransactionServiceImpl) Start(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		return
	}

	s.isRunning = true
	s.executionTicker = time.NewTicker(1 * time.Minute) // Check every minute

	log.Info().Msg("Starting scheduled transaction executor")

	go s.executionLoop(ctx)
}

// Stop stops the background execution of scheduled transactions
func (s *ScheduledTransactionServiceImpl) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return
	}

	s.isRunning = false
	if s.executionTicker != nil {
		s.executionTicker.Stop()
	}
	close(s.stopChan)

	log.Info().Msg("Stopped scheduled transaction executor")
}

// executionLoop runs in the background to execute scheduled transactions
func (s *ScheduledTransactionServiceImpl) executionLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-s.executionTicker.C:
			if err := s.ExecuteScheduledTransactions(); err != nil {
				log.Error().Err(err).Msg("Failed to execute scheduled transactions")
			}
		}
	}
}
