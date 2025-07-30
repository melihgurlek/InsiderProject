package domain

import (
	"time"
)

// ValidationError is a custom error type for validation failures
type ValidationError struct {
	Msg string
}

func (e *ValidationError) Error() string {
	return e.Msg
}

// ScheduledTransaction represents a transaction that will be executed at a future time
type ScheduledTransaction struct {
	ID          int        `json:"id"`
	UserID      int        `json:"user_id"`
	ToUserID    *int       `json:"to_user_id,omitempty"` // for transfers
	Amount      float64    `json:"amount"`
	Type        string     `json:"type"`   // "credit", "debit", "transfer"
	Status      string     `json:"status"` // "pending", "completed", "failed", "cancelled"
	ScheduleAt  time.Time  `json:"schedule_at"`
	Recurring   bool       `json:"recurring"`
	Recurrence  string     `json:"recurrence,omitempty"` // "daily", "weekly", "monthly", "yearly"
	NextRunAt   *time.Time `json:"next_run_at,omitempty"`
	MaxRuns     *int       `json:"max_runs,omitempty"`
	RunsCount   int        `json:"runs_count"`
	Description string     `json:"description,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Validate validates the scheduled transaction's business logic
func (st *ScheduledTransaction) Validate() error {
	if st.UserID <= 0 {
		return &ValidationError{Msg: "user_id must be positive"}
	}
	if st.Amount <= 0 {
		return &ValidationError{Msg: "amount must be positive"}
	}
	// ... (other checks for type, transfer, recurrence)

	// Use UTC for all time logic and include a grace period
	if st.ScheduleAt.Before(time.Now().UTC().Add(-10 * time.Second)) {
		return &ValidationError{Msg: "schedule_at must be in the future"}
	}

	return nil
}

// ShouldExecute checks if the scheduled transaction should be executed now
func (st *ScheduledTransaction) ShouldExecute() bool {
	if st.Status != "pending" {
		return false
	}

	if st.Recurring {
		return st.NextRunAt != nil && time.Now().After(*st.NextRunAt)
	}

	return time.Now().After(st.ScheduleAt)
}

// CalculateNextRun calculates the next execution time for recurring transactions
func (st *ScheduledTransaction) CalculateNextRun() *time.Time {
	if !st.Recurring {
		return nil
	}

	var nextRun time.Time
	if st.NextRunAt != nil {
		nextRun = *st.NextRunAt
	} else {
		nextRun = st.ScheduleAt
	}

	switch st.Recurrence {
	case "daily":
		nextRun = nextRun.AddDate(0, 0, 1)
	case "weekly":
		nextRun = nextRun.AddDate(0, 0, 7)
	case "monthly":
		nextRun = nextRun.AddDate(0, 1, 0)
	case "yearly":
		nextRun = nextRun.AddDate(1, 0, 0)
	}

	return &nextRun
}

// ShouldStop checks if the recurring transaction should stop
func (st *ScheduledTransaction) ShouldStop() bool {
	if !st.Recurring {
		return true
	}

	if st.MaxRuns != nil && st.RunsCount >= *st.MaxRuns {
		return true
	}

	return false
}

// MarkCompleted marks the transaction as completed and updates next run
func (st *ScheduledTransaction) MarkCompleted() {
	st.RunsCount++
	st.UpdatedAt = time.Now()

	if st.ShouldStop() {
		st.Status = "completed"
	} else {
		st.NextRunAt = st.CalculateNextRun()
	}
}

// MarkFailed marks the transaction as failed
func (st *ScheduledTransaction) MarkFailed() {
	st.Status = "failed"
	st.UpdatedAt = time.Now()
}

// MarkCancelled marks the transaction as cancelled
func (st *ScheduledTransaction) MarkCancelled() {
	st.Status = "cancelled"
	st.UpdatedAt = time.Now()
}
