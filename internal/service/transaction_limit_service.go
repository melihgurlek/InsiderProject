package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/melihgurlek/backend-path/internal/domain"
)

type transactionLimitService struct {
	repo domain.TransactionLimitRepository
}

func NewTransactionLimitService(repo domain.TransactionLimitRepository) domain.TransactionLimitService {
	return &transactionLimitService{repo: repo}
}

// Atomically checks all rules and records the transaction if allowed.
func (s *transactionLimitService) CheckAndRecordTransaction(ctx context.Context, userID int, amount float64, currency string, timestamp time.Time) error {
	return s.repo.CheckAndRecordTransaction(ctx, userID, amount, currency, timestamp)
}

func (s *transactionLimitService) AddRule(ctx context.Context, rule domain.TransactionLimitRule) (domain.TransactionLimitRule, error) {
	// Validate RuleType
	switch rule.RuleType {
	case domain.RuleMaxPerTransaction, domain.RuleDailyTotal, domain.RuleTxCount, domain.RuleMinInterval:
		// valid
	default:
		return domain.TransactionLimitRule{}, errors.New("invalid rule type")
	}
	// Validate LimitAmount
	if rule.LimitAmount <= 0 {
		return domain.TransactionLimitRule{}, errors.New("limit amount must be positive")
	}
	// Validate Window for rules that require it
	if (rule.RuleType == domain.RuleDailyTotal || rule.RuleType == domain.RuleTxCount || rule.RuleType == domain.RuleMinInterval) && rule.Window <= 0 {
		return domain.TransactionLimitRule{}, errors.New("window must be positive for this rule type")
	}
	// Generate UUID if not set
	if rule.ID == "" {
		rule.ID = uuid.NewString()
	}
	// Set CreatedAt/UpdatedAt if not set
	if rule.CreatedAt.IsZero() {
		rule.CreatedAt = time.Now().UTC()
	}
	if rule.UpdatedAt.IsZero() {
		rule.UpdatedAt = rule.CreatedAt
	}
	rule, err := s.repo.AddRule(ctx, rule)
	if err != nil {
		return domain.TransactionLimitRule{}, err
	}
	return rule, nil
}

func (s *transactionLimitService) RemoveRule(ctx context.Context, userID int, ruleID string) error {
	return s.repo.RemoveRule(ctx, userID, ruleID)
}

func (s *transactionLimitService) ListRules(ctx context.Context, userID int) ([]domain.TransactionLimitRule, error) {
	return s.repo.GetRulesForUser(ctx, userID)
}
