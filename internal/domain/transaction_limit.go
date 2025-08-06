package domain

import (
	"context"
	"time"
)

// TransactionLimitRule defines a rule for limiting transactions.
type TransactionLimitRule struct {
	ID          string        // Unique rule ID
	UserID      int           // User or Account the rule applies to
	RuleType    RuleType      // e.g., MaxPerTransaction, DailyTotal, TxCount, MinInterval
	LimitAmount float64       // Amount or count, depending on rule type
	Currency    string        // Optional: for multicurrency support
	Window      time.Duration // e.g., 24h for daily, 1h for hourly, 0 for per-tx
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Active      bool
}

// RuleType enumerates supported rule types.
type RuleType string

const (
	RuleMaxPerTransaction RuleType = "max_per_transaction"
	RuleDailyTotal        RuleType = "daily_total"
	RuleTxCount           RuleType = "tx_count"
	RuleMinInterval       RuleType = "min_interval"
)

// TransactionLimitRepository abstracts rule and history storage.
type TransactionLimitRepository interface {
	GetRulesForUser(ctx context.Context, userID int) ([]TransactionLimitRule, error)
	AddRule(ctx context.Context, rule TransactionLimitRule) (TransactionLimitRule, error)
	RemoveRule(ctx context.Context, userID int, ruleID string) error
	RecordTransaction(ctx context.Context, userID int, amount float64, currency string, timestamp time.Time) error
	GetTransactionSum(ctx context.Context, userID int, window time.Duration, currency string) (float64, error)
	GetTransactionCount(ctx context.Context, userID int, window time.Duration) (int, error)
	GetLastTransactionTime(ctx context.Context, userID int) (time.Time, error)
	CheckAndRecordTransaction(ctx context.Context, userID int, amount float64, currency string, timestamp time.Time) error
}

// TransactionLimitService defines business logic for rule evaluation.
type TransactionLimitService interface {
	CheckAndRecordTransaction(ctx context.Context, userID int, amount float64, currency string, timestamp time.Time) error
	AddRule(ctx context.Context, rule TransactionLimitRule) (TransactionLimitRule, error)
	RemoveRule(ctx context.Context, userID int, ruleID string) error
	ListRules(ctx context.Context, userID int) ([]TransactionLimitRule, error)
}
